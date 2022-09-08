# chi module

## How to 

```bash
$ go get github.com/tel-io/instrumentation/middleware/echo@latest
```

### Example
```go
import (
    emw "github.com/tel-io/instrumentation/middleware/echo"
    mw "github.com/tel-io/instrumentation/middleware/http"
)

func main(){
    tele, closer := tel.New(context.Background(), cfg)
    defer closer()
	
    
    app := echo.New()
    app.Use(emw.HTTPServerMiddlewareAll(mw.WithTel(&tele)))
}
```