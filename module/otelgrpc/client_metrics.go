package otelgrpc

import (
	"context"
	"io"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClientMetrics represents a collection of metrics to be registered on a
// Prometheus metrics registry for a gRPC client.
type ClientMetrics struct {
	meter                         metric.Meter
	labels                        []attribute.KeyValue
	clientHandledHistogramEnabled bool

	counters       map[string]metric.Int64Counter
	valueRecorders map[string]metric.Float64Histogram
}

const (
	clientStartedCounter    = "grpc_client_started_total"
	clientHandledCounter    = "grpc_client_handled_total"
	clientStreamMsgReceived = "grpc_client_msg_received_total"
	clientStreamMsgSent     = "grpc_client_msg_sent_total"

	clientHandledHistogram    = "grpc_client_handling_seconds"
	clientStreamRecvHistogram = "grpc_client_msg_recv_handling_seconds"
	clientStreamSendHistogram = "grpc_client_msg_send_handling_seconds"
)

// NewClientMetrics returns a ClientMetrics object. Use a new instance of
// ClientMetrics when not using the default Prometheus metrics registry, for
// example when wanting to control which metrics are added to a registry as
// opposed to automatically adding metrics via init functions.
func NewClientMetrics(counterOpts ...Option) *ClientMetrics {
	m := &ClientMetrics{}

	c := newConfig(counterOpts...)
	m.configure(c)
	m.createMeasures()

	return m
}

func (m *ClientMetrics) configure(c *config) {
	m.meter = c.Meter
	m.labels = c.Labels

	m.clientHandledHistogramEnabled = true
}

func (m *ClientMetrics) createMeasures() {
	m.counters = make(map[string]metric.Int64Counter)
	m.valueRecorders = make(map[string]metric.Float64Histogram)

	// "grpc_type", "grpc_service", "grpc_method"
	m.counters[clientStartedCounter] = MustCounter(m.meter.Int64Counter(clientStartedCounter,
		metric.WithDescription("Total number of RPCs started on the client."),
		metric.WithUnit("1"),
	))

	// "grpc_type", "grpc_service", "grpc_method", "grpc_code"
	m.counters[clientHandledCounter] = MustCounter(m.meter.Int64Counter(clientHandledCounter,
		metric.WithDescription("Total number of RPCs completed by the client, regardless of success or failure."),
		metric.WithUnit("1"),
	))

	// "grpc_type", "grpc_service", "grpc_method"
	m.counters[clientStreamMsgReceived] = MustCounter(m.meter.Int64Counter(clientStreamMsgReceived,
		metric.WithDescription("Total number of RPC stream messages received by the client."),
		metric.WithUnit("1"),
	))

	// "grpc_type", "grpc_service", "grpc_method"
	m.counters[clientStreamMsgSent] = MustCounter(m.meter.Int64Counter(clientStreamMsgSent,
		metric.WithDescription("Total number of gRPC stream messages sent by the client."),
		metric.WithUnit("1"),
	))

	if !m.clientHandledHistogramEnabled {
		return
	}

	m.valueRecorders[clientHandledHistogram] = MustHistogram(m.meter.Float64Histogram(clientHandledHistogram,
		metric.WithDescription("Histogram of response latency (milliseconds) of the gRPC until it is finished by the application."),
		metric.WithUnit("ms"),
	))

	m.valueRecorders[clientStreamRecvHistogram] = MustHistogram(m.meter.Float64Histogram(clientStreamRecvHistogram,
		metric.WithDescription("Histogram of response latency (milliseconds) of the gRPC single message receive."),
		metric.WithUnit("ms"),
	))

	m.valueRecorders[clientStreamSendHistogram] = MustHistogram(m.meter.Float64Histogram(clientStreamSendHistogram,
		metric.WithDescription("Histogram of response latency (seconds) of the gRPC single message send."),
		metric.WithUnit("ms"),
	))
}

// UnaryClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Unary RPCs.
func (m *ClientMetrics) UnaryClientInterceptor() func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		monitor := newClientReporter(ctx, m, Unary, method)
		monitor.SentMessage(ctx)
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			monitor.ReceivedMessage(ctx)
		}
		st, _ := status.FromError(err)
		monitor.Handled(ctx, st.Code())
		return err
	}
}

// StreamClientInterceptor is a gRPC client-side interceptor that provides Prometheus monitoring for Streaming RPCs.
func (m *ClientMetrics) StreamClientInterceptor() func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		monitor := newClientReporter(ctx, m, clientStreamType(desc), method)
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			st, _ := status.FromError(err)
			monitor.Handled(ctx, st.Code())
			return nil, err
		}
		return &monitoredClientStream{clientStream, monitor}, nil
	}
}

func clientStreamType(desc *grpc.StreamDesc) grpcType {
	if desc.ClientStreams && !desc.ServerStreams {
		return ClientStream
	} else if !desc.ClientStreams && desc.ServerStreams {
		return ServerStream
	}
	return BidiStream
}

// monitoredClientStream wraps grpc.ClientStream allowing each Sent/Recv of message to increment counters.
type monitoredClientStream struct {
	grpc.ClientStream
	monitor *clientReporter
}

func (s *monitoredClientStream) SendMsg(m interface{}) error {
	start := time.Now()
	err := s.ClientStream.SendMsg(m)
	s.monitor.SendMessageTimer(s.Context(), start)

	if err == nil {
		s.monitor.SentMessage(s.Context())
	}
	return err
}

func (s *monitoredClientStream) RecvMsg(m interface{}) error {
	start := time.Now()
	err := s.ClientStream.RecvMsg(m)
	s.monitor.ReceiveMessageTimer(s.Context(), start)

	if err == nil {
		s.monitor.ReceivedMessage(s.Context())
	} else if err == io.EOF {
		s.monitor.Handled(s.Context(), codes.OK)
	} else {
		st, _ := status.FromError(err)
		s.monitor.Handled(s.Context(), st.Code())
	}
	return err
}
