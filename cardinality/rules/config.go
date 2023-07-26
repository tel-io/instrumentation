package rules

import (
	"github.com/tel-io/instrumentation/cardinality"
)

func WithMaxRuleCount(max int) Option {
	return optionFunc(func(c *config) {
		c.maxRuleCount = max
	})
}

func WithMaxSeparatorCount(max int) Option {
	return optionFunc(func(c *config) {
		c.maxSeparatorCount = max
	})
}

func WithPathSeparator(separator string) Option {
	return optionFunc(func(c *config) {
		c.pathSeparator = separator
	})
}

const (
	DefaultMaxRuleCount      = 100
	DefaultMaxSeparatorCount = 10
	DefaultPathSeparator     = cardinality.PathSeparator
)

type Option interface {
	apply(*config)
}

type config struct {
	maxRuleCount      int
	maxSeparatorCount int
	pathSeparator     string
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

func defaultConfig() *config {
	return &config{
		pathSeparator:     DefaultPathSeparator,
		maxRuleCount:      DefaultMaxRuleCount,
		maxSeparatorCount: DefaultMaxSeparatorCount,
	}
}
