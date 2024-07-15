package nats

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
)

const (
	KindKey    = "kind_of"
	PayloadKey = "payload"
)

// Attribute keys that can be added to a span.
const (
	Subject  = attribute.Key("subject")
	Reply    = attribute.Key("reply")
	IsError  = attribute.Key("error")
	Kind     = attribute.Key(KindKey)
	Duration = attribute.Key("duration")
)

const (
	KindUnk     = "UNK"
	KindSub     = "SUB"
	KindPub     = "PUB"
	KindRequest = "REQUEST"
	KindRespond = "RESPOND"
	KindReply   = "REPLY"
)

// Server NATS metrics
const (
	Count         = "nats.count"          // Incoming request count total
	ContentLength = "nats.content_length" // Incoming request bytes total
	Latency       = "nats.duration"       // Incoming end to end duration, microseconds

	SubscriptionsPendingCount = "nats.subscriptions.pending.msgs"
	SubscriptionsPendingBytes = "nats.subscriptions.pending.bytes"
	SubscriptionsDroppedMsgs  = "nats.subscriptions.dropped.count"
	SubscriptionCountMsgs     = "nats.subscriptions.send.count"
)

func extractBaggageKind(ctx context.Context) string {
	b := baggage.FromContext(ctx)
	if v := b.Member(KindKey).Value(); v != "" {
		return v
	}

	return KindUnk
}
