package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/baggage"
)

type Publish interface {
	PublishWithContext(ctx context.Context, subj string, data []byte) error
	PublishMsgWithContext(ctx context.Context, msg *nats.Msg) error
	PublishRequestWithContext(ctx context.Context, subj, reply string, data []byte) error
	RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error)
	RequestMsgWithContext(ctx context.Context, msg *nats.Msg) (*nats.Msg, error)
}

type CommonPublish struct {
	interceptor Interceptor

	*nats.Conn
}

func NewCommonPublish(conn *nats.Conn, interceptor Interceptor) *CommonPublish {
	return &CommonPublish{Conn: conn, interceptor: interceptor}
}

// PublishMsgWithContext publishes the Msg structure, which includes the
// Subject, an optional Reply and an optional Data field.
func (c *CommonPublish) PublishMsgWithContext(ctx context.Context, msg *nats.Msg) error {
	if k, e := baggage.NewMember(KindKey, KindPub); e == nil {
		if b, ee := baggage.New(k); ee == nil {
			ctx = baggage.ContextWithBaggage(ctx, b)
		}
	}

	return c.interceptor(func(ctx context.Context, msg *nats.Msg) error {
		return c.Conn.PublishMsg(msg)
	})(ctx, msg)
}

// PublishWithContext publishes the data argument to the given subject. The data
// argument is left untouched and needs to be correctly interpreted on
// the receiver.
func (c *CommonPublish) PublishWithContext(ctx context.Context, subj string, data []byte) error {
	return c.PublishMsgWithContext(ctx, &nats.Msg{Subject: subj, Data: data})
}

// PublishRequestWithContext will perform a Publish() expecting a response on the
// reply subject. Use Request() for automatically waiting for a response
// inline.
func (c *CommonPublish) PublishRequestWithContext(ctx context.Context, subj, reply string, data []byte) error {
	return c.PublishMsgWithContext(ctx, &nats.Msg{Subject: subj, Reply: reply, Data: data})
}

// RequestMsgWithContext takes a context, a subject and payload
// in bytes and request expecting a single response.
func (c *CommonPublish) RequestMsgWithContext(ccx context.Context, msg *nats.Msg) (res *nats.Msg, err error) {
	var ctx = ccx

	if k, e := baggage.NewMember(KindKey, KindPub); e == nil {
		if b, ee := baggage.New(k); ee == nil {
			ctx = baggage.ContextWithBaggage(ctx, b)
		}
	}

	_ = c.interceptor(func(ctx context.Context, msg *nats.Msg) error {
		res, err = c.Conn.RequestMsgWithContext(ctx, msg)
		var infRes = res

		if err != nil {
			infRes = &nats.Msg{Subject: msg.Subject, Header: msg.Header}
		} else {
			infRes = &nats.Msg{
				Subject: msg.Subject,
				Header:  res.Header,
				Data:    res.Data,
			}
		}

		c.infiltrateResponse(ccx, infRes, err)
		return err
	})(ctx, msg)

	return res, err
}

// RequestWithContext will send a request payload and deliver the response message,
// or an error, including a timeout if no message was received properly.
func (c *CommonPublish) RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error) {
	return c.RequestMsgWithContext(ctx, &nats.Msg{Subject: subj, Data: data})
}

// infiltrateResponse just register
func (c *CommonPublish) infiltrateResponse(ctx context.Context, msg *nats.Msg, err error) {
	if k, e := baggage.NewMember(KindKey, KindRespond); e == nil {
		if b, ee := baggage.New(k); ee == nil {
			ctx = baggage.ContextWithBaggage(ctx, b)
		}
	}

	_ = c.interceptor(func(ctx context.Context, _ *nats.Msg) error {
		return err
	})(ctx, msg)
}
