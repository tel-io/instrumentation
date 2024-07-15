package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	natsmw "github.com/tel-io/instrumentation/middleware/nats/v2"
	"github.com/tel-io/tel/v2"
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

	xxx(ctx)

	con, err := nats.Connect(addr)
	if err != nil {
		t.Panic("connect", tel.Error(err))
	}

	mw := natsmw.New(natsmw.WithTel(t), natsmw.WithDump(true), natsmw.WithDump(true)).Use(con)

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

			// sync
			sh := mw.BuildWrappedHandler(func(ctx context.Context, msg *nats.Msg) error {
				// perform respond with possible error returning
				return msg.Respond([]byte("SYNC HANDLER"))
			})
			ch := make(chan *nats.Msg)
			_, _ = mw.QueueSubscribeSyncWithChan("nats.ch_subscriber", "consumer", ch)

		X:
			for {
				select {
				case <-ctx.Done():
					break X
				case msg := <-ch:
					sh(msg)
				}
			}

			//<-ctx.Done()

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

func xxx(ccx context.Context) {
	span, ctx := tel.FromCtx(ccx).StartSpan(ccx, "SOME INFO")
	defer span.End()

	// fiels
	tel.FromCtx(ctx).PutFields(tel.String("SOME KEY", "SOME VALUE"))

	tel.FromCtx(ccx).Info(">>>>>> INFO IT IS",
		tel.Bool("xxx", true), tel.String("vvv", "qqq"))

	tel.FromCtx(ctx).Info(">>>>> EMBED INFO",
		tel.Bool("xxx", true), tel.String("vvv", "qqq"))

	// DB
	func(cci context.Context) {
		tel.FromCtx(cci).PutFields(tel.String("USER_ID", "123"))
	}(ctx)

	// Redis
	func(cci context.Context) {
		sp, _ := tel.FromCtx(cci).StartSpan(cci, "REDDDD")
		defer sp.End()

		time.Sleep(time.Second)

		tel.FromCtx(cci).PutFields(tel.String("REDDDD", "xxxxxxx"))
	}(ctx)

	err := func(cci context.Context) error {
		e := func() error {
			return nil
		}()
		if e != nil {
			return errors.WithStack(e)
		}

		e2 := func() error {
			return errors.WithStack(fmt.Errorf(">>> error some"))
		}()

		if e2 != nil {
			return errors.WithStack(e2)
		}

		return nil

	}(ctx)

	if err != nil {
		tel.FromCtx(ctx).Error("some err", tel.Error(err))
	}

}
