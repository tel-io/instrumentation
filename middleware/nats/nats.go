package nats

import (
	"context"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/baggage"
)

// PostFn callback function which got new instance of tele inside ctx
// and msg sub + data
// Deprecated: legacy function, but we use it via conn wrapper: QueueSubscribeMW or SubscribeMW just for backport compatibility
type PostFn func(ctx context.Context, sub string, data []byte) ([]byte, error)

var ErrMultipleMiddleWare = errors.New("not allow create multiple instances")

var mx sync.RWMutex

// Core features for context
type Core struct {
	*config

	subInter Interceptor
	pubInter Interceptor

	subMeter *SubscriptionStatMetric
}

// New subMiddleware instance
func New(opts ...Option) *Core {
	// We don't allow to create multiple instances
	if !mx.TryLock() {
		panic(ErrMultipleMiddleWare)
	}

	cfg := newConfig(opts)

	// create only once
	sb, err := NewSubscriptionStatMetrics(opts...)
	if err != nil {
		cfg.tele.Panic("wrap connection", tel.Error(err))
	}

	// create instances of pub mw only once
	plist := cfg.pubMiddleware()

	// create instances of mw only once
	list := cfg.subMiddleware()

	return &Core{
		config:   cfg,
		subMeter: sb,
		subInter: MiddlewareChain(list...),
		pubInter: MiddlewareChain(plist...),
	}
}

// Use connection with subMiddleware
func (c *Core) Use(conn *nats.Conn) *ConnContext {
	return &ConnContext{
		conn:       conn,
		Publish:    NewCommonPublish(conn, c.pubInter),
		Subscriber: NewCommonSubscriber(conn, c),
		Core:       c,
	}
}

// subWrap wrapper for subscriber
func (c *Core) subWrap(next MsgHandler) nats.MsgHandler {
	in := c.subInter(next)

	var bag *baggage.Baggage
	if k, e := baggage.NewMember(KindKey, KindSub); e == nil {
		if b, ee := baggage.New(k); ee == nil {
			bag = &b
		}
	}

	return func(msg *nats.Msg) {
		// init context for instance
		ctx := c.config.tele.Ctx()

		if bag != nil {
			ctx = baggage.ContextWithBaggage(ctx, *bag)
		}

		_ = in(ctx, msg)
	}
}
