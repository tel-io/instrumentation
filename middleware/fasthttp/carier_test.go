package fasthttp

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// use trace propagation
// due to NewCompositeTextMapPropagator it's possible send both of them
func TestTrace(t *testing.T) {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		))

	ctx := context.Background()
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: trace.TraceID{0x03},
		SpanID:  trace.SpanID{0x03},
	})
	ctx = trace.ContextWithRemoteSpanContext(ctx, sc)

	propagators := otel.GetTextMapPropagator()

	req := fasthttp.AcquireRequest()
	t.Run("inject", func(t *testing.T) {
		x := HeaderCarrier{&req.Header}
		propagators.Inject(ctx, x)
		assert.Equal(t, req.Header.Len(), 1)
	})

	// depend on inject
	t.Run("extract", func(t *testing.T) {
		x := HeaderCarrier{&req.Header}
		cxt := propagators.Extract(ctx, x)
		baggage := baggage.FromContext(cxt)
		spanContext := trace.SpanContextFromContext(cxt)

		assert.Equal(t, sc.SpanID(), spanContext.SpanID())
		assert.Equal(t, sc.TraceID(), spanContext.TraceID())

		fmt.Println(baggage)
	})
}
