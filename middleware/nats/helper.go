package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/tel-io/instrumentation/middleware/nats/v2/natsprop"
)

// ReplyFn reply helper which send reply with wrapping trace information
func ReplyFn(ctx context.Context, msg *nats.Msg, data []byte) error {
	if msg.Reply == "" {
		return nil
	}

	resMsg := &nats.Msg{Data: data}
	natsprop.Inject(ctx, msg)

	return errors.WithStack(msg.RespondMsg(resMsg))
}
