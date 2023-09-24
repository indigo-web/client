package render

import (
	"fmt"
	"github.com/indigo-web/client/http"
	"github.com/indigo-web/client/http/protocol"
	"github.com/indigo-web/client/internal/render/http1"
	"github.com/indigo-web/client/internal/tcp"
)

type Renderer struct {
	http1 *http1.Renderer
}

func NewRenderer(client tcp.Client, buff []byte) Renderer {
	return Renderer{
		http1: http1.NewRenderer(client, buff),
	}
}

func (r Renderer) Send(request *http.Request) error {
	switch request.Proto {
	case protocol.HTTP09, protocol.HTTP10, protocol.HTTP11:
		return r.http1.Send(request)
	}

	return fmt.Errorf("unsupported protocol: %s", request.Proto)
}
