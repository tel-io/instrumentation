package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	th "github.com/stretchr/testify/http"
	"github.com/stretchr/testify/suite"
	"github.com/tel-io/tel/v2"
)

// want to checl if response will write to log via mw
const testString = "Hello World"

// want to check if it write to log via mw
const postContent = "XXX"

type Suite struct {
	suite.Suite

	tel   tel.Telemetry
	close func()

	buf *bytes.Buffer
}

func (s *Suite) SetupSuite() {
	c := tel.DefaultDebugConfig()
	c.LogLevel = "debug"
	c.OtelConfig.Enable = false

	s.tel, s.close = tel.New(context.Background(), c)
	s.buf = tel.SetLogOutput(&s.tel)
}

func (s *Suite) TearDownSuite() {
	s.close()
}

func (s *Suite) TearDownTest() {
	s.buf.Reset()
}

func TestHTTP(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestFilter() {
	var ok bool

	mw := NewServeMux(WithTel(&s.tel))
	mw.Handle("/ws", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))
	mw.Handle("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok = true
	}))

	ss := httptest.NewServer(mw)
	defer ss.Close()

	s.Run("health", func() {
		ok = false

		_, err := ss.Client().Get(ss.URL + "/health")
		s.NoError(err)

		s.True(ok)
		s.Empty(s.buf.Bytes())
	})

	s.Run("websocket", func() {
		ok = false

		b := bytes.NewReader(nil)
		r := httptest.NewRequest(http.MethodGet, ss.URL+"/ws", b)
		r.Header.Set("Upgrade", "websocket")
		r.RequestURI = ""

		_, err := ss.Client().Do(r)
		s.NoError(err)

		s.True(ok)
		s.Empty(s.buf.Bytes())
	})
}

func (s *Suite) TestContextConsistency() {
	// key value helps check if our middleware not damage already existent context with own values
	type key struct{}

	handler := http.Handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(testString))

		val := request.Context().Value(key{})

		s.Require().NotNil(val)
		s.Equal("*****", val.(string))
	}))

	handler = func(next http.Handler) http.Handler {
		fn := func(writer http.ResponseWriter, request *http.Request) {
			ctx := context.WithValue(request.Context(), key{}, "*****")
			next.ServeHTTP(writer, request.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}(handler)

	mw := NewServeMux(
		WithTel(&s.tel),
		WithDumpRequest(true),
		WithDumpResponse(true),
		WithDumpPayloadOnError(false),
	)

	mw.Handle("/test", handler)

	r, err := http.NewRequest(http.MethodGet, "/test", nil)
	s.Require().NoError(err)

	w := &th.TestResponseWriter{}

	mw.ServeHTTP(w, r)
}

func (s *Suite) TestHttpServerMiddlewareAll() {
	const url = "/test"

	tests := []struct {
		name string
		req  *http.Request
		body []byte

		hasRequest  bool
		hasResponse bool
		headers     http.Header
	}{
		{
			name:        "ALL",
			req:         NewRequest(http.MethodPost, url, bytes.NewBufferString(postContent)),
			body:        []byte(testString),
			hasResponse: true,
			hasRequest:  true,
		},
		{
			name:        "Request",
			req:         NewRequest(http.MethodPost, url, bytes.NewBufferString(postContent)),
			body:        []byte(testString),
			hasResponse: false,
			hasRequest:  true,
		},
		{
			name:        "Response",
			req:         NewRequest(http.MethodPost, url, bytes.NewBufferString(postContent)),
			body:        []byte(testString),
			hasResponse: true,
			hasRequest:  false,
		},
		{
			name:        "None",
			req:         NewRequest(http.MethodPost, url, bytes.NewBufferString(postContent)),
			body:        []byte(testString),
			hasResponse: false,
			hasRequest:  false,
		},
		{
			name:        "Headers",
			req:         NewRequest(http.MethodPost, url, bytes.NewBufferString(postContent)),
			body:        []byte(testString),
			hasResponse: false,
			hasRequest:  false,
			headers: map[string][]string{
				"IS":       {"OK"},
				"ONE MORE": {strings.Repeat("x", 10)},
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.buf.Reset()

			mw := NewServeMux(
				WithTel(&s.tel),
				WithDumpRequest(test.hasRequest),
				WithDumpResponse(test.hasResponse),
				WithDumpPayloadOnError(false), //
				WithHeaders(len(test.headers) > 0),
			)

			mw.Handle(url, writeHandlerBody([]byte(testString)))

			w := &th.TestResponseWriter{}

			// just set headers if exists for request
			if len(test.headers) > 0 {
				test.req.Header = test.headers
			}

			mw.ServeHTTP(w, test.req)

			// request check
			if test.hasRequest {
				s.Contains(s.buf.String(), postContent)
			} else {
				s.NotContains(s.buf.String(), postContent)
			}

			// response check
			if test.hasResponse {
				s.Contains(s.buf.String(), testString)
			} else {
				s.NotContains(s.buf.String(), testString)
			}

			// header check
			if len(test.headers) > 0 {
				s.Contains(s.buf.String(), keyHeader)
				for k, list := range test.headers {
					s.Contains(s.buf.String(), k)

					for _, val := range list {
						s.Contains(s.buf.String(), val)
					}
				}
			} else {
				s.NotContains(s.buf.String(), keyHeader)
			}
		})
	}
}

func NewRequest(method, url string, body io.Reader) *http.Request {
	r, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}

	return r
}

func writeHandlerBody(body []byte) http.Handler {
	return http.Handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(body)
	}))
}

// LogEncode: none, WithDumpRequest=true, WithDumpResponse=true, WithHeaders=true, WithDumpPayloadOnError=false
// BenchmarkMw-10    	   15398	    113394 ns/op
//
// LogEncode: none, WithDumpRequest=false, WithDumpResponse=false, WithHeaders=false, WithDumpPayloadOnError=false
// BenchmarkMw-10    	  237624	      6713 ns/op
//
// LogEncode: console, WithDumpRequest=false, WithDumpResponse=false, WithHeaders=false, WithDumpPayloadOnError=false
// BenchmarkMw-10    	  204169	      6709 ns/op
//
// LogEncode: json, WithDumpRequest=false, WithDumpResponse=false, WithHeaders=false, WithDumpPayloadOnError=false
// BenchmarkMw-10    	  211017	      6787 ns/op
func BenchmarkMw(b *testing.B) {
	c := tel.DefaultDebugConfig()
	c.LogLevel = "debug"
	c.LogEncode = "none"
	c.OtelConfig.Enable = false

	t, closer := tel.New(context.Background(), c)
	defer closer()

	mw := NewServeMux(
		WithTel(&t),
		WithDumpRequest(false),
		WithDumpResponse(false),
		WithHeaders(false),
		WithDumpPayloadOnError(false),
	)

	req := NewRequest(http.MethodPost, "/", bytes.NewBufferString(strings.Repeat("y", 100)))
	req.Header = map[string][]string{
		"TEST": {strings.Repeat("x", 100)},
	}

	mw.Handle("/", writeHandlerBody([]byte(strings.Repeat("x", 100))))

	for i := 0; i < b.N; i++ {
		w := &th.TestResponseWriter{}
		mw.ServeHTTP(w, req)
	}
}
