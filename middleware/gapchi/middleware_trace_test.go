package gaphttp

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestMiddlewareTrace(t *testing.T) {
	tc := testCase{
		getReq: func(url string) *http.Request {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s=%s", url, "name", "test"), nil)
			require.Nil(t, err)

			req = req.WithContext(trace.ContextWithRemoteSpanContext(
				req.Context(), trace.NewSpanContext(
					trace.SpanContextConfig{
						TraceID:    trace.TraceID{0x1},
						SpanID:     trace.SpanID{},
						TraceFlags: 0,
						TraceState: trace.TraceState{},
						Remote:     false,
					},
				),
			))

			return req
		},
		postReq: func(url string) *http.Request {
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(`{"id":"11"}`)))
			require.Nil(t, err)

			return req
		},

		get: func(writer http.ResponseWriter, request *http.Request) {
			fmt.Println(
				trace.SpanFromContext(request.Context()).SpanContext().TraceID().String(),
				trace.SpanFromContext(request.Context()).SpanContext().SpanID().String(),
			)
		},
		post: func(writer http.ResponseWriter, request *http.Request) {
			fmt.Println(
				trace.SpanFromContext(request.Context()).SpanContext().TraceID().String(),
				trace.SpanFromContext(request.Context()).SpanContext().SpanID().String(),
			)
		},
	}

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	router.Use(NewMiddlewareTracer(tracer, &MiddlewareOptions{EnabledTracer: true}))
	router.Get(testURL, tc.get)
	router.Post(testURL, tc.post)

	server := httptest.NewServer(router)

	reqs := []*http.Request{
		tc.getReq(fmt.Sprintf("%s%s", server.URL, testURL)),
		tc.postReq(fmt.Sprintf("%s%s", server.URL, testURL)),
	}

	// Attention!   NewMiddlewareRoundTrip
	cli := http.Client{Transport: NewMiddlewareRoundTrip(http.DefaultTransport, true, zap.NewExample())}
	for _, req := range reqs {
		_, err := cli.Do(req)
		require.Nil(t, err)
	}

}
