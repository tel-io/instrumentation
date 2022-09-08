package mongo

import (
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
)

func Inject(c *options.ClientOptions, opts ...Option) {
	cfg := newConfig(opts...)

	c.Monitor = otelmongo.NewMonitor(cfg.otelmongo...)
}
