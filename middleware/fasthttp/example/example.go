package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	mw "github.com/tel-io/instrumentation/middleware/fasthttp"
	"github.com/tel-io/tel/v2"
	"github.com/valyala/fasthttp"
)

const (
	port = ":9000"
)

func main() {
	ccx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		cn := make(chan os.Signal, 1)
		signal.Notify(cn, syscall.SIGTERM, syscall.SIGINT, syscall.SIGTERM)
		<-cn
		cancel()
	}()

	cfg := tel.GetConfigFromEnv()
	cfg.LogEncode = "console"
	cfg.Namespace = "TEST"
	cfg.Service = "DEMO-FASTHTTP"

	t, cc := tel.New(ccx, cfg)
	defer cc()

	t.Info("start", tel.String("addr", port))

	middleware := mw.ServerMiddleware()

	go func() {
		if err := fasthttp.ListenAndServe(port, middleware(func(ctx *fasthttp.RequestCtx) {
			// EXTRACT OTEL CONTEXT WITH SPAN AND EVERYTHING.....
			ctn := mw.GetNativeContext(ctx)

			tel.FromCtx(ctn).PutFields(tel.String("XXXX", "YYYYY"))
			span, cxx := tel.FromCtx(ctn).StartSpan(ctn, "VALERA")
			defer span.End()

			tel.FromCtx(cxx).Info("world", tel.Bool("tex", true))

			//panic(true)

			_, _ = ctx.Write([]byte("HELLO"))

		})); err != nil {
			t.Error("start", tel.Error(err))
		}
	}()

	<-ccx.Done()
}
