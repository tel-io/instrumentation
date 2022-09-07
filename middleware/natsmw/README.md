# NATS module

## How to start

```bash
go get github.com/tel-io/instrumentation/middleware/natsmw 
```

### Consumer

```go
import (
    "github.com/d7561985/tel/v2"
    "github.com/tel-io/instrumentation/middleware/natsmw"
)

func sub(){
    mw := natsmw.New(natsmw.WithTel(t))
	
    subscribe, err := con.Subscribe("nats.demo", mw.Handler(func (ctx context.Context, sub string, data []byte) ([]byte, error) {
    // send as reply
        return nil, []byte("HELLO WORLD")
    }))
    
    crash, err := con.Subscribe("nats.crash", mw.Handler(func (ctx context.Context, sub string, data []byte) ([]byte, error) {
        time.Sleep(time.Millisecond)
        panic("some panic")
        return nil, nil
    }))
    
    someErr, err := con.Subscribe("nats.err", mw.Handler(func (ctx context.Context, sub string, data []byte) ([]byte, error) {
        time.Sleep(time.Millisecond)
    return nil, fmt.Errorf("some errro")
    }))
}
```

### Producer
`ToDo`