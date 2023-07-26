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

	if len(rules) == 0 {
		return nil, errors.New("redundant using empty list of rules")
	}

	if len(rules) >= c.maxRuleCount {
		return nil, errors.New("using too large a list of rules slows down processing")
	}

	m := &module{
		cfg: c,
	}

	err := m.prepareRules(rules)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type module struct {
	cfg     *config
	mutates []pipe
}

// Replace cardinality parts in path
func (m *module) Replace(path string) string {
	for _, p := range m.mutates {
		if nv, state := p.exec(path); state {
			path = nv

			break
		}
	}

	return path
}

func (m *module) prepareRules(rules []string) (err error) {
	for _, val := range rules {
		var isPartial bool
		if strings.HasPrefix(val, m.cfg.pathSeparator) {
			isPartial = true
			val = strings.TrimLeft(val, m.cfg.pathSeparator)
		}

		pathParts := strings.Split(val, m.cfg.pathSeparator)
		pLen := len(pathParts)

		if pLen >= m.cfg.maxSeparatorCount {
			return errors.New("using too large condition slows down processing")
		}

		var parts = make([]part, 0, pLen)
		var valuePos, placeholderPos *int
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

			parts = append(parts, part{
				isPlaceholder: placeholderMatch,
				value:         p,
			})
		}

		if placeholderPos == nil {
			return errors.New("the condition without placeholder")
		}

		m.mutates = append(m.mutates, pipe{
			parts:            parts,
			skip:             valuePos,
			firstPlaceholder: *placeholderPos,
			isPartial:        isPartial,
		})
	}

	return nil
}

type part struct {
	isPlaceholder bool
	value         string
}

type pipe struct {
	parts            []part
	skip             *int
	firstPlaceholder int
	isPartial        bool
}

func (m *pipe) exec(path string) (string, bool) {
	return "", true
}

func (m *module) serialize(list []string) string {
	return m.cfg.pathSeparator + strings.Join(list, m.cfg.pathSeparator)
}

func (m *module) newStringList(path string) []string {
	return strings.Split(strings.TrimLeft(path, m.cfg.pathSeparator), m.cfg.pathSeparator)
}

func (m *module) applyPipeEqual(urlParts []string, patternPartCount int, p pipe, path string) (string, bool) {
	for i := 0; i < patternPartCount; i++ {
		if p.parts[i].isPlaceholder {
			urlParts[i] = p.parts[i].value
			continue
		}

		if strings.Compare(urlParts[i], p.parts[i].value) != 0 {
			return path, false //skip rule
		}
	}

	return m.serialize(urlParts), true
}

func (m *module) applyPipePartial(urlParts []string, urlPartCount int, patternPartCount int, p pipe, path string) (string, bool) {
	return m.serialize(urlParts), true
}

func (m *module) applyPipe(p pipe, path string) (string, bool) {
	urlParts := m.newStringList(path)

	urlPartCount := len(urlParts)
	patternPartCount := len(p.parts)

	if urlPartCount < patternPartCount {
		return path, false //skip rule (matches by min)
	}

	if urlPartCount == patternPartCount {
		return m.applyPipeEqual(urlParts, patternPartCount, p, path)
	}

	return m.applyPipePartial(urlParts, urlPartCount, patternPartCount, p, path)
}
