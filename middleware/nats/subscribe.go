package nats

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Subscriber interface {
	Subscribe(subj string, cb MsgHandler) (*nats.Subscription, error)
	QueueSubscribe(subj, queue string, cb MsgHandler) (*nats.Subscription, error)
	QueueSubscribeMW(subj, queue string, next PostFn) (*nats.Subscription, error)
	SubscribeMW(subj string, cb PostFn) (*nats.Subscription, error)

	QueueSubscribeSyncWithChan(subj, queue string, ch chan *nats.Msg) (*nats.Subscription, error)

	BuildWrappedHandler(next MsgHandler) nats.MsgHandler
}

type CommonSubscribe struct {
	*Core

	conn *nats.Conn
}

func NewCommonSubscriber(conn *nats.Conn, core *Core) *CommonSubscribe {
	return &CommonSubscribe{
		Core: core,
		conn: conn,
	}
}

// Subscribe will express interest in the given subject. The subject
// can have wildcards.
// There are two type of wildcards: * for partial, and > for full.
// A subscription on subject time.*.east would receive messages sent to time.us.east and time.eu.east.
// A subscription on subject time.us.> would receive messages sent to
// time.us.east and time.us.east.atlanta, while time.us.* would only match time.us.east
// since it can't match more than one token.
// Messages will be delivered to the associated MsgHandler.
func (c *CommonSubscribe) Subscribe(subj string, cb MsgHandler) (*nats.Subscription, error) {
	return c.subMeter.Hook(
		c.conn.Subscribe(subj, c.BuildWrappedHandler(cb)),
	)
}

// QueueSubscribe creates an asynchronous queue subscriber on the given subject.
// All subscribers with the same queue name will form the queue group and
// only one member of the group will be selected to receive any given
// message asynchronously.
func (c *CommonSubscribe) QueueSubscribe(subj, queue string, cb MsgHandler) (*nats.Subscription, error) {
	return c.subMeter.Hook(
		c.conn.QueueSubscribe(subj, queue, c.BuildWrappedHandler(cb)),
	)
}

// QueueSubscribeSyncWithChan will express interest in the given subject.
// All subscribers with the same queue name will form the queue group
// and only one member of the group will be selected to receive any given message,
// which will be placed on the channel.
// You should not close the channel until sub.Unsubscribe() has been called.
//
// NOTE: middleware only subscription hook performed
func (c *CommonSubscribe) QueueSubscribeSyncWithChan(subj, queue string, ch chan *nats.Msg) (*nats.Subscription, error) {
	return c.subMeter.Hook(c.conn.QueueSubscribeSyncWithChan(subj, queue, ch))
}

// QueueSubscribeMW mw callback function, just legacy
// Deprecated: just backport compatibility for PostFn legacy
func (c *CommonSubscribe) QueueSubscribeMW(subj, queue string, next PostFn) (*nats.Subscription, error) {
	return c.QueueSubscribe(subj, queue, func(ctx context.Context, msg *nats.Msg) error {
		resp, err := next(ctx, msg.Subject, msg.Data)
		if err != nil || c.config.postHook == nil {
			return nil
		}

		_ = c.config.postHook(ctx, msg, resp)
		return nil
	})
}

// SubscribeMW backport compatible function for previous mw approach
// Deprecated: just backport compatibility for PostFn legacy
func (c *CommonSubscribe) SubscribeMW(subj string, cb PostFn) (*nats.Subscription, error) {
	return c.QueueSubscribeMW(subj, "", cb)
}

// BuildWrappedHandler allow to create own mw, for bach processing for example or so on
func (c *CommonSubscribe) BuildWrappedHandler(next MsgHandler) nats.MsgHandler {
	return c.subWrap(next)
}
