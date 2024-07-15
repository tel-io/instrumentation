package nats

import (
	"context"

	"go.opentelemetry.io/otel/baggage"
)

// WrapKindOfContext create baggage which update context subMiddleware related info about kind of event
// return ctx contained baggage with KindKey
// or just return none touched source ctx
func WrapKindOfContext(ctx context.Context, kindOf string) context.Context {
	if member, e := baggage.NewMember(KindKey, kindOf); e == nil {
		if b, ee := baggage.New(member); ee == nil {
			return baggage.ContextWithBaggage(ctx, b)
		}
	}

	return ctx
}
