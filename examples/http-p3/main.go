package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"

	_ "github.com/spinframework/spin-go-sdk/v3/exports/wasi_http_service_0_3_0_rc_2026_03_15/wit_exports"
	client "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_client"
	types "github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_3_0_rc_2026_03_15_types"
	wittypes "go.bytecodealliance.org/pkg/wit/types"

	spinhttp "github.com/spinframework/spin-go-sdk/v3/httpp3"
)

func init() {
	// handler.Exports.Handle = Handle

	spinhttp.Handle(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("foo", "bar")

		if r.Method == http.MethodGet && r.URL.Path == "/hello" {
			fmt.Fprintln(w, "Hello spinframework!")
		}
	})
}

func Handle(request *types.Request) wittypes.Result[*types.Response, types.ErrorCode] {
	req, err := spinhttp.NewHttpRequest(request)
	if err != nil {
		fmt.Printf("%#v\n", err)
	}

	fmt.Printf("%#v\n", req)

	method := request.GetMethod().Tag()
	path := request.GetPathWithQuery().SomeOr("/")

	if method == types.MethodGet && path == "/hello" {
		// Say hello!

		tx, rx := types.MakeStreamU8()

		go func() {
			defer tx.Drop()
			tx.WriteAll([]uint8("hello, world!"))
		}()

		response, send := types.ResponseNew(
			types.FieldsFromList([]wittypes.Tuple2[string, []uint8]{
				{"content-type", []uint8("text/plain")},
			}).Ok(),
			wittypes.Some(rx),
			trailersFuture(),
		)
		send.Drop()

		return wittypes.Ok[*types.Response, types.ErrorCode](response)

	} else if method == types.MethodGet && path == "/hash-all" {
		// Collect one or more "url" headers, download their contents
		// concurrently, compute their SHA-256 hashes incrementally
		// (i.e. without buffering the response bodies), and stream the
		// results back to the client as they become available.

		urls := make([]string, 0)
		for _, pair := range request.GetHeaders().CopyAll() {
			if pair.F0 == "url" {
				urls = append(urls, string(pair.F1))
			}
		}

		tx, rx := types.MakeStreamU8()

		go func() {
			defer tx.Drop()

			channel := make(chan wittypes.Tuple2[string, string])
			for _, url := range urls {
				go func() {
					channel <- wittypes.Tuple2[string, string]{url, getSha256(url)}
				}()
			}

			for i := 0; i < len(urls); i++ {
				pair := (<-channel)
				tx.WriteAll([]byte(fmt.Sprintf("%v: %v\n", pair.F0, pair.F1)))
			}
		}()

		response, send := types.ResponseNew(
			types.FieldsFromList([]wittypes.Tuple2[string, []uint8]{
				wittypes.Tuple2[string, []uint8]{"content-type", []uint8("text/plain")},
			}).Ok(),
			wittypes.Some(rx),
			trailersFuture(),
		)
		send.Drop()

		return wittypes.Ok[*types.Response, types.ErrorCode](response)

	} else if method == types.MethodPost && path == "/echo" {
		// Echo the request body back to the client without buffering.

		requestHeaders := request.GetHeaders().CopyAll()

		rx, trailers := types.RequestConsumeBody(request, unitFuture())

		responseHeaders := make([]wittypes.Tuple2[string, []uint8], 0, 1)
		for _, pair := range requestHeaders {
			if pair.F0 == "content-type" {
				responseHeaders = append(responseHeaders, pair)
			}
		}

		response, send := types.ResponseNew(
			types.FieldsFromList(responseHeaders).Ok(),
			wittypes.Some(rx),
			trailers,
		)
		send.Drop()

		return wittypes.Ok[*types.Response, types.ErrorCode](response)

	} else {
		// Bad request

		response, send := types.ResponseNew(
			types.MakeFields(),
			wittypes.None[*wittypes.StreamReader[uint8]](),
			trailersFuture(),
		)
		send.Drop()
		response.SetStatusCode(400).Ok()

		return wittypes.Ok[*types.Response, types.ErrorCode](response)

	}
}

// Download the contents of the specified URL, computing the SHA-256
// incrementally as the response body arrives.
//
// This returns a tuple of the original URL and either the hex-encoded hash or
// an error message.
func getSha256(urlString string) string {
	parsed, err := url.Parse(urlString)
	if err != nil {
		return err.Error()
	}

	var scheme types.Scheme
	switch parsed.Scheme {
	case "http":
		scheme = types.MakeSchemeHttp()
	case "https":
		scheme = types.MakeSchemeHttps()
	default:
		scheme = types.MakeSchemeOther(parsed.Scheme)
	}

	request, send := types.RequestNew(
		types.MakeFields(),
		wittypes.None[*wittypes.StreamReader[uint8]](),
		trailersFuture(),
		wittypes.None[*types.RequestOptions](),
	)
	send.Drop()
	request.SetScheme(wittypes.Some(scheme)).Ok()
	request.SetAuthority(wittypes.Some(parsed.Host)).Ok()
	request.SetPathWithQuery(wittypes.Some(parsed.Path)).Ok()

	result := client.Send(request)
	switch result.Tag() {
	case wittypes.ResultOk:
		response := result.Ok()
		status := response.GetStatusCode()
		if status < 200 || status > 299 {
			return fmt.Sprintf("unexpected status: %v", status)
		}

		rx, trailers := types.ResponseConsumeBody(response, unitFuture())
		trailers.Drop()
		defer rx.Drop()

		buffer := make([]uint8, 16*1024)
		hash := sha256.New()
		for !rx.WriterDropped() {
			count := rx.Read(buffer)
			writeCount, err := hash.Write(buffer[:count])
			if err != nil || uint32(writeCount) != count {
				panic("unreachable")
			}
		}
		return hex.EncodeToString(hash.Sum([]uint8{}))

	case wittypes.ResultErr:
		return "error sending request"

	default:
		panic("unreachable")
	}
}

func trailersFuture() *wittypes.FutureReader[wittypes.Result[wittypes.Option[*types.Fields], types.ErrorCode]] {
	tx, rx := types.MakeFutureResultOptionFieldsErrorCode()
	go tx.Write(wittypes.Ok[wittypes.Option[*types.Fields], types.ErrorCode](wittypes.None[*types.Fields]()))
	return rx
}

func unitFuture() *wittypes.FutureReader[wittypes.Result[wittypes.Unit, types.ErrorCode]] {
	tx, rx := types.MakeFutureResultUnitErrorCode()
	go tx.Write(wittypes.Ok[wittypes.Unit, types.ErrorCode](wittypes.Unit{}))
	return rx
}

func main() {}
