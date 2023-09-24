package http1

import (
	"fmt"
	"github.com/indigo-web/client/http"
	"github.com/indigo-web/client/http/method"
	"github.com/indigo-web/client/http/protocol"
	"github.com/indigo-web/client/internal/tcp"
	"io"
	"os"
	"strconv"
)

type Renderer struct {
	client tcp.Client
	buff   []byte
}

func NewRenderer(client tcp.Client, buff []byte) *Renderer {
	return &Renderer{
		client: client,
		buff:   buff,
	}
}

func (r *Renderer) Send(request *http.Request) error {
	r.method(request.Method)
	r.sp()
	r.path(request.Path)
	r.sp()
	r.proto(request.Proto)
	r.crlf()

	for headersIter := request.Headers.Iter(); ; {
		pair, cont := headersIter.Next()
		if !cont {
			break
		}

		r.header(pair.Key, pair.Value)
		r.crlf()
	}

	r.crlf()

	if request.File != nil {
		return r.file(request.File)
	}

	r.buff = append(r.buff, request.Body...)

	fmt.Println("sending:", strconv.Quote(string(r.buff)))

	return r.client.Write(r.buff)
}

func (r *Renderer) file(fd *os.File) error {
	// TODO: implement chunked streaming for files with size>N, where N tends to be more than 1mb
	content, err := io.ReadAll(fd)
	if err != nil {
		return err
	}

	r.buff = append(r.buff, content...)

	return r.client.Write(r.buff)
}

func (r *Renderer) method(m method.Method) {
	r.buff = append(r.buff, m...)
}

func (r *Renderer) sp() {
	r.buff = append(r.buff, ' ')
}

func (r *Renderer) path(path string) {
	r.buff = append(r.buff, path...)
}

func (r *Renderer) proto(proto protocol.Protocol) {
	r.buff = append(r.buff, proto...)
}

func (r *Renderer) crlf() {
	r.buff = append(r.buff, '\r', '\n')
}

func (r *Renderer) header(key, value string) {
	r.buff = append(r.buff, key...)
	r.buff = append(r.buff, ':', ' ')
	r.buff = append(r.buff, value...)
}
