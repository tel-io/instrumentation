package nats

import "go.opentelemetry.io/otel/attribute"

// Attribute keys that can be added to a span.
const (
	Subject = attribute.Key("subject")
	IsError = attribute.Key("error")
	Kind    = attribute.Key("kind_of")
)

const (
	KindSub     = "sub"
	KindPub     = "pub"
	KindRequest = "request"
	KindRespond = "respond"
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
