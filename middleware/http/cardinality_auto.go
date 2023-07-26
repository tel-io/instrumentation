package http

import (
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

func NewAutoGrouper() CardinalityGrouper {
	return func(path string) string {
		return decreasePathCardinality(path)
	}
}
