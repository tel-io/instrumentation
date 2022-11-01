package fasthttp

import (
	"context"
	"net/http"
	"net/http/httptest"

	mw "github.com/tel-io/instrumentation/middleware/http"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type Middleware func(next fasthttp.RequestHandler) fasthttp.RequestHandler

func ServerMiddleware(opts ...mw.Option) Middleware {
	instance := mw.ServerMiddlewareAll(opts...)
	w := httptest.NewRecorder() // not use so can do like this

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			req, err := http.NewRequest(http.MethodGet, "", nil)
			if err == nil {
				_ = fasthttpadaptor.ConvertRequest(ctx, req, true)
			}

			instance(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				WithNativeContext(ctx, r.Context())
				next(ctx)

				w.WriteHeader(ctx.Response.StatusCode())
				_, _ = w.Write(ctx.Response.Body())
			})).ServeHTTP(w, req)
		}
	}
}

const cVal = "tcx"

// GetNativeContext retrieve a standard library context
// from FastHTTP request context.
func GetNativeContext(ctx *fasthttp.RequestCtx) context.Context {
	return ctx.UserValue(cVal).(context.Context)
}

// WithNativeContext stores a standard library context
//into FastHTTP request context.
func WithNativeContext(ctx *fasthttp.RequestCtx, nativeCtx context.Context) {
	ctx.SetUserValue(cVal, nativeCtx)
}
