package nats

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

type metrics struct {
	counters       map[string]syncint64.Counter
	valueRecorders map[string]syncfloat64.Histogram
}

func createMeasures(tele tel.Telemetry, meter metric.Meter) *metrics {
	counters := make(map[string]syncint64.Counter)
	valueRecorders := make(map[string]syncfloat64.Histogram)

	counter, err := meter.SyncInt64().Counter(Count)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", Count))
	}

	requestBytesCounter, err := meter.SyncInt64().Counter(ContentLength)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", ContentLength))
	}

	serverLatencyMeasure, err := meter.SyncFloat64().Histogram(Latency)
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
				Subject.String(replacers.Apply(msg.Subject)),
				Kind.String(kind),
			}

			t.counters[Count].Add(ctx, 1, attr...)
			t.counters[ContentLength].Add(ctx, int64(len(msg.Data)), attr...)
			t.valueRecorders[Latency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)

		}(time.Now())

		return next(ctx, msg)
	}
}
