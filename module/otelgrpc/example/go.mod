module github.com/tel-io/instrumentation/module/otelgrpc/example

go 1.21

toolchain go1.22.2

replace github.com/tel-io/instrumentation/module/otelgrpc => ../

require (
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.53.0
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/sdk/metric v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
	golang.org/x/net v0.27.0
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2 // indirect
)

require (
	github.com/golang/protobuf v1.5.4
	github.com/tel-io/instrumentation/module/otelgrpc v1.0.1
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.28.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/google/uuid v1.6.0 // indirect
	go.opentelemetry.io/otel/metric v1.28.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1 // indirect
)
