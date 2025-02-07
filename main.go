package main

import (
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
	f, err := os.OpenFile(fName, os.O_CREATE|os.O_TRUNC, 0664)
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
		log.Printf("could write data to %s: %v", fName, err)
	}
}

type httpClient interface {
	Get(url string) (resp *http.Response, err error)
}

type Record struct {
	Name            string   `json:"name"`
	URL             string   `json:"url"`
	Brand           string   `json:"brand"`
	Collection      string   `json:"collection"`
	OriginCountry   string   `json:"manufacturerCountry"`
	Format          string   `json:"format"`
	Form            string   `json:"form"`
	Blender         string   `json:"blender"`
	FillingType     string   `json:"fillingType"`
	WrapperType     string   `json:"wrapperType"`
	WrapperCountry  []string `json:"wrapperCountry"`
	FillerCountry   []string `json:"fillingCountry"`
	BinderCountry   []string `json:"binderCountry"`
	Strength        string   `json:"strength"`
	Price           float64  `json:"price"`
	Diameter        float64  `json:"diameter_mm"`
	Gauge           int      `json:"gauge"`
	Length          int      `json:"length_mm"`
	Aroma           string   `json:"aroma"`
	Aromas          []string `json:"aromas"`
	SmokingDuration string   `json:"smokingDuration"`
	LimitedEdition  bool     `json:"limitedEdition"`
	include         bool
}

func ExtractNobelgo(c httpClient) (o []Record, err error, warn error) {
	page := 1
	for page < 2 {
		log.Printf("fetch data from page %d\n", page)

		baseURL := fmt.Sprintf("https://www.noblego.de/zigarren/?in_stock=1&limit=96&p=%d", page)
		var (
			er   error
			resp *http.Response
		)
		if resp, er = c.Get(baseURL); er == nil {
			v, found, er, warnInner := readPage(resp.Body, c)
			if er == nil {
				_ = resp.Body.Close()
				if found {
					o = append(o, v...)
				} else {
					break
				}
			}
			warn = errors.Join(warn, warnInner)
		}
		if er != nil {
			err = errors.Join(err, fmt.Errorf("error reading page %d: %w", page, er))
		}
		page++
	}
	return o, err, warn
}

func readPage(v io.Reader, c httpClient) (o []Record, found bool, err error, warn error) {
	o, err, warn = readList(v)
	var flag = make(chan struct{}, len(o))
	for range o {
		flag <- struct{}{}
	}
	var mu = new(sync.Mutex)
	for i, el := range o {
		i := i
		el := el
		go func() {
			var (
				er   error
				resp *http.Response
			)
			if resp, er = c.Get(el.URL); er == nil {
				if er = readDetails(resp.Body, &el); er == nil {
					o[i] = el
				}
				_ = resp.Body.Close()
			}
			if er != nil {
				mu.Lock()
				err = errors.Join(err, fmt.Errorf("error reading details using %s: %w", el.URL, er))
				mu.Unlock()
			}
			<-flag
		}()
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
		for nn := range doc.Descendants() {
			if nn.Type == html.ElementNode && nn.DataAtom == atom.Li {
				for _, attr := range nn.Attr {
					if attr.Key == "class" && attr.Val == "item" {
						log.Printf("fetching data for the item %d\n", cnt)

						go func(wg *sync.WaitGroup) {
							wg.Add(1)
							defer wg.Done()
							r, er := readListElement(nn)
							mu.Lock()
							defer mu.Unlock()
							if er == nil {
								o = append(o, r)
							} else {
								warn = errors.Join(warn, fmt.Errorf("item %d: %w", cnt, er))
							}
						}(wg)

						cnt++
					}
				}
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
	spanReader := func(n *html.Node, fn func(n *html.Node) string) (o string) {
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

	dataFromFirstSpanChild := func(n *html.Node) string {
		return spanReader(n, func(n *html.Node) string { return n.FirstChild.Data })
	}
	dataFromATagText := func(n *html.Node) string {
		return spanReader(n, func(n *html.Node) string {
			for nnn := range n.Descendants() {
				if nnn.DataAtom == atom.A {
					return nnn.LastChild.Data
				}
			}
			return ""
		})
	}

	for nn := range n.Descendants() {
		if nn.DataAtom == atom.Li {
			for _, att := range nn.Attr {
				if att.Key == "class" {
					switch att.Val {
					case "product-attribute-herkunft":
						r.OriginCountry = dataFromATagText(nn)

					case "product-attribute-cig_size":
						r.Format = dataFromATagText(nn)

					case "product-attribute-cig_wrapper_origin":
						r.WrapperCountry = parseConcatSlice(dataFromFirstSpanChild(nn))

					case "product-attribute-cig_filler":
						r.FillerCountry = parseConcatSlice(dataFromFirstSpanChild(nn))

					case "product-attribute-cig_aroma":
						r.Aromas = parseConcatSlice(dataFromFirstSpanChild(nn))

					case "product-attribute-cig_form":
						r.Form = dataFromATagText(nn)

					}
				}
			}
		}
	}
}

func parseConcatSlice(s string) []string {
	o := strings.Split(s, ",")
	for i, v := range o {
		o[i] = strings.TrimSpace(v)
	}
	return o
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
	if doc, er := html.Parse(v); er == nil {
		for n := range doc.Descendants() {
			if n.DataAtom == atom.Div {
				for _, att := range n.Attr {
					if att.Key == "class" && strings.HasPrefix(att.Val, "product-secondary-column") {

					}
				}
			}
		}
	}
	return err
}
