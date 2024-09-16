package gaphttp

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

func TestMiddlewareMetrics(t *testing.T) {
	tc := testCase{
		postReq: func(url string) *http.Request {
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(`{"id":"11"}`)))
			require.Nil(t, err)

			return req
		},
		post: func(writer http.ResponseWriter, request *http.Request) {},
	}

	router := chi.NewRouter()
	metrics := newMetricsTest()
	router.Use(NewMiddlewareMetrics(metrics, &MiddlewareOptions{ServiceName: "test", EnabledMetrics: true}))
	router.Post(testURL, tc.post)

	server := httptest.NewServer(router)

	reqs := []*http.Request{
		tc.postReq(fmt.Sprintf("%s%s", server.URL, testURL)),
	}

	// Attention!   NewMiddlewareRoundTrip
	cli := http.Client{Transport: NewMiddlewareRoundTrip(http.DefaultTransport, true, zap.NewExample())}
	for _, req := range reqs {
		_, err := cli.Do(req)
		require.Nil(t, err)
	}

	require.Len(t, metrics.data, 4)
}

type metricsTest struct {
	locker *sync.Mutex
	data   map[string]struct {
		Value interface{}
		Attrs []attribute.KeyValue
	}
}

func newMetricsTest() *metricsTest {
	return &metricsTest{
		locker: &sync.Mutex{},
		data: make(map[string]struct {
			Value interface{}
			Attrs []attribute.KeyValue
		}),
	}
}

func (m *metricsTest) IncRequestCount(ctx context.Context, attrs ...attribute.KeyValue) {
	m.locker.Lock()
	defer m.locker.Unlock()

	m.data[RequestCount] = struct {
		Value interface{}
		Attrs []attribute.KeyValue
	}{Value: 1, Attrs: attrs}
}

func (m *metricsTest) AddRequestContentLength(ctx context.Context, len int64, attrs ...attribute.KeyValue) {
	m.locker.Lock()
	defer m.locker.Unlock()

	m.data[RequestContentLength] = struct {
		Value interface{}
		Attrs []attribute.KeyValue
	}{Value: len, Attrs: attrs}
}

func (m *metricsTest) AddResponseContentLength(ctx context.Context, len int64, attrs ...attribute.KeyValue) {
	m.locker.Lock()
	defer m.locker.Unlock()

	m.data[ResponseContentLength] = struct {
		Value interface{}
		Attrs []attribute.KeyValue
	}{Value: len, Attrs: attrs}
}

func (m *metricsTest) AddHandleDuration(ctx context.Context, dur float64, attrs ...attribute.KeyValue) {
	m.locker.Lock()
	defer m.locker.Unlock()

	m.data[ServerLatency] = struct {
		Value interface{}
		Attrs []attribute.KeyValue
	}{Value: dur, Attrs: attrs}
}
