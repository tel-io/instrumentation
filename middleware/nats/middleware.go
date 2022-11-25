package nats

import (
	"context"
	"github.com/nats-io/nats.go"
)

type Middleware interface {
	apply(next MsgHandler) MsgHandler
}

//MsgHandler our desired way to handle subscriptions
// ctx allow inside function continue traces or pass log attachment
// error return allow middleware to understand behaviour of system what has gone here,
// and it could change differently
type MsgHandler func(ctx context.Context, msg *nats.Msg) error
