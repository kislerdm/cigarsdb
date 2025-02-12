package cigarworld

import (
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

	"golang.org/x/net/html"
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
				cntUnitsStr := strings.TrimSuffix(strings.TrimSpace(nn.FirstChild.Data), "er")
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

func (c Client) ReadBulk(ctx context.Context, limit, page uint) (r []storage.Record, nextPage uint, err error) {
	//TODO implement me
	// TODO: filter the words "humidor", "sampler", "jar", "kiste" in the lowered text
	panic("implement me")
}
