package rules

import (
	"errors"
	"strings"

	"github.com/tel-io/instrumentation/cardinality"
)

// New instance of configured cardinality replacer
/*
Options for disable placeholders:
   - WithMaxRuleCount
   - WithMaxSeparatorCount
   - WithPathSeparator
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
		cfg: c,
	}

	err := m.preparePatterns(rules)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type module struct {
	cfg      *config
	patterns []pattern
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
		var isPartial, withValue bool
		if strings.HasPrefix(val, m.cfg.pathSeparator) {
			isPartial = true
			val = strings.TrimLeft(val, m.cfg.pathSeparator)
		}

		pathParts := strings.Split(val, m.cfg.pathSeparator)
		pLen := len(pathParts)

		if pLen >= m.cfg.maxSeparatorCount {
			return errors.New("using too large rule slows down processing")
		}

		var parts = make([]part, 0, pLen)
		var valuePos, placeholderPos *int
		var placeholderCount int

		for i, p := range pathParts {
			placeholderMatch := cardinality.PlaceholderRegexp.MatchString(p)
			if !placeholderMatch && valuePos == nil {
				iC := i
				valuePos = &iC
			}

			if placeholderMatch && placeholderPos == nil {
				iC := i
				placeholderPos = &iC
			}

			if placeholderMatch {
				placeholderCount++
			}

			parts = append(parts, part{
				isPlaceholder: placeholderMatch,
				value:         p,
			})
		}

		if placeholderPos == nil {
			return errors.New("redundant rule without placeholder")
		}

		var v int
		if valuePos != nil {
			v = *valuePos
			withValue = true
		}

		m.patterns = append(m.patterns, pattern{
			index:            len(m.patterns),
			parts:            parts,
			firstValue:       v,
			firstPlaceholder: *placeholderPos,
			withValue:        withValue,
			isPartial:        isPartial,
			placeholderCount: placeholderCount,
		})
	}

	return nil
}

type part struct {
	isPlaceholder bool
	value         string
}

type pattern struct {
	parts            []part
	firstValue       int
	firstPlaceholder int
	isPartial        bool
	withValue        bool
	placeholderCount int
	index            int
}

func (m *pattern) toStrings() []string {
	list := make([]string, len(m.parts))
	for i, v := range m.parts {
		list[i] = v.value
	}

	return list
}

func (m *module) serializePattern(pat pattern) string {
	return strings.Join(pat.toStrings(), m.cfg.pathSeparator)
}

func (m *module) exec(pat pattern, path string) (string, bool) {
	urlParts := m.newStringList(path)

	urlPartCount := len(urlParts)
	patternPartCount := len(pat.parts)

	if urlPartCount < patternPartCount {
		return path, false //firstValue rule (matches by min)
	}

	if urlPartCount == patternPartCount {
		return m.applyPipeEqual(urlParts, patternPartCount, pat, path)
	}

	return m.applyPipePartial(urlParts, urlPartCount, patternPartCount, pat, path)
}

func (m *module) serialize(list []string) string {
	return m.cfg.pathSeparator + strings.Join(list, m.cfg.pathSeparator)
}

func (m *module) newStringList(path string) []string {
	return strings.Split(strings.TrimLeft(path, m.cfg.pathSeparator), m.cfg.pathSeparator)
}

func (m *module) applyPipeEqual(urlParts []string, patternPartCount int, pat pattern, path string) (string, bool) {
	for i := 0; i < patternPartCount; i++ {
		if pat.parts[i].isPlaceholder {
			urlParts[i] = pat.parts[i].value
			continue
		}

		if strings.Compare(urlParts[i], pat.parts[i].value) != 0 {
			return path, false //firstValue rule
		}
	}

	return m.serialize(urlParts), true
}

func (m *module) mutate(i, j, pLen int, urlParts []string, pat pattern, path string) (string, bool) {
	left := j - i
	for iL := 0; iL < i; iL++ {
		if pat.parts[iL].isPlaceholder {
			urlParts[left+iL] = pat.parts[iL].value
		} else {
			return path, false
		}
	}

	//right := uLen - pLen - left
	for iR := i + 1; iR < pLen; iR++ {
		if pat.parts[iR].isPlaceholder {
			urlParts[j+iR] = pat.parts[iR].value
		} else {
			return path, false
		}
	}

	return m.serialize(urlParts), true
}

func (m *module) applyPipePartial(urlParts []string, urlPartCount int, patternPartCount int, pat pattern, path string) (string, bool) {
	for i := pat.firstValue; i < patternPartCount; i++ {
		for j := pat.firstValue; j < urlPartCount; j++ {
			if strings.Compare(urlParts[j], pat.parts[i].value) == 0 {
				if nv, state := m.mutate(i, j, patternPartCount, urlParts, pat, path); state {
					return nv, true
				}

				return path, false
			}
		}
	}

	return path, false
}
