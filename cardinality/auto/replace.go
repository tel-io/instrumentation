package auto

import (
	"regexp"
	"strings"

	"github.com/tel-io/instrumentation/cardinality"
)

// NewHttp instance of automatic cardinality replacer
/*
Options for disable placeholders:
   - WithoutId
   - WithoutResource
   - WithoutUUID
   - WithConfigReader
*/
func NewHttp(options ...Option) cardinality.Replacer {
	c := defaultConfigHttp()
	for _, opt := range options {
		opt.apply(c)
	}

	return fromCfg(c)
}

// NewNats instance of automatic cardinality replacer
/*
Options for disable placeholders:
   - WithoutPartition
   - WithoutUrl
   - WithoutInbox
   - WithConfigReader
*/
func NewNats(options ...Option) cardinality.Replacer {
	c := defaultConfigNats()
	for _, opt := range options {
		opt.apply(c)
	}

	return fromCfg(c)
}

type module struct {
	separator           string
	matches             []match
	prefixes            []prefix
	hasLeadingSeparator bool
}

// Replace cardinality parts in path
func (m *module) Replace(path string) string {
	for _, pre := range m.prefixes {
		if strings.HasPrefix(path, pre.value) {
			return pre.placeholder
		}
	}

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

func fromCfg(c *config) cardinality.Replacer {
	formatter := c.reader.PlaceholderFormatter()

	var prefixes []prefix
	for _, m := range c.prefixes {
		if !m.state {
			continue
		}

		prefixes = append(prefixes, prefix{
			value:       m.prefix,
			placeholder: formatter(m.id),
		})
	}

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
		prefixes:            prefixes,
		matches:             matches,
		separator:           c.reader.PathSeparator(),
		hasLeadingSeparator: c.reader.HasLeadingSeparator(),
	}
}

type match struct {
	*regexp.Regexp
	placeholder string
}

type prefix struct {
	value       string
	placeholder string
}
