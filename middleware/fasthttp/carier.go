package fasthttp

import (
	"github.com/valyala/fasthttp"
)

// HeaderCarrier adapts http.Header to satisfy the TextMapCarrier interface.
type HeaderCarrier struct {
	*fasthttp.RequestHeader
}

// NewCarrier implemented TextMapCarrier
func NewCarrier(in *fasthttp.RequestHeader) *HeaderCarrier {
	return &HeaderCarrier{RequestHeader: in}
}

// Get returns the value associated with the passed key.
func (hc HeaderCarrier) Get(key string) string {
	return string(hc.Peek(key))
}

// Set stores the key-value pair.
func (hc HeaderCarrier) Set(key string, value string) {
	hc.RequestHeader.Set(key, value)
}

// Keys lists the keys stored in this carrier.
func (hc HeaderCarrier) Keys() []string {
	keys := make([]string, 0)

	hc.VisitAll(func(key, value []byte) {
		keys = append(keys, string(key))
	})

	return keys
}
