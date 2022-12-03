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
	nameFn NameFn

	dumpPayloadOnError bool
	dumpRequest        bool
}

func NewLogs(fn NameFn, dumpPayloadOnError, dumpRequest bool) *Logs {
	return &Logs{
		nameFn:             fn,
		dumpPayloadOnError: dumpPayloadOnError,
		dumpRequest:        dumpRequest,
	}
}

func (t *Logs) apply(next MsgHandler) MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) (err error) {
		defer func(start time.Time) {
			kind := extractBaggageKind(ctx)

			l := tel.FromCtx(ctx).With(
				zap.String("duration", time.Since(start).String()),
			)

			lvl := zapcore.DebugLevel
			if err != nil {
				lvl = zapcore.ErrorLevel
				l = l.With(zap.Error(err))
			}

			if ((t.dumpPayloadOnError && err != nil) || t.dumpRequest) && msg.Data != nil {
				l = l.With(zap.String("request", string(msg.Data)))
			}

			l.Check(lvl, t.nameFn(kind, msg)).Write()
		}(time.Now())

		return next(ctx, msg)
	}
}
