package coding

import (
	"errors"
)

var (
	ErrUnknownToken = errors.New("coding token is not recognized")
)

type Token = string

type Encoder interface {
	// Encode TODO: do we really need to return an error here?
	Encode(input []byte) (output []byte, err error)
}

type Decoder interface {
	Decode(input []byte) (output []byte, err error)
}

type Coding interface {
	Encoder
	Decoder
}

type Manager struct {
	encoders map[Token]Encoder
	decoders map[Token]Decoder
}

func NewManager() Manager {
	return Manager{
		encoders: make(map[Token]Encoder),
		decoders: make(map[Token]Decoder),
	}
}

func (m Manager) AddEncoder(token Token, encoder Encoder) {
	addCoding(token, encoder, m.encoders)
}

func (m Manager) AddDecoder(token Token, decoder Decoder) {
	addCoding(token, decoder, m.decoders)
}

func (m Manager) Encode(token Token, input []byte) (output []byte, err error) {
	encoder, found := m.encoders[token]
	if !found {
		return nil, ErrUnknownToken
	}

	return encoder.Encode(input)
}

func (m Manager) Decode(token Token, input []byte) (output []byte, err error) {
	decoder, found := m.decoders[token]
	if !found {
		return nil, ErrUnknownToken
	}

	return decoder.Decode(input)
}

func addCoding[V any](token Token, value V, into map[Token]V) {
	into[token] = value

	// this exists in backward-capability purposes. Some old clients may use x-gzip or
	// x-compress instead of regular gzip or compress tokens respectively.
	// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding#directives
	switch token {
	case "gzip":
		into["x-gzip"] = value
	case "compress":
		into["x-compress"] = value
	}
}
