package nats

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/nats-io/nats.go"
	"github.com/tel-io/tel/v2"
)

type Recovery struct{}

func NewRecovery() *Recovery {
	return &Recovery{}
}

func (t *Recovery) apply(next MsgHandler) MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) (err error) {
		defer func() {
			hasRecovery := recover()

			if hasRecovery != nil {
				//nolint: goerr113
				err = fmt.Errorf("recovery info: %+v", hasRecovery)

				tel.FromCtx(ctx).Error("recovery", tel.Error(err))

				if tel.FromCtx(ctx).IsDebug() {
					debug.PrintStack()
				}
			}
		}()

		return next(ctx, msg)
	}
}
