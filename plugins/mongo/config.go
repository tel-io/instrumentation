package mongo

import (
	"github.com/d7561985/tel/v2"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

type config struct {
	log       *tel.Telemetry
	otelmongo []otelmongo.Option
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
		log: &l,
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

		c.otelmongo = append(c.otelmongo,
			//otelmongo.WithMeterProvider(t.MetricProvider()),
			otelmongo.WithTracerProvider(t.TracerProvider()),
		)
	})
}

func WithOtelConf(opt ...otelmongo.Option) Option {
	return optionFunc(func(c *config) {
		c.otelmongo = append(c.otelmongo, opt...)
	})
}
