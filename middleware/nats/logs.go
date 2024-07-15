package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/tel-io/tel/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logs dump some payload
type Logs struct {
	nameFn NameFn

	dumpPayloadOnError bool
	dump               bool
}

func NewLogs(fn NameFn, dumpPayloadOnError, dumpRequest bool) *Logs {
	return &Logs{
		nameFn:             fn,
		dumpPayloadOnError: dumpPayloadOnError,
		dump:               dumpRequest,
	}
}

func (t *Logs) apply(next MsgHandler) MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) (err error) {
		defer func(start time.Time) {
			var (
				kind = extractBaggageKind(ctx)
				attr = ExtractAttributes(msg, kind, true)
				tele = tel.FromCtx(ctx).Copy()
				l    = tele.PutAttr(attr...).With()
				lvl  = zapcore.DebugLevel
			)

			if err != nil {
				lvl = zapcore.ErrorLevel
				l = l.With(zap.Error(err))
			}

			if ((t.dumpPayloadOnError && err != nil) || t.dump) && msg.Data != nil {
				l = l.With(zap.String(PayloadKey, string(msg.Data)))
			}

			l.Check(lvl, t.nameFn(kind, msg)).Write(
				tel.String(string(Duration), time.Since(start).String()),
			)
		}(time.Now())

		return next(ctx, msg)
	}
}
