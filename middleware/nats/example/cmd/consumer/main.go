package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	natsmw "github.com/tel-io/instrumentation/middleware/nats"
	"github.com/tel-io/tel/v2"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go"
)

var addr = "nats://127.0.0.1:4222"

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
	cfg.Service = "NATS.CONSUMER"
	cfg.MonitorConfig.Enable = false

	t, cc := tel.New(ccx, cfg, tel.WithHealthCheckers())
	defer cc()

	ctx := tel.WithContext(ccx, t)

	t.Info("nats", tel.String("collector", cfg.Addr))

	con, err := nats.Connect(addr)
	if err != nil {
		t.Panic("connect", tel.Error(err))
	}

	mw := natsmw.New(natsmw.WithTel(t))

	x, err := natsmw.NewSubscriptionMetrics(natsmw.WithTel(t))
	nullErr(err)

	for i := 0; i < 100; i++ {
		go func() {
			subscribe, err := con.QueueSubscribe("nats.demo", "consumer", mw.Handler(func(ctx context.Context, sub string, data []byte) ([]byte, error) {
				// send as reply
				time.Sleep(time.Second)
				return []byte("HELLO"), nil
			}))
			nullErr(err)

			crash, err := con.QueueSubscribe("nats.crash", "consumer", mw.Handler(func(ctx context.Context, sub string, data []byte) ([]byte, error) {
				time.Sleep(time.Microsecond)
				panic("some panic")
			}))
			nullErr(err)

			someErr, err := con.QueueSubscribe("nats.err", "consumer", mw.Handler(func(ctx context.Context, sub string, data []byte) ([]byte, error) {
				time.Sleep(time.Millisecond)
				return nil, fmt.Errorf("some errro")
			}))
			nullErr(err)

			x.Register(subscribe, crash, someErr)

			<-ctx.Done()

			_ = subscribe.Unsubscribe()
			_ = crash.Unsubscribe()
			_ = someErr.Unsubscribe()
		}()
	}

	<-ctx.Done()
}

func nullErr(err error) {
	if err != nil {
		tel.Global().Panic("err", tel.Error(err))
	}
}
