package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type metrics struct {
	counters       map[string]metric.Int64Counter
	valueRecorders map[string]metric.Float64Histogram
}

func createMeasures(tele tel.Telemetry, meter metric.Meter) *metrics {
	counters := make(map[string]metric.Int64Counter)
	valueRecorders := make(map[string]metric.Float64Histogram)

	counter, err := meter.Int64Counter(Count)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", Count))
	}

	requestBytesCounter, err := meter.Int64Counter(ContentLength)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", ContentLength))
	}

	serverLatencyMeasure, err := meter.Float64Histogram(Latency)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", Latency))
	}

	counters[Count] = counter
	counters[ContentLength] = requestBytesCounter
	valueRecorders[Latency] = serverLatencyMeasure

	return &metrics{
		counters:       counters,
		valueRecorders: valueRecorders,
	}
}

// SubMetrics implement Middleware interface
type SubMetrics struct {
	*metrics
}

func NewMetrics(m *metrics) *SubMetrics {
	return &SubMetrics{metrics: m}
}

func (t *SubMetrics) apply(next MsgHandler) MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) (err error) {
		defer func(start time.Time) {
			kind := extractBaggageKind(ctx)

			if ctx.Err() != nil {
				err = ctx.Err()
				ctx = tel.FromCtx(ctx).Ctx()
			}

			attr := []attribute.KeyValue{
				IsError.Bool(err != nil),
				Subject.String(decreaseSubjectCardinality(msg.Subject)),
				Kind.String(kind),
			}

			t.counters[Count].Add(ctx, 1, metric.WithAttributes(attr...))
			t.counters[ContentLength].Add(ctx, int64(len(msg.Data)), metric.WithAttributes(attr...))
			t.valueRecorders[Latency].Record(ctx, float64(time.Since(start).Milliseconds()), metric.WithAttributes(attr...))

		}(time.Now())

		return next(ctx, msg)
	}
}
