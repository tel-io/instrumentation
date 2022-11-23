package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/tel-io/tel/v2"
)

// Conn wrapper for nats.Conn aks mw connection approach
//
// Features:
// Expose subscription stats via function overwrite
type Conn struct {
	conn *nats.Conn
	Publish

	*config

	list    []Middleware
	pubList []PubMiddleware

	subMeter *SubscriptionStatMetric
}

func WrapConn(conn *nats.Conn, opts ...Option) *Conn {
	cfg := newConfig(opts)

	sb, err := NewSubscriptionStatMetrics(opts...)
	if err != nil {
		cfg.tele.Panic("wrap connection", tel.Error(err))
	}

	// init publish
	var pub PubMiddleware
	pub = NewCommonPublish(conn)

	pubList := cfg.DefaultPubMiddleware()
	for _, mw := range pubList {
		pub = mw.apply(pub)
	}

	return &Conn{
		conn: conn,

		subMeter: sb,
		config:   cfg,
		Publish:  pub,

		list: cfg.Middleware(),
	}
}

// hook for subscription handling
func (c *Conn) hook(s *nats.Subscription, err error) (*nats.Subscription, error) {
	if err != nil {
		return nil, err
	}

	c.subMeter.Register(s)

	return s, nil
}

// wrap Middleware wrap
func (c *Conn) wrap(in MsgHandler) nats.MsgHandler {
	// init context for instance
	cxt := c.config.tele.Copy().Ctx()

	for _, cb := range c.list {
		in = cb.apply(in)
	}

	return func(msg *nats.Msg) {
		_ = in(cxt, msg)
	}
}

// Conn unwrap connection
func (c *Conn) Conn() *nats.Conn {
	return c.conn
}

// Subscribe will express interest in the given subject. The subject
// can have wildcards.
// There are two type of wildcards: * for partial, and > for full.
// A subscription on subject time.*.east would receive messages sent to time.us.east and time.eu.east.
// A subscription on subject time.us.> would receive messages sent to
// time.us.east and time.us.east.atlanta, while time.us.* would only match time.us.east
// since it can't match more than one token.
// Messages will be delivered to the associated MsgHandler.
func (c *Conn) Subscribe(subj string, cb MsgHandler) (*nats.Subscription, error) {
	return c.hook(
		c.conn.Subscribe(subj, c.wrap(cb)),
	)
}

// QueueSubscribe creates an asynchronous queue subscriber on the given subject.
// All subscribers with the same queue name will form the queue group and
// only one member of the group will be selected to receive any given
// message asynchronously.
func (c *Conn) QueueSubscribe(subj, queue string, cb MsgHandler) (*nats.Subscription, error) {
	return c.hook(
		c.conn.QueueSubscribe(subj, queue, c.wrap(cb)),
	)
}

// QueueSubscribeMW mw callback function, just legacy
// Deprecated: just backport compatibility for PostFn legacy
func (c *Conn) QueueSubscribeMW(subj, queue string, next PostFn) (*nats.Subscription, error) {
	return c.QueueSubscribe(subj, queue, func(ctx context.Context, msg *nats.Msg) error {
		resp, err := next(ctx, msg.Subject, msg.Data)
		if err != nil || c.config.postHook == nil {
			return nil
		}

		err = c.config.postHook(ctx, msg, resp)
		return nil
	})
}

// SubscribeMW backport compatible function for previous mw approach
// Deprecated: just backport compatibility for PostFn legacy
func (c *Conn) SubscribeMW(subj string, cb PostFn) (*nats.Subscription, error) {
	return c.QueueSubscribeMW(subj, "", cb)
}
