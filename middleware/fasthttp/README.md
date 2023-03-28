# FastHttp middleware
Provide simple middleware for fasthttp with context prapagation option

## GET
```bash
$ go get github.com/tel-io/instrumentation/middleware/fasthttp@latest
```

## Usage
Just wrap function with mw and context extractor `GetNativeContext` with tel instance and trace span


```go
import (
    mw "github.com/tel-io/instrumentation/middleware/fasthttp"
)
...
middleware := mw.ServerMiddleware()

// simple handler
if err := fasthttp.ListenAndServe(port, middleware(func(ctx *fasthttp.RequestCtx) {
    // EXTRACT OTEL CONTEXT WITH SPAN AND EVERYTHING.....
    ctn := mw.GetNativeContext(ctx)
})
```

### Client

#### Example

```go
package main

import (
	"context"

	mw "github.com/tel-io/instrumentation/middleware/fasthttp"
	"github.com/tel-io/tel/v2"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel"
)

func main() {
	t, cc := tel.New(context.Background(), tel.GetConfigFromEnv())
	defer cc()

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	ccx := t.Copy().Ctx()
	
	span, ctx := tel.FromCtx(ccx).StartSpan(ccx, "req")
	defer span.End()

	// fill request with carrier information about trace and baggage
	otel.GetTextMapPropagator().Inject(ctx, mw.NewCarrier(&req.Header))

	c := fasthttp.Client{}
	if err := c.Do(req, res); err != nil {
		span.RecordError(err)
	}
}

```
#### Propagator
Required composite map propagator: TraceContext and Baggage
```
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{},
		))
```

NOTE: tel library already uses it

## WARNING!
This experimental mw just show possibility. Under hood, it uses huge overhead.
