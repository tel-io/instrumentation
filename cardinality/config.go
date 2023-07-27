package cardinality

import (
	"fmt"
	"regexp"
)

var (
	globalConfig = NewConfig(
		"/",
		true,
		regexp.MustCompile(`^:[-\w]+$`),
		func(id string) string {
			return fmt.Sprintf(`:%s`, id)
		},
	)
)

func DefaultConfig() ConfigReader {
	return globalConfig
}

type ConfigReader interface {
	PlaceholderFormatter() func(id string) string
	PlaceholderRegexp() *regexp.Regexp
	PathSeparator() string
	HasLeadingSeparator() bool
}

func NewConfig(separator string, hasLeadingSeparator bool, exp *regexp.Regexp, formatter func(id string) string) ConfigReader {
	return &immutableCfg{
		hasLeadingSeparator:  hasLeadingSeparator,
		pathSeparator:        separator,
		placeholderRegexp:    exp,
		placeholderFormatter: formatter,
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
