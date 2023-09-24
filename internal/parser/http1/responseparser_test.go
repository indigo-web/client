package http1

import (
	"github.com/indigo-web/client/http"
	"github.com/indigo-web/client/http/headers"
	"github.com/indigo-web/client/http/protocol"
	"github.com/indigo-web/client/http/status"
	"github.com/indigo-web/utils/buffer"
	"github.com/stretchr/testify/require"
	"testing"
)

func compareResponse(t *testing.T, want, got *http.Response) {
	require.Equal(t, want.Proto, got.Proto)
	require.Equal(t, int(want.Code), int(got.Code))
	if len(want.Status) > 0 {
		require.Equal(t, want.Status, got.Status)
	}

	for _, key := range want.Headers.Keys() {
		require.True(t, got.Headers.Has(key))
		require.Equal(t, want.Headers.Values(key), got.Headers.Values(key))
	}
}

func TestResponseParser(t *testing.T) {
	resp := http.NewResponse()
	parser := NewParser(
		resp, *buffer.NewBuffer[byte](0, 4096), *buffer.NewBuffer[byte](0, 4096),
	)

	t.Run("simple response", func(t *testing.T) {
		defer parser.Release()
		defer resp.Clear()

		data := "HTTP/1.1 200 OK\r\n\r\n"
		headersCompleted, rest, err := parser.Parse([]byte(data))
		require.NoError(t, err)
		require.True(t, headersCompleted)
		require.Empty(t, rest)
		compareResponse(t, &http.Response{
			Proto:   protocol.HTTP11,
			Code:    status.OK,
			Status:  "OK",
			Headers: headers.NewHeaders(),
		}, resp)
	})

	t.Run("response with headers", func(t *testing.T) {
		defer parser.Release()
		defer resp.Clear()

		data := "HTTP/1.1 200 OK\r\nHello: world\r\nhello: nether\r\n\r\n"
		headersCompleted, rest, err := parser.Parse([]byte(data))
		require.NoError(t, err)
		require.True(t, headersCompleted)
		require.Empty(t, rest)
		compareResponse(t, &http.Response{
			Proto:  protocol.HTTP11,
			Code:   status.OK,
			Status: "OK",
			Headers: headers.FromMap(map[string][]string{
				"hello": {"world", "nether"},
			}),
		}, resp)
	})
}
