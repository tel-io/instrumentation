package main

import (
	"context"
	"fmt"
	"github.com/d7561985/tel/v2"
	natsmw "github.com/tel-io/instrumentation/middleware/nats"
	"os"
	"os/signal"
	"syscall"
	"time"

	health "github.com/d7561985/tel/v2/monitoring/heallth"
	_ "github.com/joho/godotenv/autoload"
	"github.com/nats-io/nats.go"
)

var addr = "nats://127.0.0.1:4222"
var cAddr = "http://127.0.0.1:54239"

func main() {
	ccx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		cn := make(chan os.Signal, 1)
		signal.Notify(cn, os.Kill, syscall.SIGINT, syscall.SIGTERM)
		<-cn
		cancel()
	}()

	cfg := tel.GetConfigFromEnv()
	cfg.LogEncode = "console"
	cfg.Namespace = "TEST"
	cfg.Service = "NATS.CONSUMER"
	cfg.MonitorConfig.Enable = false

	t, cc := tel.New(ccx, cfg)
	defer cc()

	ctx := tel.WithContext(ccx, t)
	t.AddHealthChecker(ctx, tel.HealthChecker{Handler: health.NewCompositeChecker()})

	t.Info("nats", tel.String("collector", cfg.Addr))

	con, err := nats.Connect(addr)
	if err != nil {
		t.Panic("connect", tel.Error(err))
	}

	mw := natsmw.New(natsmw.WithTel(t))
	subscribe, err := con.Subscribe("nats.demo", mw.Handler(func(ctx context.Context, sub string, data []byte) ([]byte, error) {
		// send as reply
		return []byte("HELLO"), nil
	}))

	crash, err := con.Subscribe("nats.crash", mw.Handler(func(ctx context.Context, sub string, data []byte) ([]byte, error) {
		time.Sleep(time.Millisecond)
		panic("some panic")
		return nil, nil
	}))

	someErr, err := con.Subscribe("nats.err", mw.Handler(func(ctx context.Context, sub string, data []byte) ([]byte, error) {
		time.Sleep(time.Millisecond)
		return nil, fmt.Errorf("some errro")
	}))

	defer subscribe.Unsubscribe()
	defer crash.Unsubscribe()
	defer someErr.Unsubscribe()

	<-ctx.Done()
}
