package nats

import (
	"context"
)

// PostFn callback function which got new instance of tele inside ctx
// and msg sub + data
// Deprecated: legacy function, but we use it via conn wrapper: QueueSubscribeMW or SubscribeMW just for backport compatibility
type PostFn func(ctx context.Context, sub string, data []byte) ([]byte, error)
