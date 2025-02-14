package cigarworld

import (
	"bytes"
	"cigarsdb/extract"
	"cigarsdb/htmlfilter"
	"cigarsdb/storage"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Client defines the client to cigarworld.de to fetch data from.
type Client struct {
	HTTPClient extract.HTTPClient
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

func readDetailsPage(v io.Reader, o *storage.Record) error {
	var err error
	var doc *html.Node
	if doc, err = html.Parse(v); err == nil {
		n := htmlfilter.Node{Node: doc}
		o.Name = readName(n)
		err = readAttributes(n, o)
		err = errors.Join(err, readPrice(n, o))
		readDescription(n, o)
		readAromaProfileCommunity(n, o)
	}
	return err
}

func readAromaProfileCommunity(n htmlfilter.Node, o *storage.Record) {
	var aromaWeights []float64
	var dataRub bool
	var cntOfVotes int
	for nn := range n.Find("div#tab-pane-tasting") {
		for nnn := range nn.Find("p") {
			if nnn.LastChild != nil {
				s := strings.TrimSpace(nnn.LastChild.Data)
				for i, el := range s {
					if el == '(' {
						var err error
						if cntOfVotes, err = strconv.Atoi(s[i+1 : len(s)-1]); err != nil || cntOfVotes == 0 {
							return
						}
						break
					}
				}
			}
			break
		}

		for nnn := range nn.Find("canvas.aromaimg") {
			for _, att := range nnn.Attr {
				switch att.Key {
				case "data-content":
					aromaWeights = readAromaWeights(att.Val)
				case "data-rub":
					dataRub = strings.TrimSpace(att.Val) == "t"
				}
			}
			break
		}
		break
	}

	if len(aromaWeights) > 0 {
		for nn := range n.Find("body") {
			var found bool
			for nnn := range nn.ChildNodes() {
				if !found && nnn.DataAtom == atom.Script {
					s := strings.TrimSpace(htmlfilter.InnerHTML(nnn))
					if strings.Contains(s, "NameArrObj") {
						aromaCat, aromaCatAlt := readAromaCats(s)
						if dataRub {
							aromaCat = aromaCatAlt
						}
						var w = make(map[string]float64, len(aromaWeights))
						for i, el := range aromaWeights {
							w[aromaCat[i]] = el
						}
						o.AromaProfileCommunity = &storage.AromaProfileCommunity{
							Weights:       w,
							NumberOfVotes: cntOfVotes,
						}
						found = true
					}
				}
			}
		}
	}
}

func readAromaWeights(s string) []float64 {
	var o = make([]float64, len(s))
	var sum int
	for i, v := range s {
		var val int
		switch v {
		case '0':
			val = 0
		case '1':
			val = 1
		case '2':
			val = 2
		case '3':
			val = 3
		case '4':
			val = 4
		case '5':
			val = 5
		case '6':
			val = 6
		case '7':
			val = 7
		case '8':
			val = 8
		case '9':
			val = 9
		}
		o[i] = float64(val)
		sum += val
	}
	for i, v := range o {
		o[i] = v / float64(sum)
	}
	return o
}

func readAromaCats(s string) (aromaCat []string, aromaCatAlt []string) {
	s = strings.TrimSpace(s)
	return readAromaCat(s, "AromaNamenArr"), readAromaCat(s, "AromaTabacNamenArr")
}

func readAromaCat(s string, filter string) []string {
	var o []string
	var atEnd bool
	for i := 0; i < len(s); i++ {
		switch {
		case len(s) >= i+len(filter) && s[i:i+len(filter)] == filter:
			i += len(filter)
			var l, r int
			for i < len(s) {
				switch {
				case s[i] == '"' && (s[i-1] == '[' || s[i-1] == ','):
					l = i + 1
				case s[i] == '"' && (s[i+1] == ']' || s[i+1] == ','):
					r = i
				case s[i] == ']':
					atEnd = true
				}
				if r > l {
					o = append(o, s[l:r])
					l = 0
					r = 0
				}
				if atEnd {
					break
				}
				i++
			}
		}
		if atEnd {
			break
		}
	}
	return o
}

func readDescription(n htmlfilter.Node, o *storage.Record) {
	for nn := range n.Find("div.contentpage__content") {
		if tmp := htmlfilter.InnerHTML(nn.Node); tmp != "" {
			o.Details = map[string]string{"description": tmp}
		}
		break
	}
}

func readPrice(n htmlfilter.Node, o *storage.Record) error {
	var err error
	var costs float64
	for nn := range n.Find("span.preis") {
		for nn = range nn.Find("span") {
			for _, att := range nn.Attr {
				if att.Key == "data-eurval" {
					costs, err = strconv.ParseFloat(att.Val, 64)
					break
				}
			}
			break
		}
		break
	}
	if err == nil {
		for nn := range n.Find("span.einheitlabel") {
			if nn.FirstChild != nil {
				cntUnitsStr := strings.SplitN(strings.TrimSpace(nn.FirstChild.Data), "er", 2)[0]
				var cntUnits int
				if cntUnits, err = strconv.Atoi(cntUnitsStr); err == nil {
					if cntUnits > 0 {
						o.Price = float64(int(costs*100) / cntUnits)
						o.Price = o.Price / 100

					} else {
						err = fmt.Errorf("could not define price")
					}
				}
			}
			break
		}
	}

	return err
}

func readAttributes(n htmlfilter.Node, o *storage.Record) error {
	var err error
	for nn := range n.Find("div#tab-pane-data") {
		for nnn := range nn.Find("div.VariantInfo-item") {
			for attr := range nnn.Find("div.ws-g.ws-c") {
				var k, v string
				for attrK := range attr.Find("div.VariantInfo-itemName") {
					if attrK.LastChild != nil {
						k = attrK.LastChild.Data
					}
					break
				}
				for attrV := range attr.Find("div.VariantInfo-itemValue") {
					switch k {
					case "Länge", "Length", "Ringmaß / Durchmesser", "Ring / Diameter":
						val := attrV.Node
						if val.FirstChild != nil {
							val = val.FirstChild
							for val.LastChild != nil {
								val = val.LastChild
								if val.DataAtom == 0 {
									v = val.Data
									break
								}
							}
						}
						err = errors.Join(err, setAttribute(o, k, v))
					}
					val := attrV.Node
					for val.LastChild != nil {
						val = val.LastChild
						if val.DataAtom == 0 {
							v = val.Data
							break
						}
					}
					break
				}
				err = errors.Join(err, setAttribute(o, k, v))
			}
		}
		break
	}
	return err
}

func setAttribute(o *storage.Record, k string, v string) error {
	var err error
	switch k {
	case "Brand", "Marke":
		o.Brand = v

	case "Size", "Format":
		o.Format = v

	case "Produkt", "Item":
		o.Series = v

	case "Fabrication", "Herstellungsart":
		o.TypeOfManufacturing = &v

	case "Aromatisiert":
		switch v {
		case "nein", "Nein":
			o.IsFlavoured = pointer(false)
		case "ja", "Ja":
			o.IsFlavoured = pointer(true)
		}

	case "Flavoured":
		switch v {
		case "no", "No":
			o.IsFlavoured = pointer(false)
		case "yes", "Yes":
			o.IsFlavoured = pointer(true)
		}

	case "Boxpressed":
		switch v {
		case "no", "No", "nein", "Nein":
			o.IsBoxpressed = pointer(false)
		case "yes", "Yes", "ja", "Ja":
			o.IsBoxpressed = pointer(true)
		}

	case "Tabacalera":
		o.Maker = &v

	case "Binder origin", "Umblatt Land":
		o.BinderOrigin = splitCommaseparatedVals(v)

	case "Outer leaf tobacco variety", "Umblatt Tabaksorte":
		o.BinderTobaccoVariety = &v

	case "Outer leaf tobacco property", "Umblatt Eigenschaft":
		o.BinderProperty = &v

	case "Filler origin", "Einlage Land":
		o.FillerOrigin = splitCommaseparatedVals(v)

	case "Einlage Tabaksorte":
		o.FillerTobaccoVariety = &v

	case "Einlage Eigenschaft":
		o.FillerProperty = &v

	case "Wrapper origin", "Deckblatt Land":
		o.WrapperOrigin = splitCommaseparatedVals(v)

	case "Topsheet / -leave tobacco variety", "Deckblatt Tabaksorte":
		o.WrapperTobaccoVariety = &v

	case "Topsheet / -leave property", "Deckblatt Eigenschaft":
		o.WrapperProperty = &v

	case "Length", "Länge":
		s := strings.SplitN(v, " ", 2)[0]
		switch {
		case strings.HasSuffix(v, "inches"):
			o.LengthInch, err = strconv.ParseFloat(s, 64)
		case strings.HasSuffix(v, "cm"):
			if o.Length, err = strconv.ParseFloat(s, 64); err == nil {
				// fix rounding
				// e.g. 18.42 -> 184.2000...02
				// FIXME(?) can it be done more efficiently?
				tmp := int(o.Length * 100)
				o.Length = float64(tmp/10) + float64(tmp-(tmp/10)*10)/10
			}
		}

	case "Ring / Diameter", "Ringmaß / Durchmesser":
		s := strings.SplitN(v, " ", 2)[0]
		switch {
		case strings.HasSuffix(v, "cm"):
			if o.Diameter, err = strconv.ParseFloat(s, 64); err == nil {
				tmp := int(o.Diameter * 100)
				o.Diameter = float64(tmp/10) + float64(tmp-(tmp/10)*10)/10
			}
		default:
			o.Ring, err = strconv.Atoi(s)
		}
	}
	return err
}

func splitCommaseparatedVals(v string) []string {
	var o []string
	for _, vv := range strings.Split(v, ",") {
		o = append(o, strings.TrimSpace(vv))
	}
	return o
}

func pointer[V bool | int | string](v V) *V {
	return &v
}

func readName(n htmlfilter.Node) string {
	var s string
	for nn := range n.Find("h1.h-alt") {
		if nn.LastChild != nil {
			s = nn.LastChild.Data
		}
		break
	}
	return s
}

const baseURL = "https://www.cigarworld.de"

func (c Client) ReadBulk(ctx context.Context, _, page uint) (r []storage.Record, nextPage uint, err error) {
	var resp *http.Response

	// read root page to extract the URL for the given page
	var (
		pages     map[string]string
		pageQuery string
	)
	resp, err = c.HTTPClient.Get(baseURL + "/zigarren")
	if err == nil {
		pages, err = newPaginator(resp.Body)
		_ = resp.Body.Close()
		if err == nil {
			var ok bool
			if pageQuery, ok = pages[fmt.Sprintf("%d", page)]; !ok {
				err = fmt.Errorf("page %d not found", page)
			}
		}
	}

	// read the urls to the Brands, or items present on the given page
	var candidateURLPaths []string
	if err == nil {
		if resp, err = c.HTTPClient.Get(fmt.Sprintf("%s/zigarren?%s", baseURL, pageQuery)); err == nil {
			candidateURLPaths, err = newCandidateURLPaths(resp.Body)
			_ = resp.Body.Close()
		}
	}

	// extract the records
	if err == nil {
		r, err = c.extractRecords(ctx, candidateURLPaths)
	}

	// increment paginator
	if err == nil {
		if _, ok := pages[fmt.Sprintf("%d", page)]; ok {
			nextPage = page + 1
		}
	}

	return r, nextPage, err
}

func (c Client) extractRecords(ctx context.Context, candidateURLPaths []string) (r []storage.Record, err error) {
	var wg = new(sync.WaitGroup)
	var mu = new(sync.Mutex)
	const maxGoroutines = 70
	var sem = make(chan struct{}, maxGoroutines)

	for _, s := range candidateURLPaths {
		wg.Add(1)
		sem <- struct{}{}

		go func(wg *sync.WaitGroup) {
			defer func() {
				<-sem
				wg.Done()
			}()
			candidateURL := baseURL + s
			resp, er := c.HTTPClient.Get(candidateURL)
			if er != nil {
				err = errors.Join(err, fmt.Errorf("error fetching %s: %w", candidateURL, er))
				return
			}

			body, er := io.ReadAll(resp.Body)
			if er != nil {
				err = errors.Join(err, fmt.Errorf("error reading data from %s: %w", candidateURL, er))
				return
			}
			_ = resp.Body.Close()

			u, er := extractURLs(body)
			if er != nil {
				err = errors.Join(err, fmt.Errorf("could not process the candidate page: %w", er))
				return
			}

			switch len(u) {
			case 0:
				var rec storage.Record
				er := readDetailsPage(bytes.NewReader(body), &rec)
				mu.Lock()
				switch er {
				case nil:
					r = append(r, rec)
				default:
					err = errors.Join(err, fmt.Errorf("failed to extract record from %s: %w", candidateURL, er))
				}
				mu.Unlock()

			default:
				u = filterBulkURLs(u)
				for _, url := range u {
					wg.Add(1)
					sem <- struct{}{}
					go func(wg *sync.WaitGroup) {
						defer func() {
							<-sem
							wg.Done()
						}()

						id := baseURL + url
						rec, er := c.Read(ctx, id)
						mu.Lock()
						switch er {
						case nil:
							r = append(r, rec)

						default:
							err = errors.Join(err, fmt.Errorf("failed to extract record from %s: %w", url, er))
						}
						mu.Unlock()
					}(wg)
				}
			}
		}(wg)
	}

	wg.Wait()
	return r, err
}

func extractURLs(v []byte) (urls []string, err error) {
	var n *html.Node
	if n, err = html.Parse(bytes.NewReader(v)); err == nil {
		nn := htmlfilter.Node{Node: n}
		for nnn := range nn.Find("div.tab-pane") {
			for ul := range nnn.Find("ul.DetailVariant-list") {
				for li := range ul.Find("li.DetailVariant") {
					for a := range li.Find("a.DetailVariant-col.DetailVariant-data") {
						for _, att := range a.Attr {
							if att.Key == "href" {
								urls = append(urls, att.Val)
							}
						}
					}
				}
			}
		}
	}
	return urls, err
}

func filterBulkURLs(v []string) []string {
	filters := map[string]struct{}{"humidor": {}, "samples": {}, "jar": {}, "kiste": {}, "set": {}}
	var o = make([]string, 0, len(v))
	for _, el := range v {
		if _, toFilter := filters[el]; !toFilter {
			o = append(o, el)
		}
	}
	return o
}

func newCandidateURLPaths(v io.Reader) (o []string, err error) {
	var n *html.Node
	if n, err = html.Parse(v); err == nil {
		nn := htmlfilter.Node{Node: n}
		for nn = range nn.Find("div.ws-g.search-result") {
			for nnn := range nn.Find("div.search-result-item") {
				for a := range nnn.Find("a.search-result-item-inner") {
					for _, att := range a.Attr {
						switch att.Key {
						case "href":
							o = append(o, att.Val)
						}
					}
				}
			}
		}
	}
	return o, err
}

func newPaginator(v io.Reader) (o map[string]string, err error) {
	var n *html.Node
	if n, err = html.Parse(v); err == nil {
		nn := htmlfilter.Node{Node: n}

		for nnn := range nn.Find("select#pagination_select") {
			var urlQueryKey string
			for _, att := range nnn.Attr {
				if att.Key == "name" {
					urlQueryKey = att.Val
					break
				}
			}
			if urlQueryKey != "" {
				o = make(map[string]string)
				for nnn := range nnn.Find("option") {
					if nnn.LastChild != nil {
						pageID := nnn.LastChild.Data
						for _, att := range nnn.Attr {
							if att.Key == "value" {
								o[pageID] = fmt.Sprintf("%s=%s", urlQueryKey, att.Val)
								break
							}
						}
					}
				}
			}
		}
	}
	return o, err
}
