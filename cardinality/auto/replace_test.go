package auto_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tel-io/instrumentation/cardinality"
	"github.com/tel-io/instrumentation/cardinality/auto"
)

func TestAuto(t *testing.T) {
	var r cardinality.Replacer

	u := "/player/update/123/file/favicon.ico/550e8400-e29b-41d4-a716-446655440000"

	r = auto.New(auto.WithoutId())
	assert.Equal(t, "/player/update/123/file/:resource/:uuid", r.Replace(u))

	r = auto.New(auto.WithoutResource())
	assert.Equal(t, "/player/update/:id/file/favicon.ico/:uuid", r.Replace(u))

	r = auto.New(auto.WithoutUUID())
	assert.Equal(t, "/player/update/:id/file/:resource/550e8400-e29b-41d4-a716-446655440000", r.Replace(u))

	cfg := cardinality.NewConfig(
		".",
		false,
		regexp.MustCompile(`^\{[-\w]+}$`),
		func(id string) string {
			return fmt.Sprintf(`{%s}`, id)
		},
	)
	r = auto.New(auto.WithConfigReader(cfg))
	assert.Equal(t, "player.update.{id}",
		r.Replace("player.update.123"))
}
