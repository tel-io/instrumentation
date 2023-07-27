package cardinality_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tel-io/instrumentation/cardinality"
	"github.com/tel-io/instrumentation/cardinality/auto"
	"github.com/tel-io/instrumentation/cardinality/rules"
)

func TestConfig(t *testing.T) {
	cfg := cardinality.GlobalConfig()
	assert.Equal(t, true, cfg.HasLeadingSeparator())
	assert.Equal(t, "/", cfg.PathSeparator())

	assert.Equal(t, ":id", cfg.PlaceholderFormatter()("id"))
	assert.True(t, cfg.PlaceholderRegexp().MatchString(":id"))

	cfg = cardinality.NewConfig(
		cardinality.WithPathSeparator(false, "."),
		cardinality.WithPlaceholder(regexp.MustCompile(`^\{[-\w]+}$`), func(id string) string {
			return fmt.Sprintf(`{%s}`, id)
		}),
	)

	assert.Equal(t, false, cfg.HasLeadingSeparator())
	assert.Equal(t, ".", cfg.PathSeparator())

	assert.Equal(t, "{id}", cfg.PlaceholderFormatter()("id"))
	assert.True(t, cfg.PlaceholderRegexp().MatchString("{id}"))
}

func TestApply(t *testing.T) {
	list := cardinality.ReplacerList{
		auto.New(),
		func() cardinality.Replacer {
			r, err := rules.New([]string{
				"/:service/:action",
			})
			assert.NoError(t, err)
			return r
		}(),
	}
	list.Apply("/player/update/123")
}
