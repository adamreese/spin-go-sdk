package httpp3

import (
	"fmt"
	"io"
	"net/http"

	client "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_client"
	types "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
)

// NewTransport returns http.RoundTripper backed by Spin SDK
func NewTransport() http.RoundTripper {
	return &Transport{}
}

// Transport implements http.RoundTripper
type Transport struct{}

// RoundTrip makes roundtrip using Spin SDK
func (r *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return Send(req)
}

// NewClient returns a new HTTP client compatible with the Spin SDK
func NewClient() *http.Client {
	return &http.Client{
		Transport: &Transport{},
	}
}

func Send(req *http.Request) (*http.Response, error) {
	wasiReq, err := NewOutgoingHttpRequest(req)
	if err != nil {
		return nil, err
	}

	result := client.Send(wasiReq)
	if result.IsErr() {
		return nil, fmt.Errorf("error sending request: %v", result.Err())
	}

	response := result.Ok()

	// read response headers
	respHeaders := http.Header{}
	for _, entry := range response.GetHeaders().CopyAll() {
		respHeaders.Add(entry.F0, string(entry.F1))
	}

	// consume response body
	rx, trailers := types.ResponseConsumeBody(response, unitFuture())
	trailers.Drop()

	var body io.ReadCloser = NewReadCloser(rx)

	resp := &http.Response{
		StatusCode: int(response.GetStatusCode()),
		Header:     respHeaders,
		Body:       body,
	}

	return resp, nil
}

func Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	return Send(req)
}

func Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return Send(req)
}
