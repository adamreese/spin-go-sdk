// Package httpp3 contains the helper functions for writing Spin HTTP components
// using the WASI HTTP 0.3.0 interface in Go, as well as for sending outbound HTTP requests.
package httpp3

import (
	"fmt"
	"net/http"
	"os"

	wasihandler "github.com/spinframework/spin-go-sdk/v3/exports/wasi_http_service_0_3_0_rc_2026_03_15/export_wasi_http_0_3_0_rc_2026_03_15_handler"
	_ "github.com/spinframework/spin-go-sdk/v3/exports/wasi_http_service_0_3_0_rc_2026_03_15/wit_exports"
	types "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

func init() {
	wasihandler.Exports.Handle = wasiHandle
}

const (
	// The application base path.
	HeaderBasePath = "spin-base-path"
	// The component route pattern matched, _excluding_ any wildcard indicator.
	HeaderComponentRoot = "spin-component-route"
	// The full URL of the request. This includes full host and scheme information.
	HeaderFullUrl = "spin-full-url"
	// The part of the request path that was matched by the route (including
	// the base and wildcard indicator if present).
	HeaderMatchedRoute = "spin-matched-route"
	// The request path relative to the component route (including any base).
	HeaderPathInfo = "spin-path-info"
	// The component route pattern matched, as written in the component
	// manifest (that is, _excluding_ the base, but including the wildcard
	// indicator if present).
	HeaderRawComponentRoot = "spin-raw-component-route"
	// The client address for the request.
	HeaderClientAddr = "spin-client-addr"
)

// handler is the function that will be called by the http trigger in Spin.
var handler = defaultHandler

// defaultHandler is a placeholder for returning a useful error to stderr when
// the handler is not set.
var defaultHandler = func(http.ResponseWriter, *http.Request) {
	fmt.Fprintln(os.Stderr, "http handler undefined")
}

// Handle sets the handler function for the http trigger.
// It must be set in an init() function.
func Handle(fn func(http.ResponseWriter, *http.Request)) {
	handler = fn
}

var wasiHandle = func(request *types.Request) wit.Result[*types.Response, types.ErrorCode] {
	// convert the incoming request to go's net/http type
	httpReq, err := NewHttpRequest(request)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to convert wasi/http request to http.Request: %s\n", err)
		return makeErrorResponse(http.StatusInternalServerError)
	}

	// create a response writer that buffers the response
	httpRes := NewHttpResponseWriter()

	// run the user's handler
	handler(httpRes, httpReq)

	// convert the buffered response to a wasi/http response
	return httpRes.toWasiResponse()
}

func makeErrorResponse(statusCode int) wit.Result[*types.Response, types.ErrorCode] {
	headers := types.MakeFields()
	response, send := types.ResponseNew(
		headers,
		wit.None[*wit.StreamReader[uint8]](),
		trailersFuture(),
	)
	send.Drop()
	response.SetStatusCode(uint16(statusCode))
	return wit.Ok[*types.Response, types.ErrorCode](response)
}

func trailersFuture() *wit.FutureReader[wit.Result[wit.Option[*types.Fields], types.ErrorCode]] {
	tx, rx := types.MakeFutureResultOptionFieldsErrorCode()
	go tx.Write(wit.Ok[wit.Option[*types.Fields], types.ErrorCode](wit.None[*types.Fields]()))
	return rx
}

func unitFuture() *wit.FutureReader[wit.Result[wit.Unit, types.ErrorCode]] {
	tx, rx := types.MakeFutureResultUnitErrorCode()
	go tx.Write(wit.Ok[wit.Unit, types.ErrorCode](wit.Unit{}))
	return rx
}
