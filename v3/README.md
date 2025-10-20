WIP

spin-go-sdk with wasip2 support

Notes:

The current version of tooling used for this work:

- wit-bindgen-go `wit-bindgen-go version v0.7.1-0.20250704174425-55a8715c694f (55a8715c694fb4bbe75c269ffd4a0ca14c80c6f1)`
- wasm-tools `wasm-tools 1.240.0`
- tinygo `tinygo version 0.39.0 darwin/arm64 (using go version go1.25.3 and LLVM version 19.1.2)`
- spin `spin 3.5.0-pre0 (44e1bef7 2025-09-19)`
- go `go version go1.25.3 darwin/arm64`
- binaryen tools `binaryen-version_123`


Regeneratin bindings:

- install tooling as specified above
- make sure they are on PATH and picking up the versions as specified above
- cd `<root>/v3`
- Run: `wit-bindgen-go generate -w spin-http -p github.com/spinframework/spin-go-sdk/v3/internal --out internal ./wit`

Testing:

- cd `<root>/v3/examples/http`
- Run `spin build`
- Run `spin up`
- In a separate terminal, run: `curl http://127.0.0.1:3000/hello`
