# NATS module

## Overview
Observability stack for NATS microservice services. 
Because of NATS not supported any middleware we was forced to reinvent by ourselves, and we had been chosen just 
wrap `nats.Conn` as one simplest way to do it. 

Unfortunately, we were not able to perform composition with original `nats.Con` structure, and we found that it could bring 
users in confusions since not full functions are covers.

Middleware concept allowed to decompose business logic in small mw components. 
Thus, we offer replacement of `MsgHandler` with context and error return.

## How to start

```bash
go get github.com/tel-io/instrumentation/middleware/nats/v2@latest 
```

```go
// create connection to nats
con, _ := nats.Connect(addr)

// wrap it 
mw := natsmw.New(con, natsmw.WithTel(t))
```

## Features
* Decorated instance has near legacy signature
* Build-IN Trace, Logs, Metrics, Recovery middlewares
* NATS Core fully supported functionality: async sub, pub, request, reply
* NATS JetStream: partial support
* Grafana Dashboard covered sub/pub,request,reply
* `*nats.Subscription` all wrapped function return attached subscription watcher who scrap metrics

### Consumer
#### NATS Core
SubscribeMW and QueueSubscribeMW function just backport compatibility for our previous version where signature was:
```go
type PostFn func(ctx context.Context, sub string, data []byte) ([]byte, error)
```

* Subscribe
* QueueSubscribe

NOTE: 

Example:
```go
func main(){
    // create nats connection
    con, _ := nats.Connect(addr)
    // create our mw from nats connection
    mw := natsmw.New(con, natsmw.WithTel(t))
    // we have handler already used our new-brand handler
    ourHandler := func(ctx context.Context, msg *nats.Msg) error {
        // perform respond with possible error returning
        return msg.Respond([]byte("HELLO"))
    }
	
    // subscribe to queue via our middleware with our `ourHandler`
    subscribe, _ := mw.QueueSubscribe("nats.demo", "consumer", ourHandler)
	
	// or without queue
    subscribe, _ := mw.Subscribe("nats.demo",  ourHandler)
}
```
#### JetStream
You should create JetStream from our wrapper
For subscription, we covered `Subscribe` and `QueueSubscribe` as `push` stack, which quite less popular as it provide not optimized for horizontal scale

Here is example:
```go
func main(){
	// create nats connection
    con, _ := nats.Connect(addr)
    // create our mw from nats connection
    mw := natsmw.New(con, natsmw.WithTel(t))
	
    // we have handler already used our new-brand handler
    ourHandler := func(ctx context.Context, msg *nats.Msg) error {
        _ = msg.Ack()
        return nil
    }

	// create wrapped js instance
    js, _ := mw.JetStream()
	
    // simple subscription with our handler to js
    handler := js.Subscribe("nats.demo",  ourHandler)
	
	// subscription with queue to js
    handler := js.QueueSubscribe("nats.demo",  "consumer", ourHandler)
}
```

#### BuildWrappedHandler
`BuildWrappedHandler` feature allow wrap any native function with middleware stack, allow to build middleware handler for function which not covered.

Here is example with jetstream:
```go
func main(){
    // some context, it could be signal closer or just stub
    ccx := context.TODO()
    
    // create tel instance with ccx provided ccx context
    t, closer := tel.New(ccx, tel.DefaultDebugConfig())
    defer closer()

	// create nats connection
    con, _ := nats.Connect(addr)
    // create our mw from nats connection
    mw := natsmw.New(con, natsmw.WithTel(t))
	
    // we have handler already used our new-brand handler
    ourHandler := func(ctx context.Context, msg *nats.Msg) error {
        _ = msg.Ack()
        return nil
    }
    
    // create wrapped js instance
    js, _ := mw.JetStream()

    // but PullSubscribe don't process any handler, but we would like observe this process
    sub, _ := js.PullSubscribe("stream.demo", "PULL")
	
    // so we just create function which func(msg *nats.Msg) but would be processed with our `ourHandler` processor
    handler := mw.BuildWrappedHandler(ourHandler)

    for{
        msgs, _ := v.Fetch(100)
        for _, msg := range msgs {
            //and here we process it
            handler(msg)
    }
}
```
### Producer
All our produced function receive ctx and name has `WithContext` suffix. 

WARNING! Context should contain `tel` context, as most important feature is to continue span of traces

#### NATS Core
Here is example:
```go
func main(){
	// some context, it could be signal closer or just stub
	ccx := context.TODO()
	
	// create tel instance with ccx provided ccx context
    t, closer := tel.New(ccx, tel.DefaultDebugConfig())
    defer closer()

	// wrap ccx with context
    ctx := tel.WithContext(ccx, t)
	
	// create nats connection
    con, _ := nats.Connect(addr)
	
    // create our mw from nats connection
    mw := natsmw.New(con, natsmw.WithTel(t))
	
	// send just with subject and body
    _ = con.PublishWithContext(ctx, "nats.test", []byte("HELLO_WORLD"))
	
	// or with native nats.Msg 
    _ = con.PublishMsgWithContext(ctx, &nats.Msg{Subject:"nats.test", Data: []byte("HELLO_WORLD"})	
	
	// or none-blocking request - assume that reply would be provided further
	_ = con.PublishRequestWithContext(ctx, "nats.test", "nats.reply", []byte("HELLO_WORLD"))

	// request with reply
    reply, _ = con.RequestWithContext(ctx, "nats.test",  []byte("HELLO_WORLD"))

    // request with reply via nats.Msg
    reply, _ = con.RequestWithContext(ctx, &nats.Msg{Subject:"nats.test", Data: []byte("HELLO_WORLD"})
}
```
#### JetStream
There is no helper yet