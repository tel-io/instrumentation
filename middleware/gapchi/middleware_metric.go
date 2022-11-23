package gaphttp

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const (
	RequestCount          = "http.server.request_count"           // Incoming request count total
	RequestContentLength  = "http.server.request_content_length"  // Incoming request bytes total
	ResponseContentLength = "http.server.response_content_length" // Incoming response bytes total
	ServerLatency         = "http.server.duration"                // Incoming end to end duration, microseconds
)

type Metrics interface {
	IncRequestCount(ctx context.Context, attrs ...attribute.KeyValue)
	AddRequestContentLength(ctx context.Context, len int64, attrs ...attribute.KeyValue)
	AddResponseContentLength(ctx context.Context, len int64, attrs ...attribute.KeyValue)
	AddHandleDuration(ctx context.Context, dur float64, attrs ...attribute.KeyValue)
}

func NewMiddlewareMetrics(metrics Metrics, option *MiddlewareOptions) Middleware {
	// metrics disabled
	if !option.EnabledMetrics {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(writer http.ResponseWriter, request *http.Request) {
					next.ServeHTTP(writer, request)
				},
			)
		}
	}

	// metrics enabled
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				requestStartTime := time.Now().UTC()

				wrappedWriter := NewWrapWriter(writer)

				next.ServeHTTP(wrappedWriter, request)

				labeler := new(Labeler)
				labeler.Add(attribute.String("http.url.path", RemoveChiPathParam(request)))
				labeler.Add(attribute.Int("http.status", wrappedWriter.Status()))

				attributes := append(
					labeler.Get(),
					semconv.HTTPServerMetricAttributesFromHTTPRequest(option.ServiceName, request)...,
				)

				elapsedTime := float64(time.Since(requestStartTime)) / float64(time.Millisecond)

				metrics.AddHandleDuration(request.Context(), elapsedTime, attributes...)
				metrics.IncRequestCount(request.Context(), attributes...)
				metrics.AddRequestContentLength(request.Context(), request.ContentLength, attributes...)
				metrics.AddResponseContentLength(request.Context(), int64(len(wrappedWriter.Body())), attributes...)
			},
		)
	}
}

func RegisterBasicMetrics(meter metric.Meter) (*BasicMetrics, error) {
	rc, err := meter.SyncInt64().Counter(RequestCount)
	if err != nil {
		return nil, fmt.Errorf("meter, syncInt64, counter: %w", err)
	}

	reqCl, err := meter.SyncInt64().Counter(RequestContentLength)
	if err != nil {
		return nil, fmt.Errorf("meter, syncInt64, counter: %w", err)
	}

	resCl, err := meter.SyncInt64().Counter(ResponseContentLength)
	if err != nil {
		return nil, fmt.Errorf("meter, syncInt64, counter: %w", err)
	}

	sl, err := meter.SyncFloat64().Histogram(ServerLatency)
	if err != nil {
		return nil, fmt.Errorf("meter, syncFloat64, histogram: %w", err)
	}

	return &BasicMetrics{
		requestCount:          rc,
		requestContentLength:  reqCl,
		responseContentLength: resCl,
		handleDuration:        sl,
	}, nil
}

type BasicMetrics struct {
	requestCount          syncint64.Counter
	requestContentLength  syncint64.Counter
	responseContentLength syncint64.Counter
	handleDuration        syncfloat64.Histogram
}

func (m *BasicMetrics) IncRequestCount(ctx context.Context, attrs ...attribute.KeyValue) {
	m.requestCount.Add(ctx, 1, attrs...)
}

func (m *BasicMetrics) AddRequestContentLength(ctx context.Context, len int64, attrs ...attribute.KeyValue) {
	m.requestContentLength.Add(ctx, len, attrs...)
}

func (m *BasicMetrics) AddResponseContentLength(ctx context.Context, len int64, attrs ...attribute.KeyValue) {
	m.responseContentLength.Add(ctx, len, attrs...)
}

func (m *BasicMetrics) AddHandleDuration(ctx context.Context, dur float64, attrs ...attribute.KeyValue) {
	m.handleDuration.Record(ctx, dur, attrs...)
}

// Labeler is used to allow instrumented HTTP handlers to add custom attributes to
// the metrics recorded by the net/http instrumentation.
type Labeler struct {
	mu         sync.Mutex
	attributes []attribute.KeyValue
}

// Add attributes to a Labeler.
func (l *Labeler) Add(ls ...attribute.KeyValue) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.attributes = append(l.attributes, ls...)
}

// Get returns a copy of the attributes added to the Labeler.
func (l *Labeler) Get() []attribute.KeyValue {
	l.mu.Lock()
	defer l.mu.Unlock()
	ret := make([]attribute.KeyValue, len(l.attributes))
	copy(ret, l.attributes)
	return ret
}
