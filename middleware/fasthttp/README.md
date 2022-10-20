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

## WARNING!
This experimental mw just show possibility. Under hood, it uses huge overhead.
