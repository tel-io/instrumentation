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
	mw "github.com/tel-io/instrumentation/middleware/nats"
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

	t, cc := tel.New(ccx, cfg, tel.WithHealthCheckers())
	defer cc()

	ctx := tel.WithContext(ccx, t)

	t.Info("nats", tel.String("collector", cfg.Addr))

	con, err := nats.Connect(addr)
	if err != nil {
		t.Panic("connect", tel.Error(err))
	}

	connection := mw.New(con, mw.WithTel(t))

	for i := 0; i < threads; i++ {
		go run(ctx, connection, i)
	}

	<-ctx.Done()
}

func run(ctx context.Context, con *mw.ConnContext, i int) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
			switch rand.Int63n(20) {
			case 0:
				_ = con.PublishWithContext(ctx, "nats.err", []byte("HELLO"))
			case 1:
				_ = con.PublishWithContext(ctx, "nats.crash", []byte("HELLO"))
			case 3:
				cxx, cancel := context.WithTimeout(ctx, time.Second)
				_, _ = con.RequestWithContext(cxx, "nats.timeout", []byte("HELLO"))
				cancel()
			case 4:
				cxx, cancel := context.WithTimeout(ctx, time.Millisecond)
				_, _ = con.RequestWithContext(cxx, "nats.no-respond", []byte("HELLO"))
				cancel()
			default:
				cxx, cancel := context.WithTimeout(ctx, time.Minute)
				_, _ = con.RequestWithContext(cxx, "nats.demo", []byte("HELLO"))
				cancel()
			}
		}
	}
}
