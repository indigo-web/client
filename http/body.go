package http

import (
	"github.com/indigo-web/utils/unreader"
	"io"
)

type (
	onBodyCallback func([]byte) error
	BodyReader     interface {
		Init(response *Response)
		Read() ([]byte, error)
	}
)

var _ io.Reader = &Body{}

type Body struct {
	reader        BodyReader
	unreader      unreader.Unreader
	encoding      Encoding
	contentLength int
	bodyBuff      []byte
}

func NewBody(reader BodyReader) *Body {
	return &Body{
		reader: reader,
	}
}

const chunkedTE = -1

// Init is a system method, that MUST not be called, otherwise connection may get
// stuck, leading to hanging connection (until read-timeout won't be exceeded)
func (b *Body) Init(resp *Response) {
	b.reader.Init(resp)
	b.unreader.Reset()
	b.encoding = resp.Encoding
	b.contentLength = resp.ContentLength

	if resp.Encoding.Chunked {
		b.contentLength = chunkedTE
	}
}

// Full returns the whole body at once
//
// WARNING: returned slice is an underlying buffer, that will be re-written during the
// next call of this method.
func (b *Body) Full() ([]byte, error) {
	// in case transfer-encoding is chunked, this condition won't be satisfied. In this case,
	// the buffer may still be nil, that'll cause many re-allocations. The problem is, that
	// chunked body size median in production is unknown (and usually varies across projects).
	// Maybe, we could add an option to the settings to pre-allocate the buffer in such cases,
	// but this will affect only cold-start stages. Not sure, whether it's time-worthy
	if len(b.bodyBuff) < b.contentLength {
		b.bodyBuff = make([]byte, 0, b.contentLength)
	}

	b.bodyBuff = b.bodyBuff[:0]

	return b.bodyBuff, b.callback(func(data []byte) error {
		b.bodyBuff = append(b.bodyBuff, data...)
		return nil
	})
}

// Read implements the io.Reader interface, so behaves respectively
func (b *Body) Read(into []byte) (n int, err error) {
	data, err := b.unreader.PendingOr(b.reader.Read)
	if err != nil {
		return 0, err
	}

	n = copy(into, data)
	if len(data) > n {
		b.unreader.Unread(data[n:])
	}

	return n, err
}

// Callback takes a function, that'll be called with body piece every time it's received.
// In case error is returned from the callback, it'll also be returned from this method
func (b *Body) Callback(onBody onBodyCallback) error {
	return b.callback(onBody)
}

// Reset resets the body.
//
// NOTE: this is a system method, that SHOULD NOT be called by user manually. However,
// this won't affect anything anyhow, except impossibility to restore the body data
func (b *Body) Reset() error {
	for {
		_, err := b.reader.Read()
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}
	}
}

func (b *Body) callback(onBody onBodyCallback) error {
	for {
		piece, err := b.reader.Read()
		switch err {
		case nil:
		case io.EOF:
			return nil
		default:
			return err
		}

		if err = onBody(piece); err != nil {
			return err
		}
	}
}
