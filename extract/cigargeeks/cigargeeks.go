package cigargeeks

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

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type Client struct {
	HTTPClient extract.HTTPClient
}

const cookie = "SMFCookie895=%7B%220%22%3A133942%2C%221%22%3A%22e30e17daccc11bb313e8a418283f6f3d2743743b0f084407b3" +
	"aebaabbdd61c5dc2a229aab8047ff6b784d8d27fc2515be6a5663d256a73bd2384080eac8bcdfc" +
	"%22%2C%222%22%3A1934324190%2C%223%22%3A%22www.cigargeeks.com%22%2C%224%22%3A%22%5C%2F%22%7D;" +
	" PHPSESSID=baa432720e6018581b89f2ec76b6ed61"

func (c Client) Read(ctx context.Context, id string) (r storage.Record, err error) {
	var res any
	res, err = c.processReq(ctx, id, readDetailsPage)
	if err == nil {
		r = res.(storage.Record)
		r.URL = id
	}
	return r, err
}

func readDetailsPage(_ context.Context, v io.ReadCloser) (any, error) {
	var o = storage.Record{}
	var err error
	var doc *html.Node
	if doc, err = html.Parse(v); err == nil {
		n := htmlfilter.Node{Node: doc}
		// 12:
		// - Brand
		// - Name
		// - Length
		// - Ring Gauge
		// - Country of Origin
		// - Filler
		// - Binder
		// - Wrapper
		// - Color
		// - Strength
		// - Shape
		// - Notes
		const l = 12
		var attrNames = make([]string, 0, l)
		var attrVals = make([]string, 0, l)
		for el := range n.Find("div.main_section") {
			for attName := range el.Find("dt") {
				for c := range attName.ChildNodes() {
					if c.DataAtom == atom.Strong {
						val := strings.TrimSpace(c.LastChild.Data)
						attrNames = append(attrNames, val)
						break
					}
				}
			}
			for attVal := range el.Find("dd") {
				var attrs = make([]string, 0)
				for c := range attVal.ChildNodes() {
					if c.DataAtom != atom.Br {
						if val := strings.TrimSpace(c.Data); val != "" {
							attrs = append(attrs, val)
						}
					}
				}
				var buf = new(bytes.Buffer)
				for i, attr := range attrs {
					buf.WriteString(attr)
					if i < len(attrs)-1 {
						buf.WriteString("\n")
					}
				}
				attrVals = append(attrVals, buf.String())
			}
			break
		}

		var er error
		for i, attrName := range attrNames {
			attrVal := attrVals[i]

			switch {
			case strings.HasPrefix(attrName, "Brand"):
				o.Brand = strings.TrimSpace(attrVal)

			case strings.HasPrefix(attrName, "Name"):
				o.Name = strings.TrimSpace(attrVal)

			case strings.HasPrefix(attrName, "Length"):
				if o.LengthInch, er = strconv.ParseFloat(attrVal, 64); er != nil {
					err = errors.Join(err, fmt.Errorf("could not parse %s as float: %w", attrName, er))
				}

			case strings.HasPrefix(attrName, "Ring Gauge"):
				if o.Ring, er = strconv.ParseFloat(attrVal, 64); er != nil {
					err = errors.Join(err, fmt.Errorf("could not parse %s as float: %w", attrName, er))
				}
			case strings.HasPrefix(attrName, "Country of Origin"):
				o.ManufactureOrigin = strings.TrimSpace(attrVal)

			case strings.Contains(attrName, "Filler"):
				o.FillerOrigin = readTobaccoOrigin(attrVal)

			case strings.Contains(attrName, "Binder"):
				o.BinderOrigin = readTobaccoOrigin(attrVal)

			case strings.Contains(attrName, "Wrapper"):
				o.WrapperOrigin = readTobaccoOrigin(attrVal)

			case strings.Contains(attrName, "Color"):
				if val := strings.TrimSpace(attrVal); val != "" {
					o.Color = &val
				}

			case strings.Contains(attrName, "Strength"):
				if val := strings.TrimSpace(attrVal); val != "" {
					o.Strength = &val
				}

			case strings.Contains(attrName, "Shape"):
				o.Format = strings.TrimSpace(attrVal)

			case strings.Contains(attrName, "Notes"):
				if val := strings.TrimSpace(attrVal); val != "" {
					o.AdditionalNotes = &val
				}
			}
		}
	}
	return o, err
}

func readTobaccoOrigin(s string) []string {
	vv := strings.Split(strings.TrimSpace(s), "\n")
	var tmp = make([]string, 0, len(vv))
	for _, val := range vv {
		switch strings.Contains(val, "<br") {
		case false:
			tmp = append(tmp, strings.TrimSpace(val))
		}
	}
	var o = make([]string, len(tmp))
	for i, val := range tmp {
		o[i] = val
	}
	return o
}

func (c Client) ReadBulk(ctx context.Context, _, page uint) (r []storage.Record, nextPage uint, err error) {
	if page == 0 {
		page = 1
	}
	skip := (int(page) - 1) * 50
	url := fmt.Sprintf("https://www.cigargeeks.com/index.php?action=cigars;area=showsearch;start=%d", skip)

	var brands = make([]string, 0, 50)
	_, err = c.processReq(ctx, url, func(_ context.Context, v io.ReadCloser) (any, error) {
		var doc *html.Node
		if doc, err = html.Parse(v); err == nil {
			n := htmlfilter.Node{Node: doc}
			for tab := range n.Find("ul#brands_list") {
				for li := range tab.Find("li") {
					for el := range li.Find("a") {
						if el.DataAtom == atom.A {
							brands = append(brands, strings.TrimSpace(el.LastChild.Data))
						}
					}
				}
			}
		}
		return brands, err
	})

	return nil, 0, err
}

func (c Client) processReq(ctx context.Context, url string,
	fn func(ctx context.Context, v io.ReadCloser) (any, error)) (any, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		err = fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Add("Cookie", cookie)

	var resp *http.Response
	resp, err = c.HTTPClient.Do(req)
	defer func() { _ = resp.Body.Close() }()

	switch err == nil {
	case true:
		return fn(ctx, resp.Body)

	case false:
		var respBytes []byte
		var er error
		respBytes, er = io.ReadAll(resp.Body)
		switch er == nil {
		case true:
			err = fmt.Errorf("error reading %s, body: %s, error: %w", url, respBytes, err)
		case false:
			err = fmt.Errorf("error reading %s, error: %w", url, err)
			err = errors.Join(err, fmt.Errorf("could node read the response: %w", er))
		}
	}

	return nil, err
}
