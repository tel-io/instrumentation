package echo

import (
	"github.com/labstack/echo/v4"
	mw "github.com/tel-io/instrumentation/middleware/http"
	"go.opentelemetry.io/otel/baggage"
	"net/http"
	"net/url"
)

func extractor(r *http.Request) string {
	b := baggage.FromContext(r.Context())

	v, err := url.PathUnescape(b.Member("path").Value())
	if err != nil {
		return r.URL.Path
	}

	return v
}

// HTTPServerMiddlewareAll all in one mw packet
func HTTPServerMiddlewareAll(opts ...mw.Option) echo.MiddlewareFunc {
	return WrapMiddleware(mw.ServerMiddlewareAll(
		append([]mw.Option{mw.WithPathExtractor(extractor)}, opts...)...,
	))
}

// WrapMiddleware wraps `func(http.Handler) http.Handler` into `echo.MiddlewareFunc`
func WrapMiddleware(m func(http.Handler) http.Handler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			method, err := baggage.NewMember("path", c.Path())
			if err == nil {
				if b, err2 := baggage.New(method); err2 == nil {
					r := c.Request().Clone(baggage.ContextWithBaggage(c.Request().Context(), b))
					c.SetRequest(r)
				}
			}

			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.SetRequest(r)
				c.SetResponse(echo.NewResponse(w, c.Echo()))

				err = next(c)
			})).ServeHTTP(c.Response(), c.Request())
			return
		}
	}
}
