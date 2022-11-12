// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package otelgrpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
)

type serverReporter struct {
	metrics     *ServerMetrics
	rpcType     grpcType
	serviceName string
	methodName  string
	startTime   time.Time
}

func newServerReporter(ctx context.Context, m *ServerMetrics, rpcType grpcType, fullMethod string) *serverReporter {
	r := &serverReporter{
		metrics: m,
		rpcType: rpcType,
	}

	if r.metrics.serverHandledHistogramEnabled {
		r.startTime = time.Now()
	}

	r.serviceName, r.methodName = splitMethodName(fullMethod)
	r.metrics.counters[serverStartedCounter].Add(ctx, 1,
		append(r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
		)...,
	)

	return r
}

func (r *serverReporter) ReceivedMessage(ctx context.Context) {
	r.metrics.counters[serverStreamMsgReceived].Add(ctx, 1,
		append(r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
		)...,
	)
}

func (r *serverReporter) SentMessage(ctx context.Context) {
	r.metrics.counters[serverStreamMsgSent].Add(ctx, 1,
		append(r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
		)...,
	)
}

func (r *serverReporter) Handled(ctx context.Context, code codes.Code) {
	r.metrics.counters[serverHandledCounter].Add(ctx, 1,
		append(r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
			attribute.String(AttrCode, code.String()),
		)...,
	)

	if r.metrics.serverHandledHistogramEnabled {
		dur := float64(time.Since(r.startTime).Seconds())

		r.metrics.valueRecorders[serverHandledHistogram].Record(ctx, dur,
			append(r.metrics.labels,
				attribute.String(AttrType, string(r.rpcType)),
				attribute.String(AttrService, r.serviceName),
				attribute.String(AttrMethod, r.methodName),
			)...,
		)
	}
}
