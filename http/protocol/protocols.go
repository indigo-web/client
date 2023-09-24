package protocol

import (
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
)

type Protocol = string

const (
	Unknown Protocol = ""
	HTTP09  Protocol = "HTTP/0.9"
	HTTP10  Protocol = "HTTP/1.0"
	HTTP11  Protocol = "HTTP/1.1"

	// Auto is the newest available protocol
	Auto
)

func FromBytes(b []byte) Protocol {
	switch proto := uf.B2S(b); {
	case strcomp.EqualFold(proto, HTTP11):
		return HTTP11
	case strcomp.EqualFold(proto, HTTP10):
		return HTTP10
	case strcomp.EqualFold(proto, HTTP09):
		return HTTP09
	}

	return Unknown
}
