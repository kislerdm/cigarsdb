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
	var skipSampler bool
	var doc *html.Node
	if doc, err = html.Parse(v); err == nil {
		n := htmlfilter.Node{Node: doc}

		var extracted bool
		for nn := range n.Find("div.product-name") {
			if extracted {
				break
			}
			for nn = range nn.Find("h1") {
				s := nn.LastChild.Data
				if strings.Contains(strings.ToLower(s), "sampler") {
					skipSampler = true
				} else {
					o.Name = s
				}
				extracted = true
			}
		}

		if !skipSampler {
			err = readAttributes(n, o)
			err = errors.Join(err, readPrice(n, o))
		}
	}
	return err
}

func readPrice(n htmlfilter.Node, o *storage.Record) error {
	var err error

	var lastPriceOption htmlfilter.Node
	var extracted bool
	for n = range n.Find("ul.product-prices") {
		if extracted {
			break
		}
		for lastPriceOption = range n.Find("li") {
		}
		extracted = true
	}

	var cost float64
	for n = range lastPriceOption.Find("span.price") {
		tmp := n.LastChild.Data
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
	if err != nil {
		o.Price = cost / float64(quantity)
	}

	return err
}

func readAttributes(n htmlfilter.Node, o *storage.Record) error {
	var err error
	for n = range n.Find("li.product-attribute-") {
		for _, attr := range n.Attr {
			if attr.Key == "class" {
				var er error

				val := n.Node
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
					if v := dataFromFirstSpanChild(val); v != nil {
						o.ManufactureOrigin = *v
					}
				case "product-attribute-cig_construction":
					o.Construction = dataFromATagText(val)
				case "product-attribute-cig_form":
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
	return spanReader(n, func(n *html.Node) *string { return pointer(strings.TrimSpace(n.FirstChild.Data)) })
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
	//TODO implement me
	panic("implement me")
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

func pointer[V string | bool | float64 | int](s V) *V {
	return &s
}
