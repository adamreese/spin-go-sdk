package export_wasi_http_0_3_0_rc_2026_03_15_handler

import (
	"github.com/spinframework/spin-go-sdk/v3/imports/wasi_http_0_2_0_types"
)

var Exports struct {
	Handle func(request *wasi_http_0_2_0_types.IncomingRequest, responseOut *wasi_http_0_2_0_types.ResponseOutparam)
}

func Handle(request *wasi_http_0_2_0_types.IncomingRequest, responseOut *wasi_http_0_2_0_types.ResponseOutparam) {
	Exports.Handle(request, responseOut)
}
