// Package noblego defines the logic and the standard implementation to extract data from noblego.de
package noblego

import (
	"cigarsdb/storage"
	"context"
	"net/http"
)

// Client defines the client to noblego.de to fetch data from.
type Client struct {
	HTTPClient HTTPClient
}

func (c Client) Read(ctx context.Context, id string) (r storage.Record, err error) {
	//TODO implement me
	panic("implement me")
}

func (c Client) ReadBulk(ctx context.Context, limit, page uint) (r []storage.Record, nextPage uint, err error) {
	//TODO implement me
	panic("implement me")
}

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}
