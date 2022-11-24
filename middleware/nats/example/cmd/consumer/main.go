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
	cfg.Service = "NATS_CONSUMER"
	cfg.MonitorConfig.Enable = false

	t, cc := tel.New(ccx, cfg, tel.WithHealthCheckers())
	defer cc()

	ctx := tel.WithContext(ccx, t)

	t.Info("nats", tel.String("collector", cfg.Addr))

	con, err := nats.Connect(addr)
	if err != nil {
		t.Panic("connect", tel.Error(err))
	}

	mw := natsmw.New(con, natsmw.WithTel(t), natsmw.WithDumpRequest(true), natsmw.WithDumpResponse(true))

	for i := 0; i < 1; i++ {
		go func() {
			subscribe, err := mw.QueueSubscribe("nats.demo", "consumer", func(ctx context.Context, msg *nats.Msg) error {
				// send as reply
				fmt.Println(string(msg.Data))
				return msg.Respond([]byte("HELLO"))
			})
			nullErr(err)
			//subscribe.SetPendingLimits(1, 10000000)

			crash, err := mw.QueueSubscribe("nats.crash", "consumer", func(ctx context.Context, msg *nats.Msg) error {
				time.Sleep(time.Microsecond)
				panic("some panic")
			})
			nullErr(err)

			someErr, err := mw.QueueSubscribe("nats.err", "consumer", func(ctx context.Context, msg *nats.Msg) error {
				time.Sleep(time.Millisecond)
				return fmt.Errorf("some errro")
			})
			nullErr(err)

			tmout, err := mw.QueueSubscribe("nats.timeout", "consumer", func(ctx context.Context, msg *nats.Msg) error {
				time.Sleep(time.Second)
				return msg.Respond([]byte("HELLO"))
			})
			nullErr(err)

			<-ctx.Done()

			_ = subscribe.Unsubscribe()
			_ = crash.Unsubscribe()
			_ = someErr.Unsubscribe()
			_ = tmout.Unsubscribe()
		}()
	}

	<-ctx.Done()
}

func nullErr(err error) {
	if err != nil {
		tel.Global().Panic("err", tel.Error(err))
	}
}
