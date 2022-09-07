package echo

import (
	"net/http"

	"github.com/labstack/echo/v4"
	mw "github.com/tel-io/instrumentation/middleware/http"
)

// GetPath extracts path from chi route for http MW for correct metric exposure
func GetPath(r *http.Request) string {
	return echo.GetPath(r)
}

// HTTPServerMiddlewareAll all in one mw packet
func HTTPServerMiddlewareAll(opts ...mw.Option) echo.MiddlewareFunc {
	return echo.WrapMiddleware(mw.ServerMiddlewareAll(
		append(opts, mw.WithPathExtractor(GetPath))...,
	))
}
