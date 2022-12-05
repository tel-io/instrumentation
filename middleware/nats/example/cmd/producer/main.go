package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go"
	mw "github.com/tel-io/instrumentation/middleware/nats/v2"
	"github.com/tel-io/tel/v2"
)

var addr = "nats://127.0.0.1:4222"

const threads = 100

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
	cfg.Service = "NATS_PRODUCER"
	cfg.MonitorConfig.Enable = false

	t, closer := tel.New(ccx, cfg, tel.WithHealthCheckers())
	defer closer()

	ctx := tel.WithContext(ccx, t)

	t.Info("nats", tel.String("collector", cfg.Addr))

	con, err := nats.Connect(addr)
	if err != nil {
		t.Panic("connect", tel.Error(err))
	}

	connection := mw.New(mw.WithTel(t))

	for i := 0; i < threads; i++ {
		go run(ctx, connection, con)
	}

	<-ctx.Done()
}

func run(cxt context.Context, core *mw.Core, cc *nats.Conn) {
	//js, _ := con.JetStream()
	for {
		tele := tel.FromCtx(cxt).Copy()

		func() {
			span, ctx := tele.StartSpan(tele.Ctx(), "test producer")
			defer span.End()

			select {
			case <-cxt.Done():
				return
			case <-time.After(time.Second):
				switch rand.Int63n(20) {
				case 0:
					_ = core.Use(cc).PublishWithContext(ctx, "nats.err", []byte("HELLO"))
				case 1:
					_ = core.Use(cc).PublishWithContext(ctx, "nats.crash", []byte("HELLO"))
				case 3:
					cxx, cancel := context.WithTimeout(ctx, time.Second)
					_, _ = core.Use(cc).RequestWithContext(cxx, "nats.timeout", []byte("HELLO"))
					cancel()
				case 4:
					cxx, cancel := context.WithTimeout(ctx, time.Millisecond)
					_, _ = core.Use(cc).RequestWithContext(cxx, "nats.no-respond", []byte("HELLO"))
					cancel()
				//case 5:
				//	_, _ = js.JS().Publish("stream.demo", []byte("HELLO")) //nats.ExpectStream("demo"),
				case 6:
					_ = core.Use(cc).PublishWithContext(context.Background(), "nats.bad_context", []byte("HELLO"))
				case 7:
					_, _ = core.Use(cc).RequestWithContext(ctx, "nats.ch_subscriber", []byte("HELLO"))
				default:
					cxx, cancel := context.WithTimeout(ctx, time.Minute)
					_, _ = core.Use(cc).RequestWithContext(cxx, "nats.demo", []byte("HELLO"))
					cancel()
				}
			}
		}()

	}
}
