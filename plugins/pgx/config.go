package pgx

import (
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type config struct {
	loggerProvider tel.Logger
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider

	defaultAttributes []attribute.KeyValue
	spanNameFormatter SpanNameFormatter

	dumpSQL bool
}

// Option interface used for setting optional config properties.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// newConfig creates a new config struct and applies opts to it.
func newConfig(opts ...Option) *config {
	l := tel.Global()

	c := &config{
		loggerProvider:    l,
		meterProvider:     l.MetricProvider(),
		tracerProvider:    l.TracerProvider(),
		dumpSQL:           false,
		spanNameFormatter: formatSpanName,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

// WithTel also add options to pass own metric and trace provider
func WithTel(t *tel.Telemetry) Option {
	return optionFunc(func(c *config) {
		c.loggerProvider = t
		c.meterProvider = t.MetricProvider()
		c.tracerProvider = t.TracerProvider()
	})
}

// WithDumpSQL perform debug dump sql and argument
func WithDumpSQL(enable bool) Option {
	return optionFunc(func(c *config) {
		c.dumpSQL = enable
	})
}

// WithMeterProvider sets meter provider.
func WithMeterProvider(p metric.MeterProvider) Option {
	return optionFunc(func(c *config) {
		c.meterProvider = p
	})
}

// WithTracerProvider sets tracer provider.
func WithTracerProvider(p trace.TracerProvider) Option {
	return optionFunc(func(c *config) {
		c.tracerProvider = p
	})
}

func WithDefaultArgs(args ...attribute.KeyValue) Option {
	return optionFunc(func(c *config) {
		c.defaultAttributes = args
	})
}
