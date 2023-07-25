package http

import (
	"errors"
	"regexp"
	"strings"
)

var (
	reID       = regexp.MustCompile(`^\d+$`)
	reResource = regexp.MustCompile(`^[a-zA-Z0-9\-]+\.\w{2,4}$`) // .css, .js, .png, .jpeg, etc.
	reUUID     = regexp.MustCompile(`^[a-f\d]{4}(?:[a-f\d]{4}-){4}[a-f\d]{12}$`)

	decreasePathCardinality = func(path string) string {
		var b strings.Builder

		path = strings.TrimLeft(path, "/")
		pathParts := strings.Split(path, "/")
		for _, part := range pathParts {
			b.WriteString("/")

			p := part
			if reID.MatchString(part) {
				p = ":id:"
			} else if reResource.MatchString(part) {
				p = ":resource:"
			} else if reUUID.MatchString(part) {
				p = ":uuid:"
			}
			b.WriteString(p)
		}

		return b.String()
	}
)

type CardinalityGrouper func(path string) string

type CardinalityGrouperList []CardinalityGrouper

func (m CardinalityGrouperList) Apply(path string) string {
	for _, grp := range m {
		path = grp(path)
	}

	return path
}

func WithCardinalityGroupers(list []CardinalityGrouper) Option {
	return optionFunc(func(c *config) {
		c.groupers = list
	})
}

func NewAutoGrouper() CardinalityGrouper {
	return func(path string) string {
		return decreasePathCardinality(path)
	}
}

var placeholderExp = regexp.MustCompile(`^:[-\w]+$`)

type part struct {
	isPlaceholder bool
	value         string
}

type pipe struct {
	parts []part
	skip  int
}

func rules2Pipe(rules []string) (mutates []pipe, err error) {
	for _, val := range rules {
		pathParts := strings.Split(strings.TrimLeft(val, "/"), "/")
		pLen := len(pathParts)

		if pLen > 10 {
			return nil, errors.New("using too large condition slows down processing")
		}

		if pLen < 2 {
			return nil, errors.New("the condition is not precise enough. (min.len:2)")
		}

		var parts = make([]part, 0, pLen)
		var placeholderPos *int
		for i, p := range pathParts {
			placeholderMatch := placeholderExp.MatchString(p)
			if placeholderMatch && placeholderPos == nil {
				placeholderPos = &i
			}

			parts = append(parts, part{
				isPlaceholder: placeholderMatch,
				value:         p,
			})
		}

		if placeholderPos == nil {
			return nil, errors.New("the condition without placeholder")
		}

		mutates = append(mutates, pipe{
			parts: parts,
			skip:  *placeholderPos,
		})
	}

	return
}

var applyFn = func(m pipe, path string) string {
	pathParts := strings.Split(strings.TrimLeft(path, "/"), "/")

	aLen := len(pathParts)
	eLen := len(m.parts)

	if eLen > aLen {
		return path
	}

	var isModified bool

	////TODO optimize range
	//for j := m.skip; j < len(m.parts); j++ {
	//	for _, a := range pathParts {
	//		//find value marker
	//		if strings.Compare(a, m.parts[j].value) == 0 {
	//			continue
	//		}
	//	}
	//}

	if !isModified {
		return path
	}

	return "/" + strings.Join(pathParts, "/")
}

func NewRulesGrouper(rules []string) (CardinalityGrouper, error) {
	if len(rules) >= 100 {
		return nil, errors.New("using too large a list of rules slows down processing")
	}

	//prepare
	pipeline, err := rules2Pipe(rules)
	if err != nil {
		return nil, err
	}

	//callback
	return func(path string) string {
		for _, m := range pipeline {
			path = applyFn(m, path)
		}

		return path
	}, nil
}
