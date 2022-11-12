module github.com/tel-io/instrumentation/module/otelgrpc/example

go 1.17

replace github.com/tel-io/instrumentation/module/otelgrpc => ../

require (
	github.com/golang/protobuf v1.5.2
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.32.0
	go.opentelemetry.io/otel v1.11.1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.11.1
	go.opentelemetry.io/otel/sdk v1.11.1
	go.opentelemetry.io/otel/sdk/metric v0.33.0
	go.opentelemetry.io/otel/trace v1.11.1
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	google.golang.org/grpc v1.50.1
)

require (
	github.com/tel-io/instrumentation/module/otelgrpc v0.0.0-00010101000000-000000000000
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v0.33.0
	go.opentelemetry.io/otel/metric v0.33.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/tel-io/otelgrpc v1.0.1 // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20220919091848-fb04ddd9f9c8 // indirect
	golang.org/x/text v0.3.5 // indirect
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)
