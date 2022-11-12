// Copyright 2016 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package otelgrpc

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
)

type clientReporter struct {
	metrics     *ClientMetrics
	rpcType     grpcType
	serviceName string
	methodName  string
	startTime   time.Time
}

func newClientReporter(ctx context.Context, m *ClientMetrics, rpcType grpcType, fullMethod string) *clientReporter {
	r := &clientReporter{
		metrics: m,
		rpcType: rpcType,
	}

	if r.metrics.clientHandledHistogramEnabled {
		r.startTime = time.Now()
	}

	r.serviceName, r.methodName = splitMethodName(fullMethod)
	r.metrics.counters[clientStartedCounter].Add(ctx, 1,
		append(
			r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
		)...,
	)

	return r
}

func (r *clientReporter) ReceiveMessageTimer(ctx context.Context, startTime time.Time) {
	r.startTimer(ctx, clientStreamRecvHistogram, startTime)
}

func (r *clientReporter) ReceivedMessage(ctx context.Context) {
	r.metrics.counters[clientStreamMsgReceived].Add(ctx, 1,
		append(r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
		)...,
	)
}

func (r *clientReporter) SendMessageTimer(ctx context.Context, startTime time.Time) {
	r.startTimer(ctx, clientStreamSendHistogram, startTime)
}

func (r *clientReporter) SentMessage(ctx context.Context) {
	r.metrics.counters[clientStreamMsgSent].Add(ctx, 1,
		append(r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
		)...,
	)
}

func (r *clientReporter) Handled(ctx context.Context, code codes.Code) {
	r.metrics.counters[clientHandledCounter].Add(ctx, 1,
		append(r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
			attribute.String(AttrCode, code.String()),
		)...,
	)

	if r.metrics.clientHandledHistogramEnabled {
		r.startTimer(ctx, clientHandledHistogram, r.startTime)
	}
}

func (r *clientReporter) startTimer(ctx context.Context, metric string, startTime time.Time) {
	if !r.metrics.clientHandledHistogramEnabled {
		return
	}

	dur := float64(time.Since(startTime).Seconds())

	r.metrics.valueRecorders[metric].Record(ctx, dur,
		append(r.metrics.labels,
			attribute.String(AttrType, string(r.rpcType)),
			attribute.String(AttrService, r.serviceName),
			attribute.String(AttrMethod, r.methodName),
		)...,
	)
}
