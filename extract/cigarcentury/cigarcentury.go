// Package cigarcentury defines the client to extract data from cigarcentury.com
package cigarcentury

import (
	"cigarsdb/extract"
	"cigarsdb/htmlfilter"
	"cigarsdb/storage"
	"context"
	"errors"
	"fmt"
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
	Dumper     storage.Writer
}

func (c Client) ReadBulk(ctx context.Context, _, page uint) (r []storage.Record, nextPage uint, err error) {
	// whole list of cigars can be extracted using the page value 473
	// note that the number may grow over time
	if page == 0 {
		page = 1
	}
	url := fmt.Sprintf("https://www.cigarcentury.com/en/cigars?page=%d", page)

	if c.Logs != nil {
		c.Logs.Info("LP", slog.String("url", url))
	}

	var urls = make([]string, 0)
	_, _ = extract.ProcessReq(ctx, c.HTTPClient, url, nil,
		func(_ context.Context, v io.ReadCloser) (any, error) {
			var doc *html.Node
			if doc, err = html.Parse(v); err == nil {
				n := htmlfilter.Node{Node: doc}
				for el := range n.Find("div.producto-item") {
					for vv := range el.Find("a") {
						for _, att := range vv.Attr {
							if att.Key == "href" {
								urls = append(urls, att.Val)
								break
							}
						}
						break
					}
				}
			}
			return nil, nil
		})

	if c.Logs != nil {
		c.Logs.Info("found urls", slog.Int("count", len(urls)))
	}

	for _, cigarURL := range urls {
		if c.Logs != nil {
			c.Logs.Info("fetching data", slog.String("url", cigarURL))
		}
		rec, er := c.Read(ctx, cigarURL)
		switch er != nil {
		case true:
			err = errors.Join(err, er)
		case false:
			if !rec.IsEmpty() {
				if c.Dumper != nil {
					if c.Logs != nil {
						c.Logs.Info("write record", slog.String("url", cigarURL))
					}
					if _, er = c.Dumper.Write(ctx, []storage.Record{rec}); er != nil {
						err = errors.Join(err, er)
						if c.Logs != nil {
							c.Logs.Info("write record error", slog.Any("error", er))
						}
					}
				}
				r = append(r, rec)
			}
		}
	}
	return r, nextPage, err
}

func (c Client) Read(ctx context.Context, id string) (r storage.Record, err error) {
	_, _ = extract.ProcessReq(ctx, c.HTTPClient, id, nil, func(_ context.Context, v io.ReadCloser) (
		any, error) {
		var doc *html.Node
		if doc, err = html.Parse(v); err == nil {
			n := htmlfilter.Node{Node: doc}

			for el := range n.Find("h1.nombre-producto") {
				r.Name = el.LastChild.Data
				break
			}

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
			if err = augmentByCategoryValues(&r, categories, values); err == nil {
				r.URL = id
			}
		}
		return storage.Record{}, err
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

		case "Specialized Ratings":
			r.SpecializedRatings = make([]storage.SpecializedRating, 0)
			n := htmlfilter.Node{values[i].Parent}
			for c := range n.Find("div.calificacion_especializada") {
				var rating storage.SpecializedRating
				for v := range c.Find("div.calificacion_valor") {
					rating.RatingOutOf100, _ = strconv.ParseFloat(
						strings.TrimSuffix(v.LastChild.Data, "%"), 64,
					)
					break
				}
				for v := range c.Find("div.calificacion_nombre") {
					rating.Who = strings.TrimSpace(v.LastChild.Data)
					break
				}
				for v := range c.Find("div.calificacion_ano") {
					rating.Year = strings.TrimSpace(v.LastChild.Data)
					break
				}
				r.SpecializedRatings = append(r.SpecializedRatings, rating)
			}
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
