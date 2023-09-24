package http1

import (
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/client/http"
	"github.com/indigo-web/client/internal/tcp"
	"io"
)

type bodyBytesLeft = int

const chunked bodyBytesLeft = -1

type Body struct {
	client        tcp.Client
	encoding      http.Encoding
	bytesLeft     bodyBytesLeft
	chunkedParser *chunkedbody.Parser
}

func NewBody(client tcp.Client, parser *chunkedbody.Parser) *Body {
	return &Body{
		client:        client,
		chunkedParser: parser,
	}
}

func (b *Body) Init(response *http.Response) {
	b.encoding = response.Encoding
	b.bytesLeft = response.ContentLength
}

func (b *Body) Read() ([]byte, error) {
	// TODO: implement decoding
	if b.bytesLeft == 0 {
		return nil, io.EOF
	}

	data, err := b.client.Read()
	if err != nil {
		return nil, err
	}

	if b.bytesLeft == chunked {
		return b.readChunked(data)
	}

	return b.readPlain(data)
}

func (b *Body) readChunked(data []byte) ([]byte, error) {
	chunk, extra, err := b.chunkedParser.Parse(data, b.encoding.HasTrailer)
	switch err {
	case nil:
	case io.EOF:
		b.bytesLeft = 0
	default:
		return nil, err
	}

	b.client.Unread(extra)

	return chunk, nil
}

func (b *Body) readPlain(data []byte) ([]byte, error) {
	if len(data) <= b.bytesLeft {
		b.bytesLeft -= len(data)

		return data, nil
	}

	body, rest := data[:b.bytesLeft], data[b.bytesLeft:]
	b.client.Unread(rest)
	b.bytesLeft = 0

	return body, nil
}
