package rules

import (
	"errors"
	"regexp"
	"strings"

	"github.com/tel-io/instrumentation/cardinality"
)

// New instance of configured cardinality replacer
/*
Options for disable placeholders:
   - WithMaxRuleCount
   - WithMaxSeparatorCount
   - WithConfigReader
*/
func New(rules []string, options ...Option) (cardinality.Replacer, error) {
	c := defaultConfig()
	for _, opt := range options {
		opt.apply(c)
	}

	if len(rules) >= c.maxRuleCount {
		return nil, errors.New("using too large a list of rules slows down processing")
	}

	m := &module{
		cfg:                 c,
		hasLeadingSeparator: c.reader.HasLeadingSeparator(),
		pathSeparator:       c.reader.PathSeparator(),
		placeholderRegexp:   c.reader.PlaceholderRegexp(),
	}

	err := m.preparePatterns(rules)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type module struct {
	cfg                 *config
	patterns            []pattern
	pathSeparator       string
	placeholderRegexp   *regexp.Regexp
	hasLeadingSeparator bool
}

// Replace cardinality parts in path
func (m *module) Replace(path string) string {
	for _, mu := range m.patterns {
		if nv, state := m.exec(mu, path); state {
			path = nv

			break
		}
	}

	return path
}

func (m *module) preparePatterns(rules []string) (err error) {
	for _, val := range rules {
		if strings.HasPrefix(val, m.pathSeparator) {
			val = strings.TrimLeft(val, m.pathSeparator)
		}

		pathParts := strings.Split(val, m.pathSeparator)
		pLen := len(pathParts)

		if pLen >= m.cfg.maxSeparatorCount {
			return errors.New("using too large rule slows down processing")
		}

		var parts = make([]part, 0, pLen)
		var valuePos int
		var valueExists, placeholderExists bool

		for i, pathPart := range pathParts {
			placeholderMatch := m.placeholderRegexp.MatchString(pathPart)
			if !placeholderMatch && !valueExists {
				valueExists = true
				valuePos = i
			}

			if placeholderMatch && !placeholderExists {
				placeholderExists = true
			}

			parts = append(parts, part{
				isPlaceholder: placeholderMatch,
				value:         pathPart,
			})
		}

		if !placeholderExists {
			return errors.New("redundant rule without placeholder")
		}

		m.patterns = append(m.patterns, pattern{
			parts:         parts,
			firstValuePos: valuePos,
		})
	}

	return nil
}

func (m *module) exec(pat pattern, path string) (string, bool) {
	urlParts := strings.Split(strings.TrimLeft(path, m.pathSeparator), m.pathSeparator)

	urlPartCount := len(urlParts)
	patternPartCount := len(pat.parts)

	if urlPartCount < patternPartCount {
		return path, false
	}

	if urlPartCount == patternPartCount {
		return m.mutateEqual(urlParts, patternPartCount, pat, path)
	}

	return m.mutatePartial(urlParts, urlPartCount, patternPartCount, pat, path)
}

func (m *module) mutateEqual(urlParts []string, patternPartCount int, pat pattern, path string) (string, bool) {
	for i := 0; i < patternPartCount; i++ {
		if pat.parts[i].isPlaceholder {
			urlParts[i] = pat.parts[i].value
			continue
		}

		if strings.Compare(urlParts[i], pat.parts[i].value) != 0 {
			return path, false
		}
	}

	return m.serialize(urlParts), true
}

func (m *module) mutatePartial(urlParts []string, urlPartCount int, patternPartCount int, pat pattern, path string) (string, bool) {
	for i := pat.firstValuePos; i < patternPartCount; i++ {
		for j := pat.firstValuePos; j < urlPartCount; j++ {
			if strings.Compare(urlParts[j], pat.parts[i].value) == 0 {
				left := j - i
				for iL := 0; iL < i; iL++ {
					if pat.parts[iL].isPlaceholder {
						urlParts[left+iL] = pat.parts[iL].value
					} else {
						return path, false
					}
				}

				//right := uLen - pLen - left
				for iR := i + 1; iR < patternPartCount; iR++ {
					if pat.parts[iR].isPlaceholder {
						urlParts[j+iR] = pat.parts[iR].value
					} else {
						return path, false
					}
				}

				return m.serialize(urlParts), true
			}
		}
	}

	return path, false
}

func (m *module) serialize(list []string) string {
	var prefix string
	if m.hasLeadingSeparator {
		prefix = m.pathSeparator
	}

	return prefix + strings.Join(list, m.pathSeparator)
}

type part struct {
	isPlaceholder bool
	value         string
}

type pattern struct {
	parts         []part
	firstValuePos int
}
