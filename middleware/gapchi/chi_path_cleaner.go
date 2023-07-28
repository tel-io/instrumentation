package gaphttp

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func RemoveChiPathParam(request *http.Request) string {
	path := request.URL.Path
	if rctx := chi.RouteContext(request.Context()); rctx != nil {
		for i, v := range rctx.URLParams.Values {
			path = strings.Replace(path, v, rctx.URLParams.Keys[i], 1)
		}
	}

	return path
}
