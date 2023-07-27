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
	KeyId       = "id"
	KeyResource = "resource"
	KeyUUID     = "uuid"
)

func WithConfigReader(reader cardinality.ConfigReader) Option {
	return optionFunc(func(c *config) {
		c.reader = reader
	})
}

func WithoutId() Option {
	return optionFunc(func(c *config) {
		for i, m := range c.matches {
			if m.id == KeyId {
				c.matches[i].state = false
				return
			}
		}
	})
}

func WithoutResource() Option {
	return optionFunc(func(c *config) {
		for i, m := range c.matches {
			if m.id == KeyResource {
				c.matches[i].state = false
				return
			}
		}
	})
}

func WithoutUUID() Option {
	return optionFunc(func(c *config) {
		for i, m := range c.matches {
			if m.id == KeyUUID {
				c.matches[i].state = false
				return
			}
		}
	})
}

type Option interface {
	apply(*config)
}

type matchState struct {
	*regexp.Regexp
	state bool
	id    string
}

type config struct {
	matches []matchState //array instead of map for save order
	reader  cardinality.ConfigReader
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

func defaultConfig() *config {
	return &config{
		reader: cardinality.DefaultConfig(),
		matches: []matchState{
			{
				Regexp: reID,
				state:  true,
				id:     KeyId,
			},
			{
				Regexp: reResource,
				state:  true,
				id:     KeyResource,
			},
			{
				Regexp: reUUID,
				state:  true,
				id:     KeyUUID,
			},
		},
	}
}
