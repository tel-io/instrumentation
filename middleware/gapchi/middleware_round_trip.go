package gaphttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type MiddlewareRoundTrip struct {
	inner http.RoundTripper
	debug bool
	log   *zap.Logger
}

func NewMiddlewareRoundTrip(
	inner http.RoundTripper,
	debug bool,
	log *zap.Logger,
) *MiddlewareRoundTrip {
	return &MiddlewareRoundTrip{
		inner: inner,
		debug: debug,
		log:   log,
	}
}

func (rt *MiddlewareRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	span := trace.SpanFromContext(req.Context())
	span.AddEvent(
		fmt.Sprintf("http.client.request.started: %s", req.URL.Path),
		trace.WithAttributes(attribute.String("http.client.method", req.Method)),
	)

	req.Header.Add(RequestHeaderTraceID, span.SpanContext().TraceID().String())

	rt.LogRequest(span, req)

	res, err := rt.inner.RoundTrip(req)
	if err != nil {
		rt.log.Error("round trip", zap.Error(err))

		return res, err
	}

	span.AddEvent(
		fmt.Sprintf("http.client.request.ended: %s", req.URL.Path),
		trace.WithAttributes(attribute.Int("http.client.status", res.StatusCode)),
	)

	rt.LogResponse(span, res)

	return res, err
}

func (rt *MiddlewareRoundTrip) LogResponse(span trace.Span, response *http.Response) {
	if !rt.debug {
		return
	}

	log := rt.log.With(
		zap.String("traceID", span.SpanContext().TraceID().String()),
		zap.String("spanID", span.SpanContext().SpanID().String()),
		zap.Int("http.client.response.status", response.StatusCode),
	)

	if response.Body != nil {
		respBody, err := io.ReadAll(response.Body)
		if err != nil {
			log.Error("http.client.response.body.read", zap.Error(err))
		}

		log = log.With(zap.ByteString("http.client.response.body", respBody))

		response.Body = io.NopCloser(bytes.NewBuffer(respBody)) // Reset
	}

	log.Debug("http.client.request.stopped!")
}

func (rt *MiddlewareRoundTrip) LogRequest(span trace.Span, request *http.Request) {
	if !rt.debug {
		return
	}

	log := rt.log.With(
		zap.String("traceID", span.SpanContext().TraceID().String()),
		zap.String("spanID", span.SpanContext().SpanID().String()),
		zap.String("http.client.request.method", request.Method),
		zap.Any("http.client.request.header", request.Header),
		zap.String("http.client.request.url", request.URL.String()),
	)

	if request.Body != nil {
		reqBody, err := io.ReadAll(request.Body)
		if err != nil {
			log.Error("http.client.request.body.read", zap.Error(err))
		}

		log = log.With(zap.ByteString("http.request.body", reqBody))

		request.Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset
	}

	log.Debug("http.client.request.started!")
}
