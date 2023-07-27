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

func WithConfigReader(reader cardinality.ConfigReader) Option {
	return optionFunc(func(c *config) {
		c.reader = reader
	})
}

const (
	DefaultMaxRuleCount      = 100
	DefaultMaxSeparatorCount = 10
)

type Option interface {
	apply(*config)
}

type config struct {
	maxRuleCount      int
	maxSeparatorCount int
	reader            cardinality.ConfigReader
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

func defaultConfig() *config {
	return &config{
		reader:            cardinality.DefaultConfig(),
		maxRuleCount:      DefaultMaxRuleCount,
		maxSeparatorCount: DefaultMaxSeparatorCount,
	}
}
