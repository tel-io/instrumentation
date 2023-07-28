package auto_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tel-io/instrumentation/cardinality"
	"github.com/tel-io/instrumentation/cardinality/auto"
)

func TestHttp(t *testing.T) {
	var r cardinality.Replacer

	u := "/player/update/123/file/favicon.ico/550e8400-e29b-41d4-a716-446655440000"

	r = auto.NewHttp(auto.WithoutId())
	assert.Equal(t, "/player/update/123/file/:resource/:uuid", r.Replace(u))

	r = auto.NewHttp(auto.WithoutResource())
	assert.Equal(t, "/player/update/:id/file/favicon.ico/:uuid", r.Replace(u))

	r = auto.NewHttp(auto.WithoutUUID())
	assert.Equal(t, "/player/update/:id/file/:resource/550e8400-e29b-41d4-a716-446655440000", r.Replace(u))
}

func TestNats(t *testing.T) {
	var r cardinality.Replacer

	r = auto.NewNats()
	assert.Equal(t, "{inbox}", r.Replace("_INBOX.caWRiZOTBpF2Ol"))
	assert.Equal(t, "player.{partition}.update", r.Replace("player.1.update"))

	r = auto.NewNats(auto.WithoutInbox())
	assert.Equal(t, "_INBOX.{partition}", r.Replace("_INBOX.123"))

	r = auto.NewNats(auto.WithoutUrl())
	assert.Equal(t, "/goodle.{partition}", r.Replace("/goodle.123"))

	r = auto.NewNats(auto.WithoutPartition())
	assert.Equal(t, "update.123", r.Replace("update.123"))

	r = auto.NewNats(auto.WithoutUrl())
	assert.Equal(t, "/player/update/550e8400-e29b-41d4-a716-446655440000",
		r.Replace("/player/update/550e8400-e29b-41d4-a716-446655440000"))
}

func TestOverride(t *testing.T) {
	cfg := cardinality.NewConfig(
		cardinality.WithPathSeparator(false, "."),
		cardinality.WithPlaceholder(nil, func(id string) string {
			return fmt.Sprintf(`{%s}`, id)
		}),
	)
	r := auto.NewHttp(auto.WithConfigReader(cfg))
	assert.Equal(t, "player.update.{id}",
		r.Replace("player.update.123"))
}
