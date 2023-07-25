package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCardinalityGrouper(t *testing.T) {
	g, err := NewRulesGrouper(make([]string, 100))
	assert.Error(t, err)
	assert.Nil(t, g)

	var list = CardinalityGrouperList{
		func() CardinalityGrouper {
			gP, errP := NewRulesGrouper([]string{
				"/false/:gameId",
				"/gameBySlug/:gameId/hello/:gameName",
				"/detail/:gameName",
				"/:system/logout",
			})
			assert.NoError(t, errP)
			return gP
		}(),
		NewAutoGrouper(),
	}

	tests := map[string]string{
		"/login":                                  "/login",
		"/game/logout":                            "/:system/logout",
		"/api/v1/games/favorite/true/save":        "/api/v1/games/favorite/:state/save",
		"/api/v1/games/favorite/false/save":       "/api/v1/games/favorite/:state/save",
		"/api/v1/gameBySlug/abc/diamond-luck":     "/api/v1/gameBySlug/:gameId/:gameName",
		"/api/v1/gameBySlug/abc/d/diamond-luck":   "/api/v1/gameBySlug/:gameId/d/:gameName",
		"/api/v1/tournaments/detail/diamond-luck": "/api/v1/tournaments/detail/:gameName",
		"/api/v1/promo/detail/pina-colada":        "/api/v1/promo/detail/:gameName",
		"/api/v1/promo/detail/diamond-luck":       "/api/v1/promo/detail/:gameName",
	}

	for url, exp := range tests {
		assert.Equal(t, exp, list.Apply(url))
	}
}
