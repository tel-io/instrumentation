package nats

import "go.opentelemetry.io/otel/attribute"

// Attribute keys that can be added to a span.
const (
	Subject = attribute.Key("nats.subject")
	IsError = attribute.Key("nats.code")
)

// Server NATS metrics
const (
	SubCount         = "nats.consumer.request_count"          // Incoming request count total
	SubContentLength = "nats.consumer.request_content_length" // Incoming request bytes total
	SubLatency       = "nats.consumer.duration"               // Incoming end to end duration, microseconds

	OutLatency       = "nats.out.duration"
	OutCount         = "nats.out.count"          // Outcome publish count
	OutContentLength = "nats.out.content_length" // Outcome content bytes total

	RequestRespondContentLength = "nats.request.respond.content_length"

	SubscriptionsPendingCount = "nats.subscriptions.pending.msgs"
	SubscriptionsPendingBytes = "nats.subscriptions.pending.bytes"
	SubscriptionsDroppedMsgs  = "nats.subscriptions.dropped.count"
	SubscriptionCountMsgs     = "nats.subscriptions.send.count"
)
