package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/tel-io/tel/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// Logs dump some payload
type Logs struct {
	*config
}

func NewLogs(cfg *config) *Logs {
	return &Logs{config: cfg}
}

func (t *Logs) apply(next MsgHandler) MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) (err error) {
		defer func(start time.Time) {
			l := tel.FromCtx(ctx).With(
				zap.String("duration", time.Since(start).String()),
			)

			lvl := zapcore.DebugLevel
			if err != nil {
				lvl = zapcore.ErrorLevel
				l = l.With(zap.Error(err))
			}

			if ((t.config.dumpPayloadOnError && err != nil) || t.config.dumpRequest) && msg.Data != nil {
				l = l.With(zap.String("request", string(msg.Data)))
			}

			l.Check(lvl, t.subNameFn(msg)).Write()

		}(time.Now())

		return next(ctx, msg)
	}
}
