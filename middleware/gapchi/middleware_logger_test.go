package gaphttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testURL = "/test"

type testCase struct {
	getReq  func(url string) *http.Request
	postReq func(url string) *http.Request

	get  http.HandlerFunc
	post http.HandlerFunc
}

func TestMiddlewareLogger(t *testing.T) {
	tc := testCase{
		getReq: func(url string) *http.Request {
			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?%s=%s", url, "name", "test"), nil)
			require.Nil(t, err)

			return req
		},
		postReq: func(url string) *http.Request {
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(`{"id":"11"}`)))
			require.Nil(t, err)

			return req
		},
		get: func(writer http.ResponseWriter, request *http.Request) {
			require.EqualValues(t, "test", request.URL.Query().Get("name"))
			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(`"test": "test"`))
		},
		post: func(writer http.ResponseWriter, request *http.Request) {
			var i interface{}
			err := json.NewDecoder(request.Body).Decode(&i)
			require.Nil(t, err)

			writer.WriteHeader(http.StatusOK)
			writer.Write([]byte(`"test": "test"`))
		},
	}

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(NewMiddlewareLogger(zap.NewExample(), &MiddlewareOptions{EnabledLogger: true}))
	router.Get(testURL, tc.get)
	router.Post(testURL, tc.post)

	server := httptest.NewServer(router)

	cli := http.Client{Transport: NewMiddlewareRoundTrip(http.DefaultTransport, true, zap.NewExample())}
	go func() {
		reqs := []*http.Request{
			tc.getReq(fmt.Sprintf("%s%s", server.URL, testURL)),
			tc.postReq(fmt.Sprintf("%s%s", server.URL, testURL)),
		}

		for _, req := range reqs {
			resp, err := cli.Do(req)
			require.Nil(t, err)
			require.EqualValues(t, resp.StatusCode, http.StatusOK)
			b, err := io.ReadAll(resp.Body)
			require.Nil(t, err)
			require.EqualValues(t, string(b), `"test": "test"`)
		}
	}()

}
