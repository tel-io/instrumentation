package nats

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Middleware interface {
	apply(next MsgHandler) MsgHandler
}

// MsgHandler our desired way to handle subscriptions
// ctx allow inside function continue traces or pass log attachment
// error return allow subMiddleware to understand behaviour of system what has gone here,
// and it could change differently
type MsgHandler func(ctx context.Context, msg *nats.Msg) error

// Interceptor  ...
type Interceptor func(next MsgHandler) MsgHandler

// MiddlewareChain - MsgHandler decorator with subMiddleware
func MiddlewareChain(mw ...Middleware) Interceptor {
	return func(next MsgHandler) MsgHandler {
		for _, m := range mw {
			next = m.apply(next)
		}

		return next
	}
}

func ChainInterceptor(interceptors ...Interceptor) Interceptor {
	return func(next MsgHandler) MsgHandler {
		for _, mw := range interceptors {
			next = mw(next)
		}

		return next
	}
}
