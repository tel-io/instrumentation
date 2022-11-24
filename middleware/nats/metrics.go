package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/asyncint64"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	"sync"
	"time"
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
			if ctx.Err() != nil {
				err = ctx.Err()
				ctx = tel.FromCtx(ctx).Ctx()
			}

			attr := []attribute.KeyValue{
				IsError.Bool(err != nil),
				Subject.String(msg.Subject),
				Kind.String(KindSub),
			}

			t.counters[Count].Add(ctx, 1, attr...)
			t.counters[ContentLength].Add(ctx, int64(len(msg.Data)), attr...)
			t.valueRecorders[Latency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)

		}(time.Now())

		return next(ctx, msg)
	}
}

// PubMetric handle publish and request metrics gathering implementing PubMiddleware
type PubMetric struct {
	*metrics

	root Publish
}

var _ PubMiddleware = &PubMetric{}

func NewPubMetric(m *metrics) *PubMetric {

	return &PubMetric{
		metrics: m,
	}
}

func (p *PubMetric) apply(in PubMiddleware) PubMiddleware {
	p.root = in

	return p
}

func (p *PubMetric) PublishWithContext(ctx context.Context, subj string, data []byte) (err error) {
	defer func(start time.Time) {
		if ctx.Err() != nil {
			err = ctx.Err()
			ctx = tel.FromCtx(ctx).Ctx()
		}

		attr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(subj),
			Kind.String(KindPub),
		}

		p.counters[Count].Add(ctx, 1, attr...)
		p.counters[ContentLength].Add(ctx, int64(len(data)), attr...)
		p.valueRecorders[Latency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)
	}(time.Now())

	return p.root.PublishWithContext(ctx, subj, data)
}

func (p *PubMetric) PublishMsgWithContext(ctx context.Context, msg *nats.Msg) (err error) {
	defer func(start time.Time) {
		if ctx.Err() != nil {
			err = ctx.Err()
			ctx = tel.FromCtx(ctx).Ctx()
		}

		attr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(msg.Subject),
			Kind.String(KindPub),
		}

		p.counters[Count].Add(ctx, 1, attr...)
		p.counters[ContentLength].Add(ctx, int64(len(msg.Data)), attr...)
		p.valueRecorders[Latency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)
	}(time.Now())

	return p.root.PublishMsgWithContext(ctx, msg)
}

func (p *PubMetric) PublishRequestWithContext(ctx context.Context, subj, reply string, data []byte) (err error) {
	defer func(start time.Time) {
		if ctx.Err() != nil {
			err = ctx.Err()
			ctx = tel.FromCtx(ctx).Ctx()
		}

		attr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(subj),
			Kind.String(KindPub),
		}

		p.counters[Count].Add(ctx, 1, attr...)
		p.counters[ContentLength].Add(ctx, int64(len(data)), attr...)
		p.valueRecorders[Latency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)
	}(time.Now())

	return p.root.PublishRequestWithContext(ctx, subj, reply, data)
}

func (p *PubMetric) RequestWithContext(ctx context.Context, subj string, data []byte) (resp *nats.Msg, err error) {
	msg := &nats.Msg{Subject: subj, Data: data}

	return p.RequestMsgWithContext(ctx, msg)
}

func (p *PubMetric) RequestMsgWithContext(ctx context.Context, msg *nats.Msg) (resp *nats.Msg, err error) {
	defer func(start time.Time) {
		if ctx.Err() != nil {
			err = ctx.Err()
			ctx = tel.FromCtx(ctx).Ctx()
		}

		reqAttr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(msg.Subject),
			Kind.String(KindRequest),
		}

		p.counters[Count].Add(ctx, 1, reqAttr...)
		p.counters[ContentLength].Add(ctx, int64(len(msg.Data)), reqAttr...)
		p.valueRecorders[Latency].Record(ctx, float64(time.Since(start).Milliseconds()), reqAttr...)

		if resp != nil {
			resAttr := []attribute.KeyValue{
				IsError.Bool(err != nil),
				Subject.String(msg.Subject),
				Kind.String(KindRespond),
			}

			p.counters[Count].Add(ctx, int64(len(resp.Data)), resAttr...)
			p.counters[ContentLength].Add(ctx, int64(len(resp.Data)), resAttr...)
		}
	}(time.Now())

	return p.root.RequestMsgWithContext(ctx, msg)
}

// SubscriptionStatMetric hook provide important subscription statistics
type SubscriptionStatMetric struct {
	list     sync.Map
	counters map[string]asyncint64.Gauge
}

func NewSubscriptionStatMetrics(opts ...Option) (*SubscriptionStatMetric, error) {
	cfg := newConfig(opts)

	c := make(map[string]asyncint64.Gauge)
	msgs, _ := cfg.meter.AsyncInt64().Gauge(SubscriptionsPendingCount)
	bs, _ := cfg.meter.AsyncInt64().Gauge(SubscriptionsPendingBytes,
		instrument.WithUnit(unit.Bytes))

	dd, _ := cfg.meter.AsyncInt64().Gauge(SubscriptionsDroppedMsgs)
	cc, _ := cfg.meter.AsyncInt64().Gauge(SubscriptionCountMsgs)

	c[SubscriptionsPendingCount] = msgs
	c[SubscriptionsPendingBytes] = bs
	c[SubscriptionsDroppedMsgs] = dd
	c[SubscriptionCountMsgs] = cc

	res := &SubscriptionStatMetric{
		counters: c,
	}

	err := cfg.meter.RegisterCallback([]instrument.Asynchronous{msgs, bs, dd, cc}, res.callback)
	if err != nil {
		return nil, errors.WithMessagef(err, "reggister callback")
	}

	return res, nil
}

func (s *SubscriptionStatMetric) Hook(sub *nats.Subscription, err error) (*nats.Subscription, error) {
	if err != nil {
		return nil, err
	}

	s.Register(sub)

	return sub, nil
}

func (s *SubscriptionStatMetric) Register(sub ...*nats.Subscription) {
	for _, v := range sub {
		s.list.Store(v, v)
	}
}

func (s *SubscriptionStatMetric) callback(ctx context.Context) {
	// we could have multi subscriptions with the same subject
	// we should set total number of that
	data := make(map[string]struct {
		msgs    int64
		bytes   int64
		dropped int64
		count   int64
	})

	s.list.Range(func(key, value interface{}) bool {
		v, ok := value.(*nats.Subscription)
		if !ok {
			return true
		}

		pMsg, pBytes, _ := v.Pending()
		dropped, _ := v.Dropped()
		count, _ := v.Delivered()

		vc := data[v.Subject]
		vc.msgs += int64(pMsg)
		vc.bytes += int64(pBytes)
		vc.dropped += int64(dropped)
		vc.count += count

		data[v.Subject] = vc
		return true
	})

	for k, v := range data {
		s.counters[SubscriptionsPendingCount].Observe(ctx, v.msgs, Subject.String(k))
		s.counters[SubscriptionsPendingBytes].Observe(ctx, v.bytes, Subject.String(k))
		s.counters[SubscriptionsDroppedMsgs].Observe(ctx, v.dropped, Subject.String(k))
		s.counters[SubscriptionCountMsgs].Observe(ctx, v.count, Subject.String(k))
	}
}
