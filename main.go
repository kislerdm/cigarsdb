package main

import (
	"cigarsdb/htmlfilter"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	fName := "/tmp/cigardsdb_nobelgo.json"
	f, err := os.OpenFile(fName, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		log.Printf("could not open file: %v", err)
		return
	}
	defer func() { _ = f.Close() }()

	data, err, warn := ExtractNobelgo(http.DefaultClient)
	if err != nil {
		log.Println(err)
		return
	}
	if warn != nil {
		log.Printf("warnings: \n%v\n", warn)
	}

	if err = json.NewEncoder(f).Encode(data); err != nil {
		log.Printf("could not write data to %s: %v", fName, err)
	}
}

type httpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type Record struct {
	Name               string   `json:"name"`
	URL                string   `json:"url"`
	Brand              string   `json:"brand"`
	Series             string   `json:"series"`
	ManufactureCountry string   `json:"manufactureCountry"`
	Format             string   `json:"format"`
	Form               string   `json:"form"`
	Maker              string   `json:"maker"`
	Construction       string   `json:"construction"`
	WrapperType        string   `json:"wrapperType"`
	Strength           string   `json:"strength"`
	FlavourStrength    string   `json:"flavourStrength"`
	SmokingDuration    string   `json:"smokingDuration"`
	Special            string   `json:"special"`
	WrapperOrigin      []string `json:"wrapperOrigin"`
	FillerOrigin       []string `json:"fillerOrigin"`
	BinderOrigin       []string `json:"binderOrigin"`
	Aroma              []string `json:"aroma"`
	Price              float64  `json:"price"`
	Diameter           float64  `json:"diameter_mm"`
	Gauge              int      `json:"gauge"`
	Length             int      `json:"length_mm"`
	LimitedEdition     bool     `json:"limitedEdition"`
	include            bool
}

func ExtractNobelgo(c httpClient) (o []Record, err error, warn error) {
	page := 1
	maxPage := 2
	for page < maxPage {
		log.Printf("fetch data from page %d\n", page)

		baseURL := fmt.Sprintf("https://www.noblego.de/zigarren/?in_stock=1&limit=10&p=%d", page)
		var (
			er    error
			resp  *http.Response
			found bool
		)
		if resp, er = c.Get(baseURL); er == nil {
			var (
				warnInner error
				v         []Record
			)
			v, found, er, warnInner = readPage(resp.Body, c)
			_ = resp.Body.Close()
			warn = errors.Join(warn, warnInner)
			if er == nil {
				o = append(o, v...)
			}
		}
		if er != nil {
			err = errors.Join(err, fmt.Errorf("error reading page %d: %w", page, er))
		}
		if !found {
			break
		}
		page++
	}
	return o, err, warn
}

func readPage(v io.Reader, c httpClient) (o []Record, found bool, err error, warn error) {
	o, err, warn = readList(v)
	var flags = make(chan struct{}, len(o))
	for i, el := range o {
		i := i
		el := el
		go func() {
			var (
				er   error
				resp *http.Response
			)
			if resp, er = c.Get(el.URL); er == nil {
				log.Printf("fetch details for item %d from %s\n", i, el.URL)
				if er = readDetails(resp.Body, &el); er == nil {
					o[i] = el
				}
				_ = resp.Body.Close()
			}
			if er != nil {
				err = errors.Join(err, fmt.Errorf("error reading details using %s: %w", el.URL, er))
			}
			flags <- struct{}{}
			log.Printf("fetch details for item %d from %s\n: close", i, el.URL)
		}()
	}
	maxN := len(o)
	for maxN > 0 {
		<-flags
		maxN--
	}
	return o, len(o) > 0, err, warn
}

func readList(v io.Reader) (o []Record, err error, warn error) {
	var (
		cnt int
		mu  = new(sync.Mutex)
		wg  = new(sync.WaitGroup)
	)
	if doc, er := html.Parse(v); er == nil {
		n := htmlfilter.Node{Node: doc}

		for n = range n.Find("div.category-products") {
			for nn := range n.Find("li.item") {
				log.Printf("fetching data for the item %d\n", cnt)
				wg.Add(1)
				go func(wg *sync.WaitGroup, n *html.Node) {
					defer wg.Done()
					defer mu.Unlock()

					r, er := readListElement(n)

					mu.Lock()
					if er == nil {
						o = append(o, r)
					} else {
						warn = errors.Join(warn, fmt.Errorf("item %d: %w", cnt, er))
					}
				}(wg, nn.Node)
				cnt++
			}
		}
	} else {
		err = fmt.Errorf("could not parse the list page: %w", er)
	}
	wg.Wait()
	return o, err, warn
}

func readListElement(n *html.Node) (r Record, warn error) {
	var extractedBaseInfo, extractedAttrs, extractedPrice bool
	for nn := range n.Descendants() {
		if extractedBaseInfo && extractedPrice && extractedAttrs {
			break
		}

		if nn.DataAtom == atom.Div {
			for _, att := range nn.Attr {
				if att.Key == "class" {
					switch {
					case strings.Contains(att.Val, "product-item-info"):
						readListProductBaseInfo(nn, &r)
						extractedBaseInfo = true
					case att.Val == "product-attributes":
						readListProductDetails(nn, &r)
						extractedAttrs = true
					case att.Val == "add-to-cart":
						// the price indicates if the data point shall be skipped
						readListProductPrice(nn, &r)
						extractedPrice = true
					}
				}
			}
		}
	}

	// include the entries which prices are indicated per box
	// it's often an indication of special offers, or items beyond the scope of the database, e.g., etui/cases etc.
	if !r.include {
		r = Record{}
		warn = fmt.Errorf("no required data found")
	}
	return r, warn
}

func readListProductPrice(n *html.Node, r *Record) {
	var c *html.Node
	for nn := range n.Descendants() {
		if r.include {
			break
		}
		if nn.DataAtom == atom.Ul {
			for _, att := range nn.Attr {
				if att.Key == "class" && att.Val == "product-prices" {
					for nnn := range nn.ChildNodes() {
						for nnnn := range nnn.Descendants() {
							if nnnn.DataAtom == atom.Span {
								for _, att := range nnnn.Attr {
									if att.Key == "title" && att.Val == "Verpackungseinheit" &&
										nnnn.LastChild.Data == "Einzeln" {
										r.include = true
										c = nnn
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if r.include {
		for nnn := range c.Descendants() {
			if nnn.DataAtom == atom.Span {
				for _, att := range nnn.Attr {
					if att.Key == "class" && att.Val == "price" {
						tmp := nnn.LastChild.Data
						// remove length of euro sign with the unbreakable space
						tmp = tmp[:len(tmp)-5]
						tmp = strings.Replace(tmp, ",", ".", -1)
						r.Price, _ = strconv.ParseFloat(tmp, 64)
					}
				}
			}
		}
	}
}

func readListProductDetails(n *html.Node, r *Record) {
	for nn := range n.Descendants() {
		if nn.DataAtom == atom.Li {
			for _, att := range nn.Attr {
				if att.Key == "class" {
					switch att.Val {
					case "product-attribute-herkunft":
						r.ManufactureCountry = dataFromATagText(nn)

					case "product-attribute-cig_size":
						r.Format = dataFromATagText(nn)

					case "product-attribute-cig_wrapper_origin":
						r.WrapperOrigin = parseConcatSlice(dataFromFirstSpanChild(nn))

					case "product-attribute-cig_filler":
						r.FillerOrigin = parseConcatSlice(dataFromFirstSpanChild(nn))

					case "product-attribute-cig_aroma":
						r.Aroma = parseConcatSlice(dataFromFirstSpanChild(nn))

					case "product-attribute-cig_form":
						r.Form = dataFromATagText(nn)
					}
				}
			}
		}
	}
}

func readListProductBaseInfo(n *html.Node, r *Record) {
	for nn := range n.Descendants() {
		switch nn.DataAtom {
		case atom.H2:
			for _, att := range nn.Attr {
				if att.Key == "class" && att.Val == "product-name" {
					readUrlAndName(nn, r)
				}
			}

		case atom.Span:
			for _, att := range nn.Attr {
				if att.Key == "class" {
					switch att.Val {
					case "product-attribute-cig_diameter":
						v := readSpanChildrenValue(nn)
						r.Diameter, _ = strconv.ParseFloat(v, 64)
					case "product-attribute-cig_length":
						v := readSpanChildrenValue(nn)
						r.Length, _ = strconv.Atoi(v)
					case "product-attribute-strength":
						r.Strength = readSpanChildrenValue(nn)
					}
				}
			}
		}
	}
}

func readSpanChildrenValue(n *html.Node) string {
	var s string
	for nn := range n.Descendants() {
		if nn.DataAtom == atom.Span {
			for _, att := range nn.Attr {
				if att.Key == "class" && att.Val == "value" {
					s = nn.FirstChild.Data
				}
			}
		}
	}
	return s
}

func readUrlAndName(n *html.Node, r *Record) {
	for nn := range n.Descendants() {
		if nn.DataAtom == atom.A {
			for _, at := range nn.Attr {
				switch at.Key {
				case "href":
					r.URL = at.Val
				case "title":
					r.Name = at.Val
				}
			}
		}
	}
}

func readDetails(v io.Reader, o *Record) (err error) {
	var doc *html.Node
	if doc, err = html.Parse(v); err == nil {
		n := htmlfilter.Node{Node: doc}
		for n = range n.Find("li.product-attribute-") {
			for _, attr := range n.Attr {
				if attr.Key == "class" {
					val := n.Node
					switch attr.Val {
					case "product-attribute-brand":
						o.Brand = dataFromATagText(val)
					case "product-attribute-series":
						o.Series = dataFromATagText(val)
					case "product-attribute-cig_duration":
						o.SmokingDuration = dataFromFirstSpanChild(val)
					case "product-attribute-cig_binder":
						o.BinderOrigin = parseConcatSlice(dataFromFirstSpanChild(val))
					case "product-attribute-cig_maker":
						o.Maker = dataFromFirstSpanChild(val)
					case "product-attribute-cig_wrapper_tobacco":
						o.WrapperType = dataFromFirstSpanChild(val)
					case "product-attribute-flavour_strength":
						o.FlavourStrength = dataFromFirstSpanChild(val)
					case "product-attribute-cig_gauge":
						o.Gauge, err = strconv.Atoi(dataFromFirstSpanChild(val))
					case "product-attribute-cig_special":
						specialVal := dataFromFirstSpanChild(val)
						switch {
						case dataFromATagText(val) != "":
							o.LimitedEdition = true
						case specialVal != "":
							o.Special = specialVal
						}
					case "product-attribute-cig_construction":
						o.Construction = dataFromATagText(val)
					case "product-attribute-cig_form":
						o.Form = dataFromFirstSpanChild(val)
					}
				}
			}
		}
	}
	return err
}

func spanReader(n *html.Node, fn func(n *html.Node) string) (o string) {
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

func dataFromFirstSpanChild(n *html.Node) string {
	return spanReader(n, func(n *html.Node) string { return strings.TrimSpace(n.FirstChild.Data) })
}

func dataFromATagText(n *html.Node) string {
	return spanReader(n, func(n *html.Node) string {
		for nnn := range n.Descendants() {
			if nnn.DataAtom == atom.A {
				return strings.TrimSpace(nnn.LastChild.Data)
			}
		}
		return ""
	})
}

func parseConcatSlice(s string) []string {
	o := strings.Split(s, ",")
	for i, v := range o {
		o[i] = strings.TrimSpace(v)
	}
	return o
}
