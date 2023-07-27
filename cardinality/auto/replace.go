package auto

import (
	"regexp"
	"strings"

	"github.com/tel-io/instrumentation/cardinality"
)

// New instance of automatic cardinality replacer
/*
Options for disable placeholders:
   - WithoutId
   - WithoutResource
   - WithoutUUID
   - WithConfigReader
*/
func New(options ...Option) cardinality.Replacer {
	c := defaultConfig()
	for _, opt := range options {
		opt.apply(c)
	}

	formatter := c.reader.PlaceholderFormatter()

	var matches []match
	for _, m := range c.matches {
		if !m.state {
			continue
		}

		matches = append(matches, match{
			Regexp:      m.Regexp,
			placeholder: formatter(m.id),
		})
	}

	return &module{
		matches:             matches,
		separator:           c.reader.PathSeparator(),
		hasLeadingSeparator: c.reader.HasLeadingSeparator(),
	}
}

type match struct {
	*regexp.Regexp
	placeholder string
}

type module struct {
	separator           string
	matches             []match
	hasLeadingSeparator bool
}

// Replace cardinality parts in path
func (m *module) Replace(path string) string {
	path = strings.TrimLeft(path, m.separator)
	pathParts := strings.Split(path, m.separator)

	var b strings.Builder

	for i, part := range pathParts {
		if m.hasLeadingSeparator || (!m.hasLeadingSeparator && i != 0) {
			b.WriteString(m.separator)
		}

		p := part

		for _, exp := range m.matches {
			if exp.MatchString(part) {
				p = exp.placeholder
				break
			}
		}

		b.WriteString(p)
	}

	return b.String()
}
