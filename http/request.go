package http

import (
	"github.com/indigo-web/client/http/headers"
	"github.com/indigo-web/client/http/method"
	"github.com/indigo-web/client/http/protocol"
	"github.com/indigo-web/utils/uf"
	"io"
	"os"
)

type Request struct {
	Method  method.Method
	Path    string
	Proto   protocol.Protocol
	Headers *headers.Headers
	File    *os.File
	Body    []byte
	err     error
}

func NewRequest(hdrs *headers.Headers) *Request {
	return &Request{
		Proto:   protocol.Auto,
		Headers: hdrs,
	}
}

func (r *Request) WithMethod(m method.Method) *Request {
	r.Method = m
	return r
}

func (r *Request) WithPath(path string) *Request {
	r.Path = path
	return r
}

func (r *Request) WithProtocol(proto protocol.Protocol) *Request {
	r.Proto = proto
	return r
}

func (r *Request) WithHeader(key string, values ...string) *Request {
	r.Headers.Add(key, values...)
	return r
}

// WithFile opens a new file with os.O_RDONLY flag and perm=0
func (r *Request) WithFile(filename string) *Request {
	r.File, r.err = os.OpenFile(filename, os.O_RDONLY, 0)
	return r
}

func (r *Request) WithBody(body string) *Request {
	return r.WithBodyBytes(uf.S2B(body))
}

func (r *Request) WithBodyBytes(body []byte) *Request {
	r.Body = body
	return r
}

func (r *Request) WithBodyFrom(reader io.Reader) *Request {
	r.Body, r.err = io.ReadAll(reader)
	return r
}

// Error returns error, if occurred during request building. This may be caused
// by non-existing filename, passed via File, or BodyFrom, if error occurred during
// reading from it
func (r *Request) Error() error {
	return r.err
}

func (r *Request) WithClear() *Request {
	r.Method = method.Unknown
	r.Path = ""
	r.Proto = protocol.Auto
	r.Headers.Clear()
	r.File = nil
	r.Body = nil
	r.err = nil
	return r
}
