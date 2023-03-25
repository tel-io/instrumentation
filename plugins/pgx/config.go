package pgx

import (
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type LoggerConfig struct {
	Dump bool

	NameFormatter NameFormatter
}

type TraceConfig struct {
	AllowRootTrace    bool
	NameFormatter     NameFormatter
	DefaultAttributes []attribute.KeyValue
	ErrorToStatus     ErrorToSpanStatus
}

type RecordConfig struct {
	meterProvider metric.MeterProvider

	DefaultAttributes []attribute.KeyValue
}

type config struct {
	tele *tel.Telemetry

	RecordConfig
	TraceConfig
	LoggerConfig
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
		tele: &l,

		LoggerConfig: LoggerConfig{
			Dump:          false,
			NameFormatter: formatSpanName,
		},
		TraceConfig: TraceConfig{
			AllowRootTrace: false,
			NameFormatter:  formatSpanName,
			ErrorToStatus:  spanStatusFromError,
		},
		RecordConfig: RecordConfig{
			meterProvider: l.MetricProvider(),
		},
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

// WithTel also add options to pass own metric and trace provider
func WithTel(t *tel.Telemetry) Option {
	return optionFunc(func(c *config) {
		c.tele = t
		c.meterProvider = t.MetricProvider()
	})
}

// WithLoggerDumpSQL perform by logging debug dump sql and arguments
func WithLoggerDumpSQL(enable bool) Option {
	return optionFunc(func(c *config) {
		c.LoggerConfig.Dump = enable
	})
}

// WithTraceRoot create trace if nod parent span occurred
func WithTraceRoot(enable bool) Option {
	return optionFunc(func(c *config) {
		c.TraceConfig.AllowRootTrace = enable
	})
}

// WithMeterProvider sets meter provider.
func WithMeterProvider(p metric.MeterProvider) Option {
	return optionFunc(func(c *config) {
		c.meterProvider = p
	})
}

func WithLoggerNameFormatter(fn NameFormatter) Option {
	return optionFunc(func(c *config) {
		c.LoggerConfig.NameFormatter = fn
	})
}

func WithTracerNameFormatter(fn NameFormatter) Option {
	return optionFunc(func(c *config) {
		c.TraceConfig.NameFormatter = fn
	})
}

func WithTraceDefaultArgs(args ...attribute.KeyValue) Option {
	return optionFunc(func(c *config) {
		c.TraceConfig.DefaultAttributes = args
	})
}

func WithRecordDefaultArgs(args ...attribute.KeyValue) Option {
	return optionFunc(func(c *config) {
		c.RecordConfig.DefaultAttributes = args
	})
}

// WithLoggerConfig overwrite default logger configuration
func WithLoggerConfig(cfg LoggerConfig) Option {
	return optionFunc(func(c *config) {
		c.LoggerConfig = cfg
	})
}

// WithTraceConfig overwrite default trace configuration
func WithTraceConfig(cfg TraceConfig) Option {
	return optionFunc(func(c *config) {
		c.TraceConfig = cfg
	})
}

// WithRecordConfig overwrite default metric configuration
func WithRecordConfig(cfg RecordConfig) Option {
	return optionFunc(func(c *config) {
		c.RecordConfig = cfg
	})
}
