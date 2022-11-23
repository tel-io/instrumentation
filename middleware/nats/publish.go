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

type PubMiddleware interface {
	Publish

	apply(in PubMiddleware) PubMiddleware
}

type CommonPublish struct {
	*nats.Conn
}

func (c *CommonPublish) apply(_ PubMiddleware) PubMiddleware {
	return c
}

func NewCommonPublish(conn *nats.Conn) *CommonPublish {
	return &CommonPublish{Conn: conn}
}

// PublishWithContext publishes the data argument to the given subject. The data
// argument is left untouched and needs to be correctly interpreted on
// the receiver.
func (c *CommonPublish) PublishWithContext(ctx context.Context, subj string, data []byte) error {
	return c.Conn.Publish(subj, data)
}

// PublishMsgWithContext publishes the Msg structure, which includes the
// Subject, an optional Reply and an optional Data field.
func (c *CommonPublish) PublishMsgWithContext(ctx context.Context, msg *nats.Msg) error {
	return c.Conn.PublishMsg(msg)
}

// PublishRequestWithContext will perform a Publish() expecting a response on the
// reply subject. Use Request() for automatically waiting for a response
// inline.
func (c *CommonPublish) PublishRequestWithContext(_ context.Context, subj, reply string, data []byte) error {
	return c.Conn.PublishRequest(subj, reply, data)
}

// RequestWithContext will send a request payload and deliver the response message,
// or an error, including a timeout if no message was received properly.
func (c *CommonPublish) RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error) {
	return c.Conn.RequestWithContext(ctx, subj, data)
}

// RequestMsgWithContext takes a context, a subject and payload
// in bytes and request expecting a single response.
func (c *CommonPublish) RequestMsgWithContext(ctx context.Context, msg *nats.Msg) (*nats.Msg, error) {
	return c.Conn.RequestMsgWithContext(ctx, msg)
}
