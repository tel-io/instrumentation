package nats

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/metric"
)

var (
	rePartition = regexp.MustCompile(`\d+`)
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
type NameFn func(kind string, msg *nats.Msg) string

// defaultOperationFn default name convention
func defaultOperationFn(kind string, msg *nats.Msg) string {
	var b strings.Builder

	b.WriteString("NATS:")
	b.WriteString(kind)

	if msg.Sub != nil {
		b.WriteByte('/')
		b.WriteString(msg.Sub.Queue)
	}

	b.WriteByte('/')
	subjectParts := strings.Split(msg.Subject, ".")
	subjectPartsLastIdx := len(subjectParts) - 1
	for i, part := range subjectParts {
		p := part
		if rePartition.MatchString(part) {
			p = ":partition:"
		} else if strings.HasPrefix(part, "_INBOX") {
			p = ":inbox:"
		} else if strings.HasPrefix(part, "/") {
			p = ":url:"
		}

		b.WriteString(p)
		if i != subjectPartsLastIdx {
			b.WriteByte('.')
		}
	}

	return b.String()
}

type config struct {
	// Deprecated: only for legacy usage
	postHook PostHook

	tele    tel.Telemetry
	meter   metric.Meter
	metrics *metrics

	dump               bool
	dumpPayloadOnError bool

	nameFn NameFn

	// subMiddleware processors
	notUserDefaultMW bool
	pubList          []Middleware
	subList          []Middleware
}

func newConfig(opts []Option) *config {
	c := &config{
		tele:               tel.Global(),
		dumpPayloadOnError: true,
		notUserDefaultMW:   false,
		nameFn:             defaultOperationFn,
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

// DefaultMiddleware subInter interceptor
func (c *config) DefaultMiddleware() []Middleware {
	return []Middleware{
		NewRecovery(),
		NewLogs(c.nameFn, c.dumpPayloadOnError, c.dump),
		NewTracer(c.nameFn),
		NewMetrics(c.metrics),
	}
}

func (c *config) subMiddleware() []Middleware {
	if c.notUserDefaultMW {
		return c.subList
	}

	return append(c.DefaultMiddleware(), c.subList...)
}

func (c *config) pubMiddleware() []Middleware {
	if c.notUserDefaultMW {
		return c.pubList
	}

	return append(c.DefaultMiddleware(), c.pubList...)
}

// WithTel in some cases we should put another version
func WithTel(t tel.Telemetry) Option {
	return optionFunc(func(c *config) {
		c.tele = t
	})
}

// WithDump dump request as plain text to log and trace
// i guess we can go further and perform option with encoding requests
func WithDump(enable bool) Option {
	return optionFunc(func(c *config) {
		c.dump = enable
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
		c.nameFn = fn
	})
}

// WithSubMiddleware for subscriptions
func WithSubMiddleware(list ...Middleware) Option {
	return optionFunc(func(c *config) {
		c.subList = append(c.subList, list...)
	})
}

// WithPubMiddleware for publish
func WithPubMiddleware(list ...Middleware) Option {
	return optionFunc(func(c *config) {
		c.pubList = append(c.subList, list...)
	})
}

// WithDisableDefaultMiddleware disable default middleware usage
func WithDisableDefaultMiddleware() Option {
	return optionFunc(func(c *config) {
		c.notUserDefaultMW = true
	})
}

// WithReply extend mw with automatically sending reply on nats requests if they ask with data provided
// @inject - wrap nats.Msg handler with OTEL propagation data - extend traces, baggage and etc.
// Deprecated: legacy usage only
func WithReply(inject bool) Option {
	return WithPostHook(func(ctx context.Context, msg *nats.Msg, data []byte) error {
		return ReplyFn(ctx, msg, data)
	})
}

// WithPostHook set (only one) where you can perform post handle operation with data provided by handler
// Deprecated: legacy usage only
func WithPostHook(cb PostHook) Option {
	return optionFunc(func(c *config) {
		c.postHook = cb
	})
}
