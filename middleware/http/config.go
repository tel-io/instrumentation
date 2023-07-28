package http

import (
	"net/http"
	"strings"

	"github.com/tel-io/instrumentation/cardinality"
	"github.com/tel-io/instrumentation/cardinality/auto"
	"github.com/tel-io/tel/v2"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	replacers = cardinality.ReplacerList{
		auto.NewHttp(),
	}
	DefaultSpanNameFormatter = func(_ string, r *http.Request) string {
		var b strings.Builder

		b.WriteString(r.Method)
		b.WriteString(":")
		b.WriteString(replacers.Apply(r.URL.Path))

		return b.String()
	}

	DefaultFilter = func(r *http.Request) bool {
		if k, ok := r.Header["Upgrade"]; ok {
			for _, v := range k {
				if v == "websocket" {
					return false
				}
			}
		}

		return !(r.Method == http.MethodGet && strings.HasPrefix(r.URL.RequestURI(), "/health"))
	}
)

type PathExtractor func(r *http.Request) string

type config struct {
	log           *tel.Telemetry
	operation     string
	otelOpts      []otelhttp.Option
	pathExtractor PathExtractor
	filters       []otelhttp.Filter

	readRequest        bool
	readHeader         bool
	writeResponse      bool
	dumpPayloadOnError bool
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
		log:       &l,
		operation: "HTTP",
		otelOpts: []otelhttp.Option{
			otelhttp.WithSpanNameFormatter(DefaultSpanNameFormatter),
			otelhttp.WithFilter(DefaultFilter),
		},
		pathExtractor:      DefaultURI,
		filters:            []otelhttp.Filter{DefaultFilter},
		dumpPayloadOnError: true,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

// WithTel also add options to pass own metric and trace provider
func WithTel(t *tel.Telemetry) Option {
	return optionFunc(func(c *config) {
		c.log = t

		c.otelOpts = append(c.otelOpts,
			otelhttp.WithMeterProvider(t.MetricProvider()),
			otelhttp.WithTracerProvider(t.TracerProvider()),
		)
	})
}

func WithOperation(name string) Option {
	return optionFunc(func(c *config) {
		c.operation = name
	})
}

func WithOtelOpts(opts ...otelhttp.Option) Option {
	return optionFunc(func(c *config) {
		c.otelOpts = append(c.otelOpts, opts...)
	})
}

func WithPathExtractor(in PathExtractor) Option {
	return optionFunc(func(c *config) {
		c.pathExtractor = in
	})
}

// WithFilter append filter to default
func WithFilter(f ...otelhttp.Filter) Option {
	return optionFunc(func(c *config) {
		c.filters = append(c.filters, f...)

		for _, filter := range f {
			c.otelOpts = append(c.otelOpts, otelhttp.WithFilter(filter))
		}
	})
}

// WithDumpRequest dump request as plain text to log and trace
// i guess we can go further and perform option with encoding requests
func WithDumpRequest(enable bool) Option {
	return optionFunc(func(c *config) {
		c.readRequest = enable
	})
}

// WithHeaders explicitly set possibility to write http headers
func WithHeaders(enable bool) Option {
	return optionFunc(func(c *config) {
		c.readHeader = enable
	})
}

// WithDumpResponse dump response as plain text to log and trace
func WithDumpResponse(enable bool) Option {
	return optionFunc(func(c *config) {
		c.writeResponse = enable
	})
}

func DefaultURI(r *http.Request) string {
	return r.URL.RequestURI()
}

// WithDumpPayloadOnError write dump request and response on faults
//
// Default: true
func WithDumpPayloadOnError(enable bool) Option {
	return optionFunc(func(c *config) {
		c.dumpPayloadOnError = enable
	})
}
