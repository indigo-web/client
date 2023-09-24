package http

type Encoding struct {
	// Transfer represents Transfer-Encoding header value, split by comma
	Transfer []string
	// Content represents Content-Encoding header value, split by comma
	Content []string
	// Chunked doesn't belong to any of encodings, as it is still must be processed individually
	Chunked, HasTrailer bool
}

func (e Encoding) Clear() Encoding {
	e.Transfer = e.Transfer[:0]
	e.Content = e.Content[:0]
	e.Chunked = false
	e.HasTrailer = false
	return e
}
