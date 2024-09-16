package gaphttp

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type HandlerBasic struct {
	ServiceName string
	Logger      *zap.Logger
}

func NewHandlerBasic(svcName string, log *zap.Logger) *HandlerBasic {
	return &HandlerBasic{
		ServiceName: svcName,
		Logger:      log,
	}
}

func (h *HandlerBasic) OK(writer http.ResponseWriter, request *http.Request) {
	h.Write(request.Context(), writer, http.StatusOK, h.ServiceName)
}

func (h *HandlerBasic) NotFound(writer http.ResponseWriter, request *http.Request) {
	h.Write(request.Context(), writer, http.StatusNotFound, "invalid path")
}

func (h *HandlerBasic) Write(ctx context.Context, writer http.ResponseWriter, statusCode int, response interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	if out, ok := response.([]byte); ok {
		if _, err := writer.Write(out); err != nil {
			h.Log(ctx).Error("write response", zap.Error(err), zap.ByteString("response", out))
		}

		return
	}

	if err := json.NewEncoder(writer).Encode(response); err != nil {
		h.Log(ctx).Error("json encoder, write response", zap.Error(err), zap.Any("response", response))
	}
}

func (h *HandlerBasic) Log(ctx context.Context) *zap.Logger {
	spCtx := trace.SpanContextFromContext(ctx)

	return h.Logger.With(
		zap.String("traceID", spCtx.TraceID().String()),
		zap.String("spanID", spCtx.SpanID().String()),
	)
}
