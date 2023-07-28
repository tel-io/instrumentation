package gaphttp

import (
	"bytes"
	"io"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func NewMiddlewareLogger(logger *zap.Logger, option *MiddlewareOptions) Middleware {
	// logger enabled
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				// set trace id and span id to logger
				span := trace.SpanFromContext(request.Context())

				log := logger.With(
					zap.String("traceID", span.SpanContext().TraceID().String()),
					zap.String("spanID", span.SpanContext().SpanID().String()),
				)

				log = log.With(
					zap.String("http.request.method", request.Method),
					zap.Any("http.request.header", request.Header),
					zap.Any("http.request.ip", request.RemoteAddr),
					zap.Any("http.request.path", request.URL.Path),
					zap.Any("http.request.query", request.URL.Query()),
					zap.Any("http.request.cookie", request.Cookies()),
				)

				// set request body to logger if exist
				if request.Body != nil {
					reqBody, err := io.ReadAll(request.Body)
					if err != nil {
						logger.Error("http.request.body.read", zap.Error(err))
					}

					log = log.With(zap.ByteString("http.request.body", reqBody))

					request.Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset
				}

				wrapWriter := NewWrapWriter(writer)

				log.Debug("handle started!")

				next.ServeHTTP(wrapWriter, request)

				if wrapWriter.status == http.StatusOK {
					log.Debug(
						"handle stopped!",
						zap.Int("http.response.status", wrapWriter.Status()),
						zap.ByteString("http.response.body", wrapWriter.Body()),
					)
				} else {
					log.Error("handle stopped!",
						zap.Int("http.response.status", wrapWriter.Status()),
						zap.ByteString("http.response.body", wrapWriter.Body()),
					)
				}
			},
		)
	}
}
