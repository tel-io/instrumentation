package cardinality

import (
	"fmt"
	"regexp"
)

var (
	globalConfig = defaultConfig()
)

func GlobalConfig() ConfigReader {
	return globalConfig
}

type ConfigReader interface {
	PlaceholderFormatter() func(id string) string
	PlaceholderRegexp() *regexp.Regexp
	PathSeparator() string
	HasLeadingSeparator() bool
}

func NewConfig(options ...Option) ConfigReader {
	c := defaultConfig()
	for _, opt := range options {
		opt.apply(c)
	}

	return c
}

func WithPathSeparator(hasLeadingSeparator bool, separator string) Option {
	return optionFunc(func(c *immutableCfg) {
		c.hasLeadingSeparator = hasLeadingSeparator
		c.pathSeparator = separator
	})
}

func WithPlaceholder(placeholderRegexp *regexp.Regexp, placeholderFormatter func(string) string) Option {
	return optionFunc(func(c *immutableCfg) {
		c.placeholderRegexp = placeholderRegexp
		c.placeholderFormatter = placeholderFormatter
	})
}

type Option interface {
	apply(*immutableCfg)
}

type optionFunc func(*immutableCfg)

func (o optionFunc) apply(c *immutableCfg) {
	o(c)
}

func defaultConfig() *immutableCfg {
	return &immutableCfg{
		true,
		"/",
		regexp.MustCompile(`^:[-\w]+$`),
		func(id string) string {
			return fmt.Sprintf(`:%s`, id)
		},
	}
}

type immutableCfg struct {
	hasLeadingSeparator  bool
	pathSeparator        string
	placeholderRegexp    *regexp.Regexp
	placeholderFormatter func(id string) string
}

func (m *immutableCfg) HasLeadingSeparator() bool {
	return m.hasLeadingSeparator
}

func (m *immutableCfg) PlaceholderFormatter() func(id string) string {
	return m.placeholderFormatter
}

func (m *immutableCfg) PlaceholderRegexp() *regexp.Regexp {
	return m.placeholderRegexp
}

func (m *immutableCfg) PathSeparator() string {
	return m.pathSeparator
}
