# HTTP module

## How to use
```bash
$ go get github.com/tel-io/instrumentation/middleware/http@latest
```

### Client

```go
import (
    "github.com/d7561985/tel/v2"
	mw "github.com/tel-io/instrumentation/middleware/http"
)

func main(){
    t, closer := tel.New(context.Background(), tel.GetConfigFromEnv())
    defer closer()

	// create wrapped client
    client:= mw.NewClient(nil)

	ctx, span := t.StartSpan(context.Background(), path, trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
    defer span.End()

	// create context based request
    req, _ := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	//
    res, err := client.Do(req)
}
```

### Server

```go
import (
    mw "github.com/tel-io/instrumentation/middleware/http"
)

func (s *Server) Start(ctx context.Context) (err error) {
	m := mw.ServerMiddlewareAll()

	mx := http.NewServeMux()
	mx.Handle("/hello", m(http.HandlerFunc(s.helloHttp)))
	mx.Handle("/crash", m(http.HandlerFunc(s.crashHttp)))
	mx.Handle("/error", m(http.HandlerFunc(s.errorHttp)))

	srv := &http.Server{}
	srv.Handler = mx
}
```