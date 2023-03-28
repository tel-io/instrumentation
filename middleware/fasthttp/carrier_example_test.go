package fasthttp_test

import (
	"github.com/tel-io/tel/v2"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"

	mw "github.com/tel-io/instrumentation/middleware/fasthttp"
)

func ExampleClient() {
	t := tel.NewNull()
	_, ctx := t.StartSpan(t.Ctx(), "DEMO_SPAN")

	propagators := otel.GetTextMapPropagator()

	cxt := t.Copy().Ctx()
	span, ctx := tel.FromCtx(cxt).StartSpan(cxt, "req")

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	// fill request with carrier information about trace and baggage
	propagators.Inject(ctx, mw.NewCarrier(&req.Header))

	c := fasthttp.Client{}
	if err := c.Do(req, res); err != nil {
		span.RecordError(err)
	}

	span.End()
}
