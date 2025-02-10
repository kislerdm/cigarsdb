// Package noblego defines the logic and the standard implementation to extract data from noblego.de
package noblego

import (
	"cigarsdb/htmlfilter"
	"cigarsdb/storage"
	"context"
	"errors"
	"fmt"
	"io"
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
	}
	return err
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

func (c Client) ReadBulk(_ context.Context, limit, page uint) (r []storage.Record, nextPage uint, err error) {
	const baseURL = "https://www.noblego.de/zigarren/?limit=%d&p=%d"
	const itemsPerPage = 96
	if limit == 0 || limit > itemsPerPage {
		limit = itemsPerPage
	}
	if page == 0 {
		page++
	}
	u := fmt.Sprintf(baseURL, limit, page)
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
				i := i
				u := u
				go func() {
					var er error
					if r[i], er = c.Read(nil, u); er != nil {
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
