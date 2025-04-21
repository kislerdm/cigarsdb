package extract

import (
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
