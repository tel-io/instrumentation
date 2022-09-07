# GRPC module

## How to use

```bash
go get github.com/tel-io/instrumentation/middleware/grpc@latest
```

### Client


```go
import (
	//...
    "github.com/d7561985/tel/v2"
    grpcx "github.com/tel-io/instrumentation/middleware/grpc"
    //...
)

func NewConn(addr string)  {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(
		insecure.NewCredentials()),
		// for unary use tel module
		grpc.WithChainUnaryInterceptor(grpcx.UnaryClientInterceptorAll()),
		// for stream use stand-alone trace + metrics no recover
		grpc.WithChainStreamInterceptor(grpcx.StreamClientInterceptor(
			grpcx.WithMetricOption(
				otelgrpc.WithServerHandledHistogram(true),
				otelgrpc.WithConstLabels(
					attribute.String("xxx", "example"),
					attribute.String("yyy", "client"),
				)),
		)),
		grpc.WithBlock(),
	)

    //...
}
```

### Server

```go
    import (

    "github.com/d7561985/tel/v2"
    grpcx "github.com/tel-io/instrumentation/middleware/grpc"
)

func Start(ctx context.Context, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.WithMessagef(err, "failed to listen: %v", err)
	}

	otmetr := []otelgrpc.Option{
		otelgrpc.WithServerHandledHistogram(true),
		otelgrpc.WithConstLabels(
			attribute.String("userID", "e64916d9-bfd0-4f79-8ee3-847f2d034d20"),
			attribute.String("xxx", "example"),
			attribute.String("yyy", "server"),
		),
	}

	s := grpc.NewServer(
		// for unary use tel module
		grpc.ChainUnaryInterceptor(grpcx.UnaryServerInterceptorAll(
			grpcx.WithTel(tel.FromCtx(ctx)),
			grpcx.WithMetricOption(otmetr...),
		)),
		// for stream use stand-alone trace + metrics no recover
		grpc.ChainStreamInterceptor(grpcx.StreamServerInterceptor(grpcx.WithTel(tel.FromCtx(ctx)),
			grpcx.WithMetricOption(otmetr...),
		)),
	)

	api.RegisterHelloServiceServer(s, &server{})

	go func() {
		<-ctx.Done()
		tel.FromCtx(ctx).Info("grpc down")

		s.Stop()
	}()

	return errors.WithStack(s.Serve(lis))
}
```