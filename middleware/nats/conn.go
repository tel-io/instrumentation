package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
)

// ConnContext wrapper for nats.ConnContext aks mw connection approach
//
// Features:
// Expose subscription stats via function overwrite
type ConnContext struct {
	conn *nats.Conn

	Publish
	Subscriber

	*Core
}

// Conn unwrap connection
func (c *ConnContext) Conn() *nats.Conn {
	return c.conn
}

// JetStream returns a JetStreamContext wrapper for consumer
func (c *ConnContext) JetStream(opts ...nats.JSOpt) (*JetStreamContext, error) {
	js, err := c.conn.JetStream(opts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &JetStreamContext{
		js:   js,
		Core: c.Core,
	}, nil
}
