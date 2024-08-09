package gaphttp

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const RequestHeaderTraceID = "x-trace-id"

type Header interface {
	Set(key string, value string)
	Get(key string) string
}

func ExtractRemoteTraceID(header Header) (trace.TraceID, bool) {
	traceIdAsString := header.Get(RequestHeaderTraceID)
	if traceIdAsString == "" {
		return trace.TraceID{}, false
	}

	traceId, err := trace.TraceIDFromHex(traceIdAsString)
	if err != nil {
		return trace.TraceID{}, false
	}

	return traceId, true
}

func NewMiddlewareTracer(tracer trace.Tracer, option *MiddlewareOptions) Middleware {
	// tracer disabled
	if !option.EnabledTracer {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(writer http.ResponseWriter, request *http.Request) {
					next.ServeHTTP(writer, request)
				},
			)
		}
	}

	// tracer enabled
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				reqCtx := request.Context()
				rTraceID, isRemote := ExtractRemoteTraceID(request.Header)
				if isRemote {
					remoteSpanCtx := trace.NewSpanContext(trace.SpanContextConfig{
						TraceID:    rTraceID,
						SpanID:     trace.SpanID{},
						TraceFlags: 0,
						TraceState: trace.TraceState{},
						Remote:     isRemote,
					})

					reqCtx = trace.ContextWithRemoteSpanContext(request.Context(), remoteSpanCtx)
				}

				path := RemoveChiPathParam(request)
				ctxWithSpan, span := tracer.Start(reqCtx, makeSpanName(path))

				span.SetAttributes(
					attribute.Bool("trace.id.remoted", isRemote),
					attribute.String("http.method", request.Method),
					attribute.String("http.user_agent", request.UserAgent()),
					attribute.String("http.client_ip", request.RemoteAddr),
					attribute.Int64("http.request_content_length", request.ContentLength),
					attribute.String("http.target", request.URL.Path),
				)

				request = request.WithContext(ctxWithSpan)
				wrappedWriter := NewWrapWriter(writer)

				next.ServeHTTP(wrappedWriter, request)

				status := codes.Ok
				if wrappedWriter.status != http.StatusOK {
					status = codes.Error
				}

				span.SetStatus(status, fmt.Sprintf("http.status: %d", wrappedWriter.Status()))
				span.End()
			},
		)
	}
}

func makeSpanName(path string) string {
	return fmt.Sprintf("http.handler.%s", path)
}
