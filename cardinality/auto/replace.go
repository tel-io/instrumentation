package auto

import (
	"strings"

	"github.com/tel-io/instrumentation/cardinality"
)

// New instance of automatic cardinality replacer
/*
Options for disable placeholders:
   - WithoutId
   - WithoutResource
   - WithoutUUID
*/
func New(options ...Option) cardinality.Replacer {
	c := defaultConfig()
	for _, opt := range options {
		opt.apply(c)
	}

	return &module{
		cfg: c,
	}
}

type module struct {
	cfg *config
}

// Replace cardinality parts in path
func (m *module) Replace(path string) string {
	var b strings.Builder

	path = strings.TrimLeft(path, m.cfg.RuleSeparator)
	pathParts := strings.Split(path, m.cfg.RuleSeparator)
	for _, part := range pathParts {
		b.WriteString(m.cfg.RuleSeparator)

		p := part

		for id, exp := range m.cfg.Matches {
			if exp.MatchString(part) {
				p = id
				break
			}
		}

		b.WriteString(p)
	}

	return b.String()
}
