package nats

import (
	"context"

	"github.com/nats-io/nats.go"
)

type Publish interface {
	PublishWithContext(ctx context.Context, subj string, data []byte) error
	PublishMsgWithContext(ctx context.Context, msg *nats.Msg) error
	PublishRequestWithContext(ctx context.Context, subj, reply string, data []byte) error
	RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error)
	RequestMsgWithContext(ctx context.Context, msg *nats.Msg) (*nats.Msg, error)
}

//go:generate mockery --name natsPublisher --dir . --outpkg natsmockery  --exported
type natsPublisher interface {
	PublishMsg(m *nats.Msg) error
	RequestMsgWithContext(ctx context.Context, msg *nats.Msg) (*nats.Msg, error)
}

var _ natsPublisher = &nats.Conn{}

type CommonPublish struct {
	interceptor Interceptor

	Conn natsPublisher
}

// NewCommonPublish create instance of wrapper publisher
func NewCommonPublish(conn *nats.Conn, interceptor Interceptor) *CommonPublish {
	return &CommonPublish{Conn: conn, interceptor: interceptor}
}

// PublishMsgWithContext publishes the Msg structure, which includes the
// Subject, an optional Reply and an optional Data field.
func (c *CommonPublish) PublishMsgWithContext(ctx context.Context, msg *nats.Msg) error {
	ctx = WrapKindOfContext(ctx, KindPub)

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
func (c *CommonPublish) RequestMsgWithContext(ccx context.Context, m *nats.Msg) (res *nats.Msg, err error) {
	var ctx = WrapKindOfContext(ccx, KindRequest)

	err = c.interceptor(func(ctx context.Context, msg *nats.Msg) error {
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

			if infRes.Header == nil {
				infRes.Header = nats.Header{}
			}

			// just pass reply info for some usage
			infRes.Header.Set(KindReply, res.Subject)
		}

		// ccx - none-wrapped context
		c.infiltrateResponse(ccx, infRes, err)
		return err
	})(ctx, m)

	return res, err
}

// RequestWithContext will send a request payload and deliver the response message,
// or an error, including a timeout if no message was received properly.
func (c *CommonPublish) RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error) {
	return c.RequestMsgWithContext(ctx, &nats.Msg{Subject: subj, Data: data})
}

// infiltrateResponse just register
func (c *CommonPublish) infiltrateResponse(ctx context.Context, msg *nats.Msg, err error) {
	ctx = WrapKindOfContext(ctx, KindRespond)

	_ = c.interceptor(func(ctx context.Context, _ *nats.Msg) error {
		return err
	})(ctx, msg)
}
