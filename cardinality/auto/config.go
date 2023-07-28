package auto

import (
	"fmt"
	"regexp"

	"github.com/tel-io/instrumentation/cardinality"
)

var (
	reID       = regexp.MustCompile(`^\d+$`)
	reResource = regexp.MustCompile(`^[a-zA-Z0-9\-]+\.\w{2,4}$`) // .css, .js, .png, .jpeg, etc.
	reUUID     = regexp.MustCompile(`^[a-f\d]{4}(?:[a-f\d]{4}-){4}[a-f\d]{12}$`)
)

const (
	KeyId        = "id"
	KeyResource  = "resource"
	KeyUUID      = "uuid"
	KeyPartition = "partition"
	KeyInbox     = "inbox"
	KeyUrl       = "url"
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

func WithoutPartition() Option {
	return optionFunc(func(c *config) {
		for i, m := range c.matches {
			if m.id == KeyPartition {
				c.matches[i].state = false
				return
			}
		}
	})
}

func WithoutInbox() Option {
	return optionFunc(func(c *config) {
		for i, m := range c.prefixes {
			if m.id == KeyInbox {
				c.prefixes[i].state = false
				return
			}
		}
	})
}

func WithoutUrl() Option {
	return optionFunc(func(c *config) {
		for i, m := range c.prefixes {
			if m.id == KeyUrl {
				c.prefixes[i].state = false
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
	state  bool
	id     string
	prefix string
}

type config struct {
	//array instead of map for save order
	prefixes []matchState
	matches  []matchState

	reader cardinality.ConfigReader
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

func defaultConfigNats() *config {
	cfg := cardinality.NewConfig(
		cardinality.WithPathSeparator(false, "."),
		cardinality.WithPlaceholder(nil, func(s string) string {
			return fmt.Sprintf("{%s}", s)
		}),
	)

	return &config{
		reader: cfg,
		prefixes: []matchState{
			{
				state:  true,
				id:     KeyInbox,
				prefix: "_INBOX",
			},
			{
				state:  true,
				id:     KeyUrl,
				prefix: "/",
			},
		},
		matches: []matchState{
			{
				state:  true,
				Regexp: reID,
				id:     KeyPartition,
			},
		},
	}
}

func defaultConfigHttp() *config {
	cfg := cardinality.NewConfig(
		cardinality.WithPathSeparator(true, "/"),
		cardinality.WithPlaceholder(nil, func(s string) string {
			return fmt.Sprintf(":%s", s)
		}),
	)

	return &config{
		reader: cfg,
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
