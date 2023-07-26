package http

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	placeholderExp = regexp.MustCompile(`^:[-\w]+$`)
)

type (
	stringList []string
	part       struct {
		isPlaceholder bool
		value         string
	}
	pipe struct {
		parts            []part
		skip             *int
		firstPlaceholder int
	}
)

const (
	maxRuleCount     = 100
	maxRulePartCount = 10
)

func NewRulesGrouper(rules []string) (CardinalityGrouper, error) {
	if len(rules) == 0 {
		return nil, errors.New("redundant using empty list of rules")
	}

	if len(rules) >= maxRuleCount {
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
			if nv, state := applyPipe(m, path); state {
				path = nv

				break
			}
		}

		return path
	}, nil
}

func (m stringList) serialize() string {
	return "/" + strings.Join(m, "/")
}

func newStringList(path string) stringList {
	return strings.Split(strings.TrimLeft(path, "/"), "/")
}

func applyPipe(m pipe, path string) (string, bool) {
	urlParts := newStringList(path)

	urlPartCount := len(urlParts)
	patternPartCount := len(m.parts)

	if urlPartCount < patternPartCount {
		return path, false //skip rule (matches by min)
	}

	var startCompareFlag bool
	var startCompareOffset int
	var elapsedPattern int
	var skipped int

	//fmt.Println(path, m.parts)
	//fmt.Println("p", "u")

	for p := 0; p < patternPartCount; p++ {
		if m.parts[p].isPlaceholder {
			skipped++
			urlParts[p] = m.parts[p].value
			//fmt.Println(p, ":", "skip")
			continue
		}

		elapsedPattern = patternPartCount - p

		for u := 0 + skipped; u < urlPartCount; u++ {
			if strings.Compare(urlParts[u], m.parts[p].value) == 0 {
				//fmt.Println(p, u, "+")
				if !startCompareFlag {
					startCompareFlag = true

					startCompareOffset = u
				}

				skipped++

				break
			} else {
				return path, false //skip rule (different suffix)
			}
		}
	}

	_ = elapsedPattern
	_ = startCompareOffset

	return urlParts.serialize(), true
}

func applyPipe2(m pipe, path string) string {
	urlParts := newStringList(path)

	urlPartCount := len(urlParts)
	patternPartCount := len(m.parts)

	if patternPartCount > urlPartCount {
		return path //skip rule (matches by min)
	}

	//only placeholders rule. Depends only on the number of separators
	if m.skip == nil {
		if patternPartCount == urlPartCount {
			//copy from pipe
			for i, p := range m.parts {
				urlParts[i] = p.value
			}

			return urlParts.serialize()
		}

		return path //skip rule (matches by max)
	}

	//placeholders at the end
	if m.skip != nil && *m.skip == 0 && patternPartCount == urlPartCount {
		for i, p := range m.parts {
			if p.isPlaceholder {
				urlParts[i] = p.value
			}

			if strings.Compare(urlParts[i], p.value) != 0 {
				return path //skip rule (matches)
			}
		}

		return urlParts.serialize()
	}

	var offset int
	var startCompareOffset int
	var startCompareFlag bool

	fmt.Println(path, m.parts)
	fmt.Println("p", "u")
	//pattern
	for p := *m.skip; p < patternPartCount; p++ {
		//url
		for u := *m.skip + offset; u < urlPartCount; u++ {
			fmt.Println(p, u)
			if m.parts[p].isPlaceholder {
				urlParts[u] = m.parts[p].value
				if startCompareFlag {
					offset++
				}

				continue
			}

			if strings.Compare(urlParts[u], m.parts[p].value) == 0 {
				if startCompareFlag == false {
					//right scenario
					if *m.skip == 0 {
						offset = 0
						startCompareOffset = 0
					} else {
						offset = u
						startCompareOffset = u - 1
					}

					startCompareFlag = true
				} else {
					offset++
				}

				break
			} else if startCompareFlag {
				return path //skip rule (different suffix)
			}
		}
	}

	if !startCompareFlag {
		return path
	}

	//replace the placeholders that are before the beginning of the comparison
	for j := 0; j < *m.skip; j++ {
		//fmt.Println(j, j+startCompareOffset)
		if !m.parts[j].isPlaceholder {
			continue
		}

		urlParts[j+startCompareOffset] = m.parts[j].value
	}

	return urlParts.serialize()
}

func rules2Pipe(rules []string) (mutates []pipe, err error) {
	for _, val := range rules {
		pathParts := strings.Split(strings.TrimLeft(val, "/"), "/")
		pLen := len(pathParts)

		if pLen >= maxRulePartCount { //TODO to config
			return nil, errors.New("using too large condition slows down processing")
		}

		var parts = make([]part, 0, pLen)
		var valuePos, placeholderPos *int
		for i, p := range pathParts {
			placeholderMatch := placeholderExp.MatchString(p)
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
			return nil, errors.New("the condition without placeholder")
		}

		mutates = append(mutates, pipe{
			parts:            parts,
			skip:             valuePos,
			firstPlaceholder: *placeholderPos,
		})
	}

	return
}
