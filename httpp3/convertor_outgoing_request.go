package httpp3

import (
	"fmt"
	"io"
	"net/http"

	types "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	wit "go.bytecodealliance.org/pkg/wit/types"
)

// NewOutgoingHttpRequest converts a standard net/http Request to a WASI HTTP 0.3.0 Request.
func NewOutgoingHttpRequest(req *http.Request) (*types.Request, error) {
	// build headers
	headers := types.MakeFields()
	for key, vals := range req.Header {
		for _, val := range vals {
			if result := headers.Append(key, []uint8(val)); result.IsErr() {
				return nil, fmt.Errorf("failed to set header %s", key)
			}
		}
	}

	// handle body
	var bodyOpt wit.Option[*wit.StreamReader[uint8]]
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body.Close()
		if len(bodyBytes) > 0 {
			tx, rx := types.MakeStreamU8()
			go func() {
				defer tx.Drop()
				tx.WriteAll(bodyBytes)
			}()
			bodyOpt = wit.Some(rx)
		} else {
			bodyOpt = wit.None[*wit.StreamReader[uint8]]()
		}
	} else {
		bodyOpt = wit.None[*wit.StreamReader[uint8]]()
	}

	request, send := types.RequestNew(
		headers,
		bodyOpt,
		trailersFuture(),
		wit.None[*types.RequestOptions](),
	)
	send.Drop()

	request.SetMethod(toWasiMethod(req.Method))

	// set authority
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	request.SetAuthority(wit.Some(host))

	// set path with query
	path := req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}
	request.SetPathWithQuery(wit.Some(path))

	// set scheme
	switch req.URL.Scheme {
	case "http":
		request.SetScheme(wit.Some(types.MakeSchemeHttp()))
	case "https":
		request.SetScheme(wit.Some(types.MakeSchemeHttps()))
	default:
		if req.URL.Scheme != "" {
			request.SetScheme(wit.Some(types.MakeSchemeOther(req.URL.Scheme)))
		}
	}

	return request, nil
}

func toWasiMethod(s string) types.Method {
	switch s {
	case http.MethodConnect:
		return types.MakeMethodConnect()
	case http.MethodDelete:
		return types.MakeMethodDelete()
	case http.MethodGet:
		return types.MakeMethodGet()
	case http.MethodHead:
		return types.MakeMethodHead()
	case http.MethodOptions:
		return types.MakeMethodOptions()
	case http.MethodPatch:
		return types.MakeMethodPatch()
	case http.MethodPost:
		return types.MakeMethodPost()
	case http.MethodPut:
		return types.MakeMethodPut()
	case http.MethodTrace:
		return types.MakeMethodTrace()
	default:
		return types.MakeMethodOther(s)
	}
}
