package httpp3

import (
	"io"

	wit "go.bytecodealliance.org/pkg/wit/types"
)

type streamReader struct {
	stream *wit.StreamReader[uint8]
}

func (r *streamReader) Close() error {
	r.stream.Drop()
	return nil
}

func (r *streamReader) Read(p []byte) (n int, err error) {
	count := r.stream.Read(p)
	if count == 0 {
		if r.stream.WriterDropped() {
			return 0, io.EOF
		}
		return 0, nil
	}
	return int(count), nil
}

// NewReader creates an io.Reader from a WASI stream
func NewReader(s *wit.StreamReader[uint8]) io.Reader {
	return &streamReader{
		stream: s,
	}
}

// NewReadCloser creates an io.ReadCloser from a WASI stream
func NewReadCloser(s *wit.StreamReader[uint8]) io.ReadCloser {
	return &streamReader{
		stream: s,
	}
}
