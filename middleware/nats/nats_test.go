package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/tel-io/tel/v2"
)

func Example_handler() {
	tele := tel.NewNull()
	ctx := tele.Ctx()

	conn, _ := nats.Connect("example.com")
	nConn := New(WithTel(tele)).Use(conn)

	// legacy backport
	cbLegacy := func(ctx context.Context, sub string, data []byte) ([]byte, error) {
		return nil, nil
	}

	cb := func(ctx context.Context, msg *nats.Msg) error {
		return nil
	}

	_, _ = nConn.QueueSubscribeMW("sub", "queue", cbLegacy)
	_, _ = nConn.QueueSubscribeMW("sub2", "queue", cbLegacy)

	// sub
	_, _ = nConn.Subscribe("sub", cb)
	_, _ = nConn.QueueSubscribe("sub", "xxx", cb)

	// sync sub with wrap
	ourHandler := nConn.BuildWrappedHandler(func(ctx context.Context, msg *nats.Msg) error {
		// perform respond with possible error returning
		return msg.Respond([]byte("HELLO"))
	})
	ch := make(chan *nats.Msg)
	_, _ = nConn.QueueSubscribeSyncWithChan("sub", "queue", ch)
	for msg := range ch {
		ourHandler(msg)
	}

	// pub
	_ = nConn.PublishWithContext(ctx, "sub", []byte("HELLO"))
	_ = nConn.PublishMsgWithContext(ctx, &nats.Msg{})
	_ = nConn.PublishRequestWithContext(ctx, "sub", "reply", []byte("HELLO"))
	_, _ = nConn.RequestWithContext(ctx, "sub", []byte("HELLO"))
	_, _ = nConn.RequestMsgWithContext(ctx, &nats.Msg{})
}
