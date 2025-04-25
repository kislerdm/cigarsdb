package cigarcentury

import (
	"cigarsdb/extract"
	"cigarsdb/htmlfilter"
	"cigarsdb/storage"
	"context"
	"io"
	"log/slog"
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
			for el := range n.Find("div.col-12.dato") {
				for ch := range el.ChildNodes() {
					if ch.DataAtom == atom.Div {
						for _, att := range ch.Attr {
							if att.Key == "class" {
								switch att.Val {
								case "valor descripcion":
									chS := htmlfilter.Node{Node: ch}
									var s = make([]string, 0)
									for el := range chS.Find("div.d-none") {
										s = append(s, el.LastChild.Data)
									}
									s = flipArr(s)
									note := strings.Join(s, " ")
									r.AdditionalNotes = &note

								case "nombre":

								}
							}
						}
					}
				}
			}
		}
		return o, err
	})
	return r, err
}

func flipArr(s []string) []string {
	var o = make([]string, len(s))
	for i, el := range s {
		j := len(s) - 1 - i
		o[j] = el
	}
	return o
}
