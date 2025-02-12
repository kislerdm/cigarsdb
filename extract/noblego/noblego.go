// Package noblego defines the logic and the standard implementation to extract data from noblego.de
package noblego

import (
	"bytes"
	"cigarsdb/htmlfilter"
	"cigarsdb/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Client defines the client to noblego.de to fetch data from.
type Client struct {
	HTTPClient HTTPClient
}

func (c Client) Read(_ context.Context, id string) (r storage.Record, err error) {
	var resp *http.Response
	if resp, err = c.HTTPClient.Get(id); err == nil {
		err = readDetailsPage(resp.Body, &r)
		_ = resp.Body.Close()
		if err == nil {
			r.URL = id
		}
	}
	return r, err
}

// readDetailsPage extracts the cigar's attributes from the html page, i.e., an adaptor between html and storage.Record.
// Note that the pages which contain the word "sampler", or "Sampler" are discarded.
func readDetailsPage(v io.ReadCloser, o *storage.Record) error {
	var err error
	var doc *html.Node
	if doc, err = html.Parse(v); err == nil {
		n := htmlfilter.Node{Node: doc}
		o.Name = readName(n)
		err = readAttributes(n, o)
		err = errors.Join(err, readPrice(n, o))
		readFreeDetails(n, o)
	}
	return err
}

func readFreeDetails(n htmlfilter.Node, o *storage.Record) {
	var tmp = make(map[string]string)
	for nn := range n.Find("div.collateral-container") {
		for nnn := range nn.Find("div.std") {
			if found := readVideoURL(nnn, o); found {
				for x := range nnn.Find("div.artikel-textblock") {
					nnn = x
					break
				}
			}
			for detail := range nnn.Find("h3") {
				if detail.LastChild != nil {
					k := detail.LastChild.Data
					var val bytes.Buffer
					s := detail.Node
					for s.NextSibling != nil {
						s = s.NextSibling
						if s.DataAtom == atom.H3 {
							break
						}
						if s.DataAtom == atom.P {
							if val.Len() > 0 {
								_, _ = val.WriteString("\n")
							}
							_, _ = val.WriteString(strings.TrimSpace(htmlfilter.InnerHTML(s)))
						}
					}
					if val.Len() > 0 {
						tmp[k] = val.String()
					}
				}
			}
			break
		}
	}
	if len(tmp) > 0 {
		o.Details = maps.Clone(tmp)
	}
}

func readVideoURL(n htmlfilter.Node, o *storage.Record) bool {
	var found bool
	for video := range n.Find("div.artikel-youtube-block") {
		var videoNode htmlfilter.Node
		for videoNode = range video.Find("object") {
		}
		if videoNode.Node == nil {
			for videoNode = range n.Find("iframe") {
				if videoNode.Node != nil {
					break
				}
			}
		}
		if videoNode.Node != nil {
			var videoURL string
			var keyRef string
			switch videoNode.DataAtom {
			case atom.Object:
				keyRef = "data"
			case atom.Iframe:
				keyRef = "src"
			}
			for _, att := range videoNode.Attr {
				if att.Key == keyRef {
					videoURL = newVideoURL(att.Val)
				}
			}
			if videoURL != "" {
				o.VideoURLs = append(o.VideoURLs, videoURL)
				found = true
			}
		}
		break
	}
	return found
}

func newVideoURL(s string) string {
	var trim = func(s string, tag string) string {
		if els := strings.SplitN(s, tag, 2); len(els) == 2 {
			s = els[1]
			s = strings.SplitN(s, "?", 2)[0]
		}
		return s
	}

	if strings.Contains(s, "youtube") {
		switch {
		case strings.Contains(s, ".com/embed/"):
			s = trim(s, ".com/embed/")
		case strings.Contains(s, ".com/v/watch?v="):
			s = trim(s, ".com/v/watch?v=")
		case strings.Contains(s, ".com/v/"):
			s = trim(s, ".com/v/")
		}
		// passthrough if the video ID was not extracted
		if !strings.Contains(s, "youtube") {
			s = fmt.Sprintf("https://www.youtube.com/watch?v=%s", s)
		}
	}

	return s
}

func readName(n htmlfilter.Node) string {
	var o string
	for n = range n.Find("div.product-name") {
		for name := range n.Find("h1") {
			if name.LastChild != nil {
				o = name.LastChild.Data
			}
		}
	}
	return o
}

func readPrice(n htmlfilter.Node, o *storage.Record) error {
	var err error

	var lastPriceOption htmlfilter.Node
	var extracted bool
	for nn := range n.Find("ul.product-prices") {
		if extracted {
			break
		}
		for lastPriceOption = range nn.Find("li") {
		}
		extracted = true
	}

	var cost float64
	for nnn := range lastPriceOption.Find("span.price") {
		tmp := nnn.LastChild.Data
		// remove length of euro sign with the unbreakable space
		tmp = tmp[:len(tmp)-5]
		tmp = strings.Replace(tmp, ",", ".", -1)
		cost, _ = strconv.ParseFloat(tmp, 64)
	}

	var quantity = 1
	for packaging := range lastPriceOption.Find("span.Verpackungseinheit") {
		if strings.HasSuffix(packaging.LastChild.Data, "er") {
			quantity, err = strconv.Atoi(strings.TrimSuffix(packaging.LastChild.Data, "er"))
		}
	}
	if err == nil {
		o.Price = cost / float64(quantity)
	}

	return err
}

func readAttributes(n htmlfilter.Node, o *storage.Record) error {
	var err error
	for nn := range n.Find("li.product-attribute-*") {
		for _, attr := range nn.Attr {
			if attr.Key == "class" {
				var er error

				val := nn.Node
				switch attr.Val {
				// Identification
				case "product-attribute-brand":
					if v := dataFromATagText(val); v != nil {
						o.Brand = *v
					}
				case "product-attribute-series":
					if v := dataFromATagText(val); v != nil {
						o.Series = *v
					}

				// Shape
				case "product-attribute-cig_diameter":
					if v := dataFromFirstSpanChild(val); v != nil {
						o.Diameter, er = strconv.ParseFloat(*v, 64)
					}
				case "product-attribute-cig_gauge":
					if v := dataFromFirstSpanChild(val); v != nil {
						o.Ring, er = strconv.Atoi(*v)
					}
				case "product-attribute-cig_length":
					if v := dataFromFirstSpanChild(val); v != nil {
						o.Length, er = strconv.ParseFloat(*v, 64)
					}
				case "product-attribute-cig_size":
					if v := dataFromATagText(val); v != nil {
						o.Format = *v
					}

				// Manufacturing
				case "product-attribute-cig_maker":
					o.Maker = dataFromFirstSpanChild(val)
				case "product-attribute-herkunft":
					if v := dataFromATagText(val); v != nil {
						o.ManufactureOrigin = *v
					}
				case "product-attribute-cig_construction":
					o.Construction = dataFromATagText(val)
				case "product-attribute-cig_form":
					o.IsBoxpressed = pointer(false)
					if v := dataFromATagText(val); v != nil && strings.ToLower(*v) == "boxpressed" {
						o.IsBoxpressed = pointer(true)
					}

				// Blend
				case "product-attribute-cig_wrapper_origin":
					if v := dataFromFirstSpanChild(val); v != nil {
						o.WrapperOrigin = parseConcatSlice(*v)
					}
				case "product-attribute-cig_filler":
					if v := dataFromFirstSpanChild(val); v != nil {
						o.FillerOrigin = parseConcatSlice(*v)
					}
				case "product-attribute-cig_binder":
					if v := dataFromFirstSpanChild(val); v != nil {
						o.BinderOrigin = parseConcatSlice(*v)
					}
				case "product-attribute-cig_wrapper_tobacco":
					o.WrapperType = dataFromFirstSpanChild(val)
				case "product-attribute-cig_aroma":
					if v := dataFromFirstSpanChild(val); v != nil {
						o.AromaProfileManufacturer = parseConcatSlice(*v)
					}
					//	TODO: add IsFlavoured
				case "product-attribute-strength":
					o.Strength = dataFromATagText(val)
				case "product-attribute-flavour_strength":
					o.FlavourStrength = dataFromFirstSpanChild(val)
				case "product-attribute-cig_duration":
					o.SmokingDuration = dataFromFirstSpanChild(val)

				case "product-attribute-cig_special":
					v := dataFromATagText(val)
					if v == nil {
						v = dataFromFirstSpanChild(val)
					}
					o.AdditionalNotes = v
				}

				if er != nil {
					err = errors.Join(err, fmt.Errorf("attributes extraction error: %w", er))
				}
			}
		}
	}
	return err
}

func spanReader(n *html.Node, fn func(n *html.Node) *string) (o *string) {
	for c := range n.ChildNodes() {
		for _, att := range c.Attr {
			if att.Key == "class" && att.Val == "data" {
				o = fn(c)
				break
			}
		}
	}
	return o
}

func dataFromFirstSpanChild(n *html.Node) *string {
	return spanReader(n, func(n *html.Node) *string {
		s := strings.TrimSpace(n.FirstChild.Data)
		return pointer(s)
	})
}

func dataFromATagText(n *html.Node) *string {
	return spanReader(n, func(n *html.Node) *string {
		for nnn := range n.Descendants() {
			if nnn.DataAtom == atom.A {
				return pointer(strings.TrimSpace(nnn.LastChild.Data))
			}
		}
		return nil
	})
}

func parseConcatSlice(s string) []string {
	o := strings.Split(s, ",")
	for i, v := range o {
		o[i] = strings.TrimSpace(v)
	}
	return o
}

func (c Client) ReadBulk(ctx context.Context, limit, page uint) (r []storage.Record, nextPage uint, err error) {
	// add gauge filter to omit lighters etc.
	const baseURL = "https://www.noblego.de/zigarren/?cig_gauge%5B0%5D=4481&cig_gauge%5B1%5D=3135&cig_gauge%5B2%5D=208&cig_gauge%5B3%5D=207&cig_gauge%5B4%5D=206&cig_gauge%5B5%5D=205&cig_gauge%5B6%5D=204&cig_gauge%5B7%5D=203&cig_gauge%5B8%5D=202&cig_gauge%5B9%5D=201&cig_gauge%5B10%5D=200&cig_gauge%5B11%5D=199&cig_gauge%5B12%5D=198&cig_gauge%5B13%5D=197&cig_gauge%5B14%5D=196&cig_gauge%5B15%5D=233&cig_gauge%5B16%5D=232&cig_gauge%5B17%5D=231&cig_gauge%5B18%5D=230&cig_gauge%5B19%5D=1488&cig_gauge%5B20%5D=229&cig_gauge%5B21%5D=228&cig_gauge%5B22%5D=227&cig_gauge%5B23%5D=2260&cig_gauge%5B24%5D=226&cig_gauge%5B25%5D=225&cig_gauge%5B26%5D=224&cig_gauge%5B27%5D=223&cig_gauge%5B28%5D=222&cig_gauge%5B29%5D=221&cig_gauge%5B30%5D=220&cig_gauge%5B31%5D=219&cig_gauge%5B32%5D=218&cig_gauge%5B33%5D=217&cig_gauge%5B34%5D=216&cig_gauge%5B35%5D=215&cig_gauge%5B36%5D=214&cig_gauge%5B37%5D=213&cig_gauge%5B38%5D=212&cig_gauge%5B39%5D=1025&cig_gauge%5B40%5D=211&cig_gauge%5B41%5D=210&cig_gauge%5B42%5D=2763&cig_gauge%5B43%5D=2604&cig_gauge%5B44%5D=1753&cig_gauge%5B45%5D=1730&cig_gauge%5B46%5D=4517&cig_gauge%5B47%5D=1022&cig_gauge%5B48%5D=209&cig_gauge%5B49%5D=1731&cig_gauge%5B50%5D=1661&cig_gauge%5B51%5D=2344&cig_gauge%5B52%5D=4624"
	const itemsPerPage = 96
	if limit == 0 || limit > itemsPerPage {
		limit = itemsPerPage
	}
	if page == 0 {
		page++
	}
	u := baseURL + fmt.Sprintf("&limit=%d&p=%d", limit, page)
	var resp *http.Response
	if resp, err = c.HTTPClient.Get(u); err == nil {
		var totalItems uint
		var urlItems []string
		totalItems, urlItems, err = readItemsFromListPage(resp.Body)
		_ = resp.Body.Close()

		if err == nil && dataInCurrentPage(page, limit, totalItems) {
			maxN := len(urlItems)
			r = make([]storage.Record, maxN)
			var flags = make(chan struct{}, maxN)
			for i, u := range urlItems {
				go func() {
					var er error
					if r[i], er = c.Read(ctx, u); er != nil {
						err = errors.Join(err, fmt.Errorf("error reading details using %s: %w", u, er))
					}
					flags <- struct{}{}
				}()
			}
			for maxN > 0 {
				<-flags
				maxN--
			}
			if dataInCurrentPage(page+1, limit, totalItems) {
				nextPage = page + 1
			}
		}
	}
	return r, nextPage, err
}

func dataInCurrentPage(page, limit, total uint) bool {
	var o bool
	switch page {
	case 0, 1:
		o = total > 0
	default:
		o = total > page*limit-limit
	}
	return o
}

func readItemsFromListPage(v io.ReadCloser) (totalItems uint, urls []string, err error) {
	var n *html.Node
	if n, err = html.Parse(v); err == nil {
		nn := htmlfilter.Node{Node: n}
		totalItems, err = readTotalNumberOfItems(nn)
		if err == nil {
			urls = readURLs(nn)
		}
	}
	return totalItems, urls, err
}

func readURLs(n htmlfilter.Node) []string {
	var o []string
	for nn := range n.Find("li.item") {
		for nn = range nn.Find("h2.product-name") {
			for a := range nn.Find("a") {
				if !skipItem(a) {
					for _, att := range a.Attr {
						if att.Key == "href" {
							o = append(o, att.Val)
						}
					}
				}
			}
		}
	}
	return o
}

func skipItem(n htmlfilter.Node) bool {
	return n.LastChild != nil && strings.Contains(strings.ToLower(n.LastChild.Data), "sampler")
}

func readTotalNumberOfItems(n htmlfilter.Node) (uint, error) {
	var (
		o   uint
		err error
	)
	for n = range n.Find("p.amount") {
		s := strings.TrimSpace(strings.Split(n.LastChild.Data, " ")[0])
		var tmp uint64
		if tmp, err = strconv.ParseUint(s, 10, 64); err == nil {
			o = uint(tmp)
		}
		break
	}
	return o, err
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

func pointer[V string | bool | float64 | int](s V) *V {
	return &s
}
