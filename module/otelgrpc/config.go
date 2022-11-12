// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package otelgrpc // import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
)

const (
	instrumentationName = "github.com/tel-io/otelgrpc"
)

// config represents the configuration options available
type config struct {
	Meter         metric.Meter
	MeterProvider metric.MeterProvider
	Labels        []attribute.KeyValue

	// grpc_server_handling_seconds metric
	Bucket []float64

	ServerHandledHistogramEnabled bool
}

// Option interface used for setting optional config properties.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// newConfig creates a new config struct and applies opts to it.
func newConfig(opts ...Option) *config {
	c := &config{
		MeterProvider: global.MeterProvider(),
	}
	for _, opt := range opts {
		opt.apply(c)
	}

	c.Meter = c.MeterProvider.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(SemVersion()),
	)

	return c
}

// WithMeterProvider specifies a meter provider to use for creating a meter.
// If none is specified, the global provider is used.
func WithMeterProvider(provider metric.MeterProvider) Option {
	return optionFunc(func(cfg *config) {
		if provider != nil {
			cfg.MeterProvider = provider
		}
	})
}

func WithConstLabels(l ...attribute.KeyValue) Option {
	return optionFunc(func(cfg *config) {
		cfg.Labels = l
	})
}

// WithBucket for grpc_server_handling_seconds metric
func WithBucket(bucket []float64) Option {
	return optionFunc(func(cfg *config) {
		cfg.Bucket = bucket
	})
}

func WithServerHandledHistogram(v bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.ServerHandledHistogramEnabled = v
	})
}
