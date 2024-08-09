package gaphttp

import (
	"net/http"

	"go.uber.org/zap"
)

func NewMiddlewareRecovery(logger *zap.Logger, option *MiddlewareOptions) Middleware {
	// recover disabled
	if !option.EnabledRecover {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(
				func(writer http.ResponseWriter, request *http.Request) {
					next.ServeHTTP(writer, request)
				},
			)
		}
	}

	// recover enabled
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(writer http.ResponseWriter, request *http.Request) {
				defer func() {
					if rvr := recover(); rvr != nil && rvr != http.ErrAbortHandler {
						logger.Error("middleware catch panic", zap.Any("panic.body", rvr))

						writer.WriteHeader(http.StatusInternalServerError)
					}
				}()

				next.ServeHTTP(writer, request)
			},
		)
	}
}
