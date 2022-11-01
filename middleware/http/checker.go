package http

import (
	"context"
	"net/http"
	"time"

	health "github.com/tel-io/tel/v2/monitoring/heallth"
	"go.opentelemetry.io/otel/attribute"
)

var code = attribute.Key("code")
var clientError = code.String(http.StatusText(http.StatusBadRequest))

type Checker struct {
	URL     string
	Timeout time.Duration
}

func NewChecker(url string) Checker {
	return Checker{URL: url, Timeout: 5 * time.Second}
}

func NewDomainWithTimeout(url string, timeout time.Duration) Checker {
	return Checker{
		URL: url, Timeout: timeout,
	}
}

func (u Checker) Check(ctx context.Context) health.ReportDocument {
	client := http.Client{
		Timeout: u.Timeout,
	}

	//nolint: noctx
	resp, err := client.Head(u.URL)

	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return health.NewReport(u.URL, false, clientError)
	}

	if resp.StatusCode < http.StatusInternalServerError {
		return health.NewReport(u.URL, true)
	}

	return health.NewReport(u.URL, false, code.String(http.StatusText(resp.StatusCode)))
}
