package nats

import (
	"context"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/metric"
)

// SubscriptionStatMetric hook provide important subscription statistics
type SubscriptionStatMetric struct {
	list     sync.Map
	counters map[string]metric.Int64ObservableGauge
}

func NewSubscriptionStatMetrics(opts ...Option) (*SubscriptionStatMetric, error) {
	conf := newConfig(opts)

	c := make(map[string]metric.Int64ObservableGauge)

	msgs, _ := conf.meter.Int64ObservableGauge(SubscriptionsPendingCount)
	bs, _ := conf.meter.Int64ObservableGauge(SubscriptionsPendingBytes,
		metric.WithUnit("By"),
	)

	dd, _ := conf.meter.Int64ObservableGauge(SubscriptionsDroppedMsgs)
	cc, _ := conf.meter.Int64ObservableGauge(SubscriptionCountMsgs)

	c[SubscriptionsPendingCount] = msgs
	c[SubscriptionsPendingBytes] = bs
	c[SubscriptionsDroppedMsgs] = dd
	c[SubscriptionCountMsgs] = cc

	res := &SubscriptionStatMetric{
		counters: c,
	}

	_, err := conf.meter.RegisterCallback(res.callback, []metric.Observable{msgs, bs, dd, cc}...)
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

func (s *SubscriptionStatMetric) callback(ctx context.Context, o metric.Observer) error {
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

		subject := decreaseSubjectCardinality(v.Subject)

		vc := data[subject]
		vc.msgs += int64(pMsg)
		vc.bytes += int64(pBytes)
		vc.dropped += int64(dropped)
		vc.count += count

		data[subject] = vc
		return true
	})

	for k, v := range data {
		o.ObserveInt64(
			s.counters[SubscriptionsPendingCount],
			v.msgs,
			metric.WithAttributes(Subject.String(k)),
		)

		o.ObserveInt64(
			s.counters[SubscriptionsPendingBytes],
			v.bytes,
			metric.WithAttributes(Subject.String(k)),
		)

		o.ObserveInt64(
			s.counters[SubscriptionsDroppedMsgs],
			v.dropped,
			metric.WithAttributes(Subject.String(k)),
		)

		o.ObserveInt64(
			s.counters[SubscriptionCountMsgs],
			v.count,
			metric.WithAttributes(Subject.String(k)),
		)
	}

	return nil
}
