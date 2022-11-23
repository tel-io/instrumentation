package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"github.com/tel-io/instrumentation/middleware/nats/natsprop"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/metric"
)

// Option allows configuration of the httptrace Extract()
// and Inject() functions.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

type PostHook func(ctx context.Context, msg *nats.Msg, data []byte) error

// NameFn operation name description
type NameFn func(msg *nats.Msg) string

// defaultSubOperationFn default name convention
func defaultSubOperationFn(msg *nats.Msg) string {
	return fmt.Sprintf("NATS:SUB/%s/%s", msg.Sub.Queue, msg.Subject)
}

func defaultPubOperationFn(msg *nats.Msg) string {
	return fmt.Sprintf("NATS:PUB/%s", msg.Subject)
}

type config struct {
	// Deprecated: only for legacy usage
	postHook PostHook

	tele    tel.Telemetry
	meter   metric.Meter
	metrics *metrics

	dumpRequest        bool
	dumpResponse       bool
	dumpPayloadOnError bool

	subNameFn NameFn
	pubNameFn NameFn

	list    []Middleware
	pubList []PubMiddleware
}

func newConfig(opts []Option) *config {
	c := &config{
		tele:               tel.Global(),
		dumpPayloadOnError: true,
		subNameFn:          defaultSubOperationFn,
		pubNameFn:          defaultPubOperationFn,
	}

	c.apply(opts)

	c.meter = c.tele.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(SemVersion()),
	)

	c.metrics = createMeasures(c.tele, c.meter)

	return c
}

func (c *config) apply(opts []Option) {
	for _, o := range opts {
		o.apply(c)
	}
}

func (c *config) DefaultMiddleware() []Middleware {
	return []Middleware{
		NewTracer(c.subNameFn),
		NewLogs(c),
		NewMetrics(c.metrics),
		NewRecovery(),
	}
}

func (c *config) Middleware() []Middleware {
	return append(c.DefaultMiddleware(), c.list...)
}

func (c *config) DefaultPubMiddleware() []PubMiddleware {
	return []PubMiddleware{
		NewPubTrace(c.pubNameFn),
		NewPubMetric(c.metrics),
	}
}

func (c *config) PubMiddleware() []PubMiddleware {
	return append(c.DefaultPubMiddleware(), c.pubList...)
}

// WithReply extend mw with automatically sending reply on nats requests if they ask with data provided
// @inject - wrap nats.Msg handler with OTEL propagation data - extend traces, baggage and etc.
// Deprecated: legacy usage only
func WithReply(inject bool) Option {
	return WithPostHook(func(ctx context.Context, msg *nats.Msg, data []byte) error {
		if msg.Reply == "" {
			return nil
		}

		resMsg := &nats.Msg{Data: data}
		if inject {
			natsprop.Inject(ctx, msg)
		}

		if err := msg.RespondMsg(resMsg); err != nil {
			return errors.WithStack(err)
		}

		return nil
	})
}

// WithPostHook set (only one) where you can perform post handle operation with data provided by handler
// Deprecated: legacy usage only
func WithPostHook(cb PostHook) Option {
	return optionFunc(func(c *config) {
		c.postHook = cb
	})
}

// WithTel in some cases we should put another version
func WithTel(t tel.Telemetry) Option {
	return optionFunc(func(c *config) {
		c.tele = t
	})
}

// WithDumpRequest dump request as plain text to log and trace
// i guess we can go further and perform option with encoding requests
func WithDumpRequest(enable bool) Option {
	return optionFunc(func(c *config) {
		c.dumpRequest = enable
	})
}

// WithDumpResponse dump response as plain text to log and trace
func WithDumpResponse(enable bool) Option {
	return optionFunc(func(c *config) {
		c.dumpResponse = enable
	})
}

// WithDumpPayloadOnError write dump request and response on faults
//
// Default: true
func WithDumpPayloadOnError(enable bool) Option {
	return optionFunc(func(c *config) {
		c.dumpPayloadOnError = enable
	})
}

func WithNameFunction(fn NameFn) Option {
	return optionFunc(func(c *config) {
		c.subNameFn = fn
	})
}
