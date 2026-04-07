package httpp3

import (
	"fmt"
	"io"
	"net/http"

	types "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
)

// NewHttpRequest converts a WASI HTTP 0.3.0 Request to a standard net/http Request.
func NewHttpRequest(request *types.Request) (req *http.Request, err error) {

	// convert the http method to string
	method, err := methodToString(request.GetMethod())
	if err != nil {
		return nil, err
	}

	fmt.Printf("METHOD: %s\n", method)

	// convert the path with query to a url
	var url string
	if pathWithQuery := request.GetPathWithQuery(); pathWithQuery.IsNone() {
		url = ""
	} else {
		url = pathWithQuery.Some()
	}

	fmt.Printf("URL: %s\n", url)

	// consume the request body
	// rx, trailers := types.RequestConsumeBody(request, unitFuture())
	// trailers.Drop()

	var body io.Reader
	// body = NewReader(rx)

	// create a new request
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("%#v\n", req)

	fmt.Println("REQUEST:")
	fmt.Printf("%#v\n", request)
	fmt.Println("HANDLE:")
	fmt.Printf("%#v\n", request.Handle())
	fmt.Println("HEADERS:")
	fmt.Printf("%#v\n", request.GetHeaders())
	fmt.Println("MOVING ON...")

	// update additional fields
	toHttpHeader(request.GetHeaders(), &req.Header)

	return req, nil
}

func methodToString(m types.Method) (string, error) {
	switch m.Tag() {
	case types.MethodConnect:
		return "CONNECT", nil
	case types.MethodDelete:
		return "DELETE", nil
	case types.MethodGet:
		return "GET", nil
	case types.MethodHead:
		return "HEAD", nil
	case types.MethodOptions:
		return "OPTIONS", nil
	case types.MethodPatch:
		return "PATCH", nil
	case types.MethodPost:
		return "POST", nil
	case types.MethodPut:
		return "PUT", nil
	case types.MethodTrace:
		return "TRACE", nil
	case types.MethodOther:
		return m.Other(), fmt.Errorf("unknown http method 'other'")
	default:
		return "", fmt.Errorf("failed to convert http method")
	}
}

func toHttpHeader(fields *types.Fields, dest *http.Header) {
	for _, entry := range fields.CopyAll() {
		key := entry.F0
		value := string(entry.F1)
		dest.Add(key, value)
	}
}
