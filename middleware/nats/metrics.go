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

// SubMetrics implement Middleware interface
type SubMetrics struct {
	counters       map[string]syncint64.Counter
	valueRecorders map[string]syncfloat64.Histogram
}

func NewMetrics(tele tel.Telemetry, meter metric.Meter) *SubMetrics {
	m := &SubMetrics{}
	m.createMeasures(tele, meter)

	return m
}

func (t *SubMetrics) createMeasures(tele tel.Telemetry, meter metric.Meter) {
	t.counters = make(map[string]syncint64.Counter)
	t.valueRecorders = make(map[string]syncfloat64.Histogram)

	counter, err := meter.SyncInt64().Counter(SubCount)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", SubCount))
	}

	requestBytesCounter, err := meter.SyncInt64().Counter(SubContentLength)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", SubContentLength))
	}

	serverLatencyMeasure, err := meter.SyncFloat64().Histogram(SubLatency)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", SubLatency))
	}

	t.counters[SubCount] = counter
	t.counters[SubContentLength] = requestBytesCounter
	t.valueRecorders[SubLatency] = serverLatencyMeasure
}

func (t *SubMetrics) apply(next MsgHandler) MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) error {
		var err error

		defer func(start time.Time) {
			attr := extractAttr(msg, err != nil)

			t.counters[SubCount].Add(ctx, 1, attr...)
			t.counters[SubContentLength].Add(ctx, int64(len(msg.Data)), attr...)
			t.valueRecorders[SubLatency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)

		}(time.Now())

		err = next(ctx, msg)

		return err
	}
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

// PubMetric handle publish and request metrics gathering implementing PubMiddleware
type PubMetric struct {
	tele  tel.Telemetry
	meter metric.Meter

	counters       map[string]syncint64.Counter
	valueRecorders map[string]syncfloat64.Histogram

	root Publish
}

var _ PubMiddleware = &PubMetric{}

func NewPubMetric(tele tel.Telemetry, meter metric.Meter) *PubMetric {
	counters := make(map[string]syncint64.Counter)
	valueRecorders := make(map[string]syncfloat64.Histogram)

	regCount := func(name string) {
		v, err := meter.SyncInt64().Counter(name)
		if err != nil {
			tele.Panic("mw", tel.String("key", name))
		}
		counters[name] = v
	}

	regCount(OutCount)
	regCount(OutContentLength)
	regCount(RequestRespondContentLength)

	serverLatencyMeasure, err := meter.SyncFloat64().Histogram(OutLatency)
	if err != nil {
		tele.Panic("nats mw", tel.String("key", OutLatency))
	}

	valueRecorders[OutLatency] = serverLatencyMeasure

	return &PubMetric{
		counters:       counters,
		valueRecorders: valueRecorders,
	}
}

func (p *PubMetric) apply(in PubMiddleware) PubMiddleware {
	p.root = in

	return p
}

func (p *PubMetric) PublishWithContext(ctx context.Context, subj string, data []byte) (err error) {
	defer func(start time.Time) {
		attr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(subj),
		}

		p.counters[OutCount].Add(ctx, 1, attr...)
		p.counters[OutContentLength].Add(ctx, int64(len(data)), attr...)
		p.valueRecorders[OutLatency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)
	}(time.Now())

	return p.root.PublishWithContext(ctx, subj, data)
}

func (p *PubMetric) PublishMsgWithContext(ctx context.Context, msg *nats.Msg) (err error) {
	defer func(start time.Time) {
		attr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(msg.Subject),
		}

		p.counters[OutCount].Add(ctx, 1, attr...)
		p.counters[OutContentLength].Add(ctx, int64(len(msg.Data)), attr...)
		p.valueRecorders[OutLatency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)
	}(time.Now())

	return p.root.PublishMsgWithContext(ctx, msg)
}

func (p *PubMetric) PublishRequestWithContext(ctx context.Context, subj, reply string, data []byte) (err error) {
	defer func(start time.Time) {
		attr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(subj),
		}

		p.counters[OutCount].Add(ctx, 1, attr...)
		p.counters[OutContentLength].Add(ctx, int64(len(data)), attr...)
		p.valueRecorders[OutLatency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)
	}(time.Now())

	return p.root.PublishRequestWithContext(ctx, subj, reply, data)
}

func (p *PubMetric) RequestWithContext(ctx context.Context, subj string, data []byte) (resp *nats.Msg, err error) {
	defer func(start time.Time) {
		attr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(subj),
		}

		p.counters[OutCount].Add(ctx, 1, attr...)
		p.counters[OutContentLength].Add(ctx, int64(len(data)), attr...)
		p.valueRecorders[OutLatency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)

		if err == nil {
			p.counters[RequestRespondContentLength].Add(ctx, int64(len(resp.Data)), attr...)
		}
	}(time.Now())

	return p.root.RequestWithContext(ctx, subj, data)
}

func (p *PubMetric) RequestMsgWithContext(ctx context.Context, msg *nats.Msg) (resp *nats.Msg, err error) {
	defer func(start time.Time) {
		attr := []attribute.KeyValue{
			IsError.Bool(err != nil),
			Subject.String(msg.Subject),
		}

		p.counters[OutCount].Add(ctx, 1, attr...)
		p.counters[OutContentLength].Add(ctx, int64(len(msg.Subject)), attr...)
		p.valueRecorders[OutLatency].Record(ctx, float64(time.Since(start).Milliseconds()), attr...)

		if err == nil {
			p.counters[RequestRespondContentLength].Add(ctx, int64(len(resp.Data)), attr...)
		}
	}(time.Now())

	return p.root.RequestMsgWithContext(ctx, msg)
}
