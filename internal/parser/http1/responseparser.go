package http1

import (
	"bytes"
	"github.com/indigo-web/client/http"
	"github.com/indigo-web/client/http/protocol"
	"github.com/indigo-web/client/http/status"
	"github.com/indigo-web/client/internal/parser"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/strcomp"
	"github.com/indigo-web/utils/uf"
	"strconv"
	"strings"
)

var _ parser.Parser = &Parser{}

type Parser struct {
	state        parserState
	response     *http.Response
	respLineBuff buffer.Buffer[byte]
	headersBuff  buffer.Buffer[byte]
	encToksBuff  []string
	headerKey    string
}

func NewParser(resp *http.Response, respLineBuff, headersBuff buffer.Buffer[byte]) *Parser {
	return &Parser{
		state:        eProto,
		response:     resp,
		respLineBuff: respLineBuff,
		headersBuff:  headersBuff,
	}
}

func (p *Parser) Parse(data []byte) (headersCompleted bool, rest []byte, err error) {
	switch p.state {
	case eProto:
		goto proto
	case eCode:
		goto code
	case eStatus:
		goto status
	case eHeaderKey:
		goto headerKey
	case eHeaderKeyCR:
		goto headerKeyCR
	case eHeaderSemicolon:
		goto headerSemicolon
	case eHeaderValue:
		goto headerValue
	default:
		panic("BUG: response parser: unknown state")
	}

proto:
	{
		sp := bytes.IndexByte(data, ' ')
		if sp == -1 {
			if !p.respLineBuff.Append(data...) {
				return false, nil, status.ErrTooLongResponseLine
			}

			return false, nil, nil
		}

		// TODO: if we received the whole protocol all-at-once, we can avoid copying
		//  the data into the buffer and win a bit more of performance
		if !p.respLineBuff.Append(data[:sp]...) {
			return false, nil, status.ErrTooLongResponseLine
		}

		p.response.Proto = protocol.FromBytes(p.respLineBuff.Finish())
		if p.response.Proto == protocol.Unknown {
			return false, nil, status.ErrHTTPVersionNotSupported
		}

		data = data[sp+1:]
		p.state = eCode
		goto code
	}

code:
	for i := 0; i < len(data); i++ {
		if data[i] == ' ' {
			data = data[i+1:]
			p.state = eStatus
			goto status
		}

		if data[i] < '0' || data[i] > '9' {
			return false, nil, status.ErrBadRequest
		}

		p.response.Code = status.Code(int(p.response.Code)*10 + int(data[i]-'0'))
	}

	// note: as status.Code is uint16, and we're not checking overflow, it may
	// actually happen. Other question is, whether it's really anyhow dangerous

	return false, nil, nil

status:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.respLineBuff.Append(data...) {
				return false, nil, status.ErrTooLongResponseLine
			}

			return false, nil, nil
		}

		if !p.respLineBuff.Append(data[:lf]...) {
			return false, nil, status.ErrTooLongResponseLine
		}

		p.response.Status = status.Status(uf.B2S(rstripCR(p.respLineBuff.Finish())))
		data = data[lf+1:]
		p.state = eHeaderKey
		goto headerKey
	}

headerKey:
	if len(data) == 0 {
		return false, nil, nil
	}

	switch data[0] {
	case '\r':
		data = data[1:]
		p.state = eHeaderKeyCR
		goto headerKeyCR
	case '\n':
		return true, data[1:], nil
	}

	{
		semicolon := bytes.IndexByte(data, ':')
		if semicolon == -1 {
			if !p.headersBuff.Append(data...) {
				return false, nil, status.ErrHeaderKeyTooLarge
			}

			return false, nil, nil
		}

		if !p.headersBuff.Append(data[:semicolon]...) {
			return false, nil, status.ErrHeaderKeyTooLarge
		}

		p.headerKey = uf.B2S(p.headersBuff.Finish())
		data = data[semicolon+1:]
		p.state = eHeaderSemicolon
		goto headerSemicolon
	}

headerKeyCR:
	if data[0] != '\n' {
		return true, nil, status.ErrBadRequest
	}

	return true, data[1:], nil

headerSemicolon:
	for i := 0; i < len(data); i++ {
		if data[i] != ' ' {
			data = data[i:]
			p.state = eHeaderValue
			goto headerValue
		}
	}

	return false, nil, nil

headerValue:
	{
		lf := bytes.IndexByte(data, '\n')
		if lf == -1 {
			if !p.headersBuff.Append(data...) {
				return false, nil, status.ErrHeaderValueTooLarge
			}

			return false, nil, nil
		}

		if !p.headersBuff.Append(data[:lf]...) {
			return false, nil, status.ErrHeaderValueTooLarge
		}

		value := uf.B2S(rstripCR(p.headersBuff.Finish()))

		switch {
		case strcomp.EqualFold(p.headerKey, "content-length"):
			p.response.ContentLength, err = strconv.Atoi(value)
		case strcomp.EqualFold(p.headerKey, "content-type"):
			p.response.ContentType = value
		case strcomp.EqualFold(p.headerKey, "transfer-encoding"):
			toks, err := p.fillEncoding(value)
			if err != nil {
				return true, nil, err
			}

			p.response.Encoding.Transfer = append(p.response.Encoding.Transfer, toks...)
		case strcomp.EqualFold(p.headerKey, "content-encoding"):
			toks, err := p.fillEncoding(value)
			if err != nil {
				return true, nil, err
			}

			p.response.Encoding.Content = append(p.response.Encoding.Content, toks...)
		case strcomp.EqualFold(p.headerKey, "trailer"):
			p.response.Encoding.HasTrailer = true
			// TODO: implement upgrade header
		}

		p.response.Headers.Add(p.headerKey, value)
		data = data[lf+1:]
		p.state = eHeaderKey
		goto headerKey
	}
}

func (p *Parser) Release() {
	p.state = eProto
	p.respLineBuff.Clear()
	p.headersBuff.Clear()
}

func (p *Parser) fillEncoding(value string) (toks []string, err error) {
	tokens, chunked, err := parseEncodingString(p.encToksBuff[:0], value)
	p.response.Encoding.Chunked = p.response.Encoding.Chunked || chunked

	return tokens, err
}

func parseEncodingString(buff []string, value string) (toks []string, chunked bool, err error) {
	var offset int

	for i := range value {
		if value[i] == ',' {
			var isChunked bool
			buff, isChunked, err = processEncodingToken(buff, value[offset:i])
			if err != nil {
				return nil, false, err
			}

			chunked = chunked || isChunked
			offset = i + 1
		}
	}

	var isChunked bool
	buff, isChunked, err = processEncodingToken(buff, value[offset:])

	return buff, chunked || isChunked, err
}

func processEncodingToken(
	buff []string, rawToken string,
) ([]string, bool, error) {
	switch token := strings.TrimSpace(rawToken); token {
	case "":
	case "chunked":
		return buff, true, nil
	default:
		if len(buff)+1 >= cap(buff) {
			return nil, false, status.ErrUnsupportedEncoding
		}

		buff = append(buff, token)
	}

	return buff, false, nil
}

func rstripCR(b []byte) []byte {
	if b[len(b)-1] == '\r' {
		b = b[:len(b)-1]
	}

	return b
}
