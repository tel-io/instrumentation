package nats

import (
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
)

// ExtractAttributes ...
// @additional - handle business cases
func ExtractAttributes(msg *nats.Msg, kind string, additional bool) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		Subject.String(msg.Subject),
		Kind.String(kind),
	}

	if msg.Reply != "" {
		attrs = append(attrs, Reply.String(msg.Reply))
	} else if val := msg.Header.Get(KindReply); val != "" {
		attrs = append(attrs, Reply.String(val))
	}

	return attrs
}
