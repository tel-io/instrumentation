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
	nConn, _ := WrapConn(conn, WithTel(tele))

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
	nConn.Subscribe("sub", cb)
	nConn.QueueSubscribe("sub", "xxx", cb)

	// pub
	nConn.PublishWithContext(ctx, "sub", []byte("HELLO"))
	nConn.PublishMsgWithContext(ctx, &nats.Msg{})
	nConn.PublishRequestWithContext(ctx, "sub", "reply", []byte("HELLO"))
	nConn.RequestWithContext(ctx, "sub", []byte("HELLO"))
	nConn.RequestMsgWithContext(ctx, &nats.Msg{})
}
