package gaphttp

import (
	"bytes"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Middleware func(next http.Handler) http.Handler

type MiddlewareOptions struct {
	ServiceName    string
	EnabledLogger  bool
	EnabledTracer  bool
	EnabledRecover bool
	EnabledMetrics bool
}

const (
	builderNameMiddlewareTrace    = "trace.middleware"
	builderNameMiddlewareLogger   = "logger.middleware"
	builderNameMiddlewareRecovery = "recovery.middleware"
	builderNameMiddlewareMetrics  = "metrics.middleware"
)

type MiddlewareBuilder struct {
	opt         *MiddlewareOptions
	logger      *zap.Logger
	tracer      trace.Tracer
	metrics     Metrics
	middlewares map[string]func() func(next http.Handler) http.Handler
}

func NewMiddlewareBuilder(
	opt *MiddlewareOptions,
	logger *zap.Logger,
	tracer trace.Tracer,
	metrics Metrics,
) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		opt:         opt,
		logger:      logger,
		tracer:      tracer,
		metrics:     metrics,
		middlewares: make(map[string]func() func(next http.Handler) http.Handler),
	}

}

func (b *MiddlewareBuilder) AddTrace() *MiddlewareBuilder {
	b.middlewares[builderNameMiddlewareTrace] = func() func(next http.Handler) http.Handler {
		return NewMiddlewareTracer(b.tracer, b.opt)
	}

	return b
}

func (b *MiddlewareBuilder) AddLogger() *MiddlewareBuilder {
	b.middlewares[builderNameMiddlewareLogger] = func() func(next http.Handler) http.Handler {
		return NewMiddlewareLogger(b.logger, b.opt)
	}

	return b
}

func (b *MiddlewareBuilder) AddMiddlewareRecover() *MiddlewareBuilder {
	b.middlewares[builderNameMiddlewareRecovery] = func() func(next http.Handler) http.Handler {
		return NewMiddlewareRecovery(b.logger, b.opt)
	}

	return b
}

func (b *MiddlewareBuilder) AddMiddlewareMeter() *MiddlewareBuilder {
	b.middlewares[builderNameMiddlewareMetrics] = func() func(next http.Handler) http.Handler {
		return NewMiddlewareMetrics(b.metrics, b.opt)
	}

	return b
}

// Build - build middleware and create correct position in stack
func (b *MiddlewareBuilder) Build() ([]func(next http.Handler) http.Handler, error) {
	if b.opt == nil {
		return nil, errors.New("options cannot be blank")
	}

	out := make([]func(next http.Handler) http.Handler, 0, len(b.middlewares))

	recMidForMid, ok := b.middlewares[builderNameMiddlewareRecovery]
	if ok {
		out = append(out, recMidForMid())
	}

	metrMid, ok := b.middlewares[builderNameMiddlewareMetrics]
	if ok {
		out = append(out, metrMid())
	}

	traceMid, ok := b.middlewares[builderNameMiddlewareTrace]
	if ok {
		out = append(out, traceMid())
	}

	logMid, ok := b.middlewares[builderNameMiddlewareLogger]
	if ok {
		out = append(out, logMid())
	}

	recMidSrv, ok := b.middlewares[builderNameMiddlewareRecovery]
	if ok {
		out = append(out, recMidSrv())
	}

	return out, nil
}

// WrapWriter supported butch loading
type WrapWriter struct {
	status int
	body   *bytes.Buffer
	inner  http.ResponseWriter
}

func NewWrapWriter(inner http.ResponseWriter) *WrapWriter {
	return &WrapWriter{body: bytes.NewBuffer([]byte{}), inner: inner}
}

func (w *WrapWriter) Header() http.Header {
	return w.inner.Header()
}

func (w *WrapWriter) Write(i []byte) (int, error) {
	w.body.Write(i)

	return w.inner.Write(i)
}

func (w *WrapWriter) WriteHeader(statusCode int) {
	w.status = statusCode

	w.inner.WriteHeader(statusCode)
}

func (w *WrapWriter) Status() int {
	return w.status
}

func (w *WrapWriter) Body() []byte {
	return w.body.Bytes()
}
