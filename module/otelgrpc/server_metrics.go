package otelgrpc

import (
	"context"

	"github.com/tel-io/instrumentation/module/otelgrpc/packages/grpcstatus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
)

const (
	serverStartedCounter    = "grpc_server_started_total"
	serverHandledCounter    = "grpc_server_handled_total"
	serverStreamMsgReceived = "grpc_server_msg_received_total"
	serverStreamMsgSent     = "grpc_server_msg_received_total"
	serverHandledHistogram  = "grpc_server_handling_seconds"
)

const (
	AttrType    = "grpc_type"
	AttrService = "grpc_service"
	AttrMethod  = "grpc_method"
	AttrCode    = "grpc_code"
)

// ServerMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a gRPC server.
type ServerMetrics struct {
	meter                         metric.Meter
	labels                        []attribute.KeyValue
	bucket                        []float64
	serverHandledHistogramEnabled bool

	counters       map[string]metric.Int64Counter
	valueRecorders map[string]metric.Float64Histogram
}

// NewServerMetrics returns a ServerMetrics object. Use a new instance of
// ServerMetrics when not using the default Prometheus metrics registry, for
// example when wanting to control which metrics are added to a registry as
// opposed to automatically adding metrics via init functions.
func NewServerMetrics(counterOpts ...Option) *ServerMetrics {
	s := &ServerMetrics{}

	c := newConfig(counterOpts...)
	s.configure(c)
	s.createMeasures()

	return s
}

func (m *ServerMetrics) configure(c *config) {
	m.meter = c.Meter
	m.labels = c.Labels
	m.bucket = c.Bucket

	m.serverHandledHistogramEnabled = true
}

func (m *ServerMetrics) createMeasures() {
	m.counters = make(map[string]metric.Int64Counter)
	m.valueRecorders = make(map[string]metric.Float64Histogram)

	m.counters[serverStartedCounter] = MustCounter(m.meter.Int64Counter(serverStartedCounter,
		metric.WithDescription("Total number of RPCs started on the server."),
		metric.WithUnit("1"),
	))

	m.counters[serverHandledCounter] = MustCounter(m.meter.Int64Counter(serverHandledCounter,
		metric.WithDescription("Total number of RPCs completed on the server, regardless of success or failure."),
		metric.WithUnit("1"),
	))

	m.counters[serverStreamMsgReceived] = MustCounter(m.meter.Int64Counter(serverStreamMsgReceived,
		metric.WithDescription("Total number of RPC stream messages received on the server."),
		metric.WithUnit("1"),
	))

	m.counters[serverStreamMsgSent] = MustCounter(m.meter.Int64Counter(serverStreamMsgSent,
		metric.WithDescription("Total number of gRPC stream messages sent by the server."),
		metric.WithUnit("1"),
	))

	if m.serverHandledHistogramEnabled {
		m.valueRecorders[serverHandledHistogram] = MustHistogram(m.meter.Float64Histogram(serverHandledHistogram,
			metric.WithDescription("Histogram of response latency (milliseconds) of gRPC that had been application-level handled by the server."),
			metric.WithUnit("ms"), // seconds
		))
	}
}

func handleErr(err error) {
	if err != nil {
		otel.Handle(err)
	}
}

// UnaryServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *ServerMetrics) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		monitor := newServerReporter(ctx, m, Unary, info.FullMethod)
		monitor.ReceivedMessage(ctx)
		resp, err := handler(ctx, req)
		st, _ := grpcstatus.FromError(err)
		monitor.Handled(ctx, st.Code())
		if err == nil {
			monitor.SentMessage(ctx)
		}
		return resp, err
	}
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (m *ServerMetrics) StreamServerInterceptor() func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		monitor := newServerReporter(ss.Context(), m, streamRPCType(info), info.FullMethod)
		err := handler(srv, &monitoredServerStream{ss, monitor})
		st, _ := grpcstatus.FromError(err)
		monitor.Handled(ss.Context(), st.Code())
		return err
	}
}

func streamRPCType(info *grpc.StreamServerInfo) grpcType {
	if info.IsClientStream && !info.IsServerStream {
		return ClientStream
	} else if !info.IsClientStream && info.IsServerStream {
		return ServerStream
	}
	return BidiStream
}

// monitoredStream wraps grpc.ServerStream allowing each Sent/Recv of message to increment counters.
type monitoredServerStream struct {
	grpc.ServerStream
	monitor *serverReporter
}

func (s *monitoredServerStream) SendMsg(m interface{}) error {
	if err := s.ServerStream.SendMsg(m); err != nil {
		return err
	}

	s.monitor.SentMessage(s.ServerStream.Context())

	return nil
}

func (s *monitoredServerStream) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if err != nil {
		return err
	}

	s.monitor.ReceivedMessage(s.ServerStream.Context())

	return nil
}

func MustCounter(v metric.Int64Counter, err error) metric.Int64Counter {
	if err != nil {
		handleErr(err)
	}

	return v
}

func MustHistogram(v metric.Float64Histogram, err error) metric.Float64Histogram {
	if err != nil {
		handleErr(err)
	}

	return v
}
