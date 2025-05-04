package cigarcentury

import (
	"cigarsdb/extract"
	"cigarsdb/htmlfilter"
	"cigarsdb/storage"
	"context"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Client struct {
	HTTPClient extract.HTTPClient
	Logs       *slog.Logger
}

func (c Client) Read(ctx context.Context, id string) (r storage.Record, err error) {
	_, _ = extract.ProcessReq(ctx, c.HTTPClient, id, nil, func(_ context.Context, v io.ReadCloser) (
		any, error) {
		var o = storage.Record{}
		var doc *html.Node
		if doc, err = html.Parse(v); err == nil {
			n := htmlfilter.Node{Node: doc}
			var categories = make([]string, 0)
			var values = make([]*html.Node, 0)
			for el := range n.Find("div.col-12.dato") {
				for ch := range el.ChildNodes() {
					if ch.DataAtom == atom.Div {
						for _, att := range ch.Attr {
							if att.Key == "class" {
								switch {
								case att.Val == "valor descripcion":
									chS := htmlfilter.Node{Node: ch}
									var s = make([]string, 0)
									for el := range chS.Find("div.d-none") {
										s = append(s, el.LastChild.Data)
									}
									s = flipArr(s)
									note := strings.Join(s, " ")
									r.AdditionalNotes = &note

								case att.Val == "nombre":
									categories = append(categories, ch.FirstChild.Data)

								case strings.HasSuffix(att.Val, "strength-rating-show"):
									values = append(values, ch)

								case strings.HasPrefix(att.Val, "valor"):
									values = append(values, ch.FirstChild)
								}
							}
						}
					}
				}
			}
			err = augmentByCategoryValues(&r, categories, values)
		}
		return o, err
	})
	return r, err
}

func augmentByCategoryValues(r *storage.Record, categories []string, values []*html.Node) error {
	var err error
	for i, category := range categories {
		switch category {
		case "Origin":
			r.ManufactureOrigin = strings.TrimSpace(values[i].Data)

		case "Brand":
			b := htmlfilter.Node{values[i]}
			var tmp = make([]string, 0)
			for el := range b.Find("li") {
				for c := range el.ChildNodes() {
					if c.DataAtom == atom.A {
						tmp = append(tmp, strings.TrimSpace(c.LastChild.Data))
					}
				}
			}
			r.Brand = strings.Join(tmp, ", ")

		case "Manufacturer":
			v := strings.TrimSpace(values[i].Data)
			r.Maker = &v

		case "Wrapper":
			if v := readSeparatedValues(values[i], ","); len(v) > 0 {
				r.WrapperOrigin = v
			}

		case "Binder":
			if v := readSeparatedValues(values[i], ","); len(v) > 0 {
				r.BinderOrigin = v
			}

		case "Filler":
			if v := readSeparatedValues(values[i], ","); len(v) > 0 {
				r.FillerOrigin = v
			}

		case "Vitola":
			r.Format = strings.TrimSpace(values[i].Data)

		case "Box-Pressed":
			switch v := strings.ToLower(strings.TrimSpace(values[i].Data)); v {
			case "yes":
				vv := true
				r.IsBoxpressed = &vv
			case "no":
				vv := false
				r.IsBoxpressed = &vv
			}

		case "Strength":
			n := htmlfilter.Node{Node: values[i]}
			for el := range n.Find("input") {
				for _, att := range el.Attr {
					if att.Key == "title" {
						v := strings.TrimSpace(att.Val)
						r.Strength = &v
						break
					}
				}
				break
			}

		case "Flavors":
			v := readSeparatedValues(values[i], ",")
			o := storage.AromaProfileCommunity{
				Weights: make(map[string]float64),
			}
			// I decided to exclude the total number of voters because it does not appear reliable
			// for the flavours votes
			var totalCnt int
			for _, vv := range v {
				els := strings.Split(vv, "(")
				flavour := strings.TrimSpace(els[0])
				votes := readVotes(els[1])
				totalCnt += votes
				o.Weights[flavour] += float64(votes)
			}
			for flavour, votes := range o.Weights {
				o.Weights[flavour] = votes / float64(totalCnt)
			}
			r.AromaProfileCommunity = &o

		case "Length":
			els := strings.Split(values[i].Data, " (")
			if len(els) == 2 {
				r.Length, _ = strconv.ParseFloat(strings.TrimSpace(strings.Split(els[1], "mm")[0]), 64)
			}

		case "Ring Gauge":
			r.Ring, _ = strconv.ParseFloat(strings.TrimSpace(values[i].Data), 64)

		case "Color":
			v := strings.TrimSpace(values[i].Data)
			r.Color = &v

			// case "Specialized Ratings":
			// 	r.SpecializedRatings
		}
	}
	return err
}

// expected input ({{.int}})
func readVotes(s string) int {
	var tmp = make([]rune, 0, len(s))
	for _, el := range strings.TrimSpace(s) {
		switch el {
		case '(', ')':
		default:
			tmp = append(tmp, el)
		}
	}
	o, _ := strconv.ParseInt(string(tmp), 10, 64)
	return int(o)
}

func readSeparatedValues(v *html.Node, delimiter string) []string {
	tmp := strings.Split(v.Data, delimiter)
	var o = make([]string, 0)
	for _, vv := range tmp {
		el := strings.TrimSpace(vv)
		if el != "-" && el != "" {
			o = append(o, el)
		}
	}
	return o
}

func flipArr(s []string) []string {
	var o = make([]string, len(s))
	for i, el := range s {
		j := len(s) - 1 - i
		o[j] = el
	}
	return o
}
