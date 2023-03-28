package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	mw "github.com/tel-io/instrumentation/middleware/fasthttp"
	"github.com/tel-io/tel/v2"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
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

	// client communication
	worker(ccx)

	<-ccx.Done()
}

func worker(ccx context.Context) {
	cfg := tel.GetConfigFromEnv()
	cfg.LogEncode = "console"
	cfg.Namespace = "TEST"
	cfg.Service = "DEMO-CLIENT"

	t, cc := tel.New(ccx, cfg)
	defer cc()

	// client communication
	propagators := otel.GetTextMapPropagator()
	c := fasthttp.Client{}

	for {
		select {
		case <-ccx.Done():
			return
		case <-time.After(time.Second):
			cxt := t.Copy().Ctx()
			span, ctx := tel.FromCtx(cxt).StartSpan(cxt, "req")

			req := fasthttp.AcquireRequest()
			res := fasthttp.AcquireResponse()
			req.SetHost("127.0.0.1" + port)

			propagators.Inject(ctx, mw.NewCarrier(&req.Header))

			if err := c.Do(req, res); err != nil {
				span.RecordError(err)
			}

			span.End()
		}
	}
}
