package auto

import (
	"regexp"

	"github.com/tel-io/instrumentation/cardinality"
)

var (
	reID       = regexp.MustCompile(`^\d+$`)
	reResource = regexp.MustCompile(`^[a-zA-Z0-9\-]+\.\w{2,4}$`) // .css, .js, .png, .jpeg, etc.
	reUUID     = regexp.MustCompile(`^[a-f\d]{4}(?:[a-f\d]{4}-){4}[a-f\d]{12}$`)
)

const (
	keyId       = "id"
	keyResource = "resource"
	keyUUID     = "uuid"
)

func WithoutId() Option {
	return optionFunc(func(c *config) {
		delete(c.Matches, cardinality.PlaceholderFormatter(keyId))
	})
}

func WithoutResource() Option {
	return optionFunc(func(c *config) {
		delete(c.Matches, cardinality.PlaceholderFormatter(keyResource))
	})
}

func WithoutUUID() Option {
	return optionFunc(func(c *config) {
		delete(c.Matches, cardinality.PlaceholderFormatter(keyUUID))
	})
}

type Option interface {
	apply(*config)
}

type config struct {
	RuleSeparator string
	Matches       map[string]*regexp.Regexp
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

func defaultConfig() *config {
	return &config{
		RuleSeparator: cardinality.PathSeparator,
		Matches: map[string]*regexp.Regexp{
			cardinality.PlaceholderFormatter(keyId):       reID,
			cardinality.PlaceholderFormatter(keyResource): reResource,
			cardinality.PlaceholderFormatter(keyUUID):     reUUID,
		},
	}
}
