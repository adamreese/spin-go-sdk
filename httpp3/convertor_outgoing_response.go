package httpp3

import (
	"net/http"

	types "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

var _ http.ResponseWriter = &responseOutparamWriter{}

type responseOutparamWriter struct {
	// go httpHeaders are reconciled when building the wasi response
	httpHeaders http.Header

	statuscode int

	// body buffers all written bytes
	body []byte

	// tracks whether WriteHeader has been called
	wroteHeader bool
}

func (row *responseOutparamWriter) Header() http.Header {
	return row.httpHeaders
}

func (row *responseOutparamWriter) Write(buf []byte) (int, error) {
	if !row.wroteHeader {
		row.WriteHeader(http.StatusOK)
	}

	row.body = append(row.body, buf...)
	return len(buf), nil
}

func (row *responseOutparamWriter) WriteHeader(statusCode int) {
	if !row.wroteHeader {
		row.statuscode = statusCode
		row.wroteHeader = true
	}
}

// NewHttpResponseWriter creates a new response writer for WASI HTTP 0.3.0.
func NewHttpResponseWriter() *responseOutparamWriter {
	return &responseOutparamWriter{
		httpHeaders: http.Header{},
	}
}

// toWasiResponse converts the buffered response to a WASI HTTP 0.3.0 Response.
func (row *responseOutparamWriter) toWasiResponse() wit.Result[*types.Response, types.ErrorCode] {
	if row.statuscode == 0 {
		row.statuscode = http.StatusOK
	}

	// build headers
	headers := types.MakeFields()
	for key, vals := range row.httpHeaders {
		for _, val := range vals {
			headers.Append(key, []uint8(val))
		}
	}

	// create body stream
	var bodyOpt wit.Option[*wit.StreamReader[uint8]]
	if len(row.body) > 0 {
		tx, rx := types.MakeStreamU8()
		go func() {
			defer tx.Drop()
			tx.WriteAll(row.body)
		}()
		bodyOpt = wit.Some(rx)
	} else {
		bodyOpt = wit.None[*wit.StreamReader[uint8]]()
	}

	response, send := types.ResponseNew(headers, bodyOpt, trailersFuture())
	send.Drop()
	response.SetStatusCode(uint16(row.statuscode))

	return wit.Ok[*types.Response, types.ErrorCode](response)
}
