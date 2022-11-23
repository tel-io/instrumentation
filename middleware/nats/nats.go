package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
)

// PostFn callback function which got new instance of tele inside ctx
// and msg sub + data
// Deprecated: legacy function, but we use it via conn wrapper: QueueSubscribeMW or SubscribeMW just for backport compatibility
type PostFn func(ctx context.Context, sub string, data []byte) ([]byte, error)

func extractAttr(m *nats.Msg, isErr bool) []attribute.KeyValue {
	return []attribute.KeyValue{
		IsError.Bool(isErr),
		Subject.String(m.Subject),
	}
}
