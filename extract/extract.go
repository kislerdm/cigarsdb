package extract

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

type MockHTTP struct {
	Body      io.ReadCloser
	BodyRoute map[string]io.ReadCloser
	BodyReq   map[*http.Request]io.ReadCloser
	Err       error
}

func (m MockHTTP) Get(url string) (*http.Response, error) {
	var r *http.Response
	if m.Err == nil {
		r = &http.Response{
			StatusCode: http.StatusOK,
			Body:       m.Body,
		}
		if v, ok := m.BodyRoute[url]; ok {
			r.Body = v
		}
	}
	return r, m.Err
}

func (m MockHTTP) Do(req *http.Request) (*http.Response, error) {
	var r *http.Response
	if m.Err == nil {
		r = &http.Response{
			StatusCode: http.StatusOK,
			Body:       m.Body,
		}
		if v, ok := m.BodyReq[req]; ok {
			r.Body = v
		}
	}
	return r, m.Err
}

func ProcessReq(ctx context.Context, c HTTPClient, url string, headers http.Header,
	fn func(ctx context.Context, v io.ReadCloser) (any, error)) (o any, err error) {
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		err = fmt.Errorf("could not create request: %w", err)
	}
	req.Header = headers

	var resp *http.Response
	resp, err = c.Do(req)

	switch err == nil {
	case true:
		o, err = fn(ctx, resp.Body)
		_ = resp.Body.Close()

	case false:
		var respBytes []byte
		var er error
		if resp != nil {
			respBytes, er = io.ReadAll(resp.Body)
			switch er == nil {
			case true:
				err = fmt.Errorf("error reading %s, body: %s, error: %w", url, respBytes, err)
			case false:
				err = fmt.Errorf("error reading %s, error: %w", url, err)
				err = errors.Join(err, fmt.Errorf("could node read the response: %w", er))
			}
		} else {
			err = fmt.Errorf("error reading %s, error: %w", url, err)
		}
	}

	return o, err
}
