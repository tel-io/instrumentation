package chi

import (
	"net/http"

	mw "github.com/tel-io/instrumentation/middleware/http"

	"github.com/go-chi/chi/v5"
)

const defaultPath = "<no-path>"

// getPath extracts path from chi route for http MW for correct metric exposure
func getPath(r *http.Request) string {
	if ctx := chi.RouteContext(r.Context()); ctx != nil {
		return ctx.RoutePattern()
	}

	return defaultPath
}

//HTTPServerMiddlewareAll including with path extractor with overwrite option via WithPathExtractor option append
func HTTPServerMiddlewareAll(opts ...mw.Option) func(http.Handler) http.Handler {
	return mw.ServerMiddlewareAll(
		append([]mw.Option{mw.WithPathExtractor(getPath)}, opts...)...,
	)
}
