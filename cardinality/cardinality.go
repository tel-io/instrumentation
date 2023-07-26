package cardinality

import (
	"fmt"
	"regexp"
)

const (
	PathSeparator = "/"
)

var (
	PlaceholderFormatter = func(id string) string { return fmt.Sprintf(`:%s`, id) }
	PlaceholderRegexp    = regexp.MustCompile(`^:[-\w]+$`)
)

type Replacer interface {
	Replace(path string) string
}

type ReplacerList []Replacer

func (m ReplacerList) Apply(path string) string {
	for _, r := range m {
		path = r.Replace(path)
	}

	return path
}
