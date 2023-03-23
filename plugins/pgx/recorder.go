package pgx

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

const (
	dbSQLClientLatencyMs = "db.sql.client.latency"
	dbSQLClientCalls     = "db.sql.client.calls"

	dbSQLConnectionsOpen           = "db.sql.connections.open"
	dbSQLConnectionsIdle           = "db.sql.connections.idle"
	dbSQLConnectionsActive         = "db.sql.connections.active"
	dbSQLConnectionsWaitCount      = "db.sql.connections.wait_count"
	dbSQLConnectionsWaitDuration   = "db.sql.connections.wait_duration"
	dbSQLConnectionsIdleClosed     = "db.sql.connections.idle_closed"
	dbSQLConnectionsLifetimeClosed = "db.sql.connections.lifetime_closed"
)

// float64Recorder adds a new value to the list of Histogram's records.
type float64Recorder = func(ctx context.Context, value float64, labels ...attribute.KeyValue)

// int64Counter adds the value to the counter's sum.
type int64Counter = func(ctx context.Context, value int64, labels ...attribute.KeyValue)

// methodRecorder records metrics about a sql method.
type methodRecorder interface {
	Record(ctx context.Context) func(method string, err error)
}

type methodRecorderImpl struct {
	recordLatency float64Recorder
	countCalls    int64Counter

	attributes []attribute.KeyValue
}

func (r *methodRecorderImpl) Record(ctx context.Context) func(method string, err error) {
	startTime := time.Now()

	attrs := make([]attribute.KeyValue, 0, len(r.attributes)+2)

	attrs = append(attrs, r.attributes...)

	return func(method string, err error) {
		elapsedTime := millisecondsSince(startTime)

		attrs = append(attrs, semconv.DBOperationKey.String(method))

		if err == nil {
			attrs = append(attrs, dbSQLStatusOK)
		} else {
			attrs = append(attrs, dbSQLStatusERROR,
				dbSQLError.String(err.Error()),
			)
		}

		r.countCalls(ctx, 1, attrs...)
		r.recordLatency(ctx, elapsedTime, attrs...)
	}
}

func (r *methodRecorderImpl) Query(ctx context.Context, data pgx.TraceQueryStartData) (context.Context, QueryFn) {
	cb := r.Record(ctx)

	return ctx, func(conn *pgx.Conn, data pgx.TraceQueryEndData) {
		cb("SQL Query", data.Err)
	}
}

func (r *methodRecorderImpl) Batch(ctx context.Context, start pgx.TraceBatchStartData) (context.Context, BatchQueryFn, BatchEndFn) {
	cb := r.Record(ctx)

	return ctx, func(conn *pgx.Conn, data pgx.TraceBatchQueryData) {},
		func(conn *pgx.Conn, data pgx.TraceBatchEndData) {
			cb("SQL Batch", data.Err)
		}
}

func (r *methodRecorderImpl) Copy(ctx context.Context, data pgx.TraceCopyFromStartData) (context.Context, CopyFn) {
	cb := r.Record(ctx)

	return ctx, func(conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
		cb("SQL CopyFrom", data.Err)
	}
}

func (r *methodRecorderImpl) Connect(ctx context.Context, data pgx.TraceConnectStartData) (context.Context, ConnectFn) {
	cb := r.Record(ctx)

	return ctx, func(data pgx.TraceConnectEndData) {
		cb("SQL Connect", data.Err)
	}
}

func (r *methodRecorderImpl) Prepare(ctx context.Context, data pgx.TracePrepareStartData) (context.Context, PrepareFn) {
	cb := r.Record(ctx)

	return ctx, func(conn *pgx.Conn, data pgx.TracePrepareEndData) {
		cb("SQL Prepare", data.Err)
	}
}

func newMethodRecorder(
	latencyRecorder float64Recorder,
	callsCounter int64Counter,
	attrs ...attribute.KeyValue,
) *methodRecorderImpl {
	return &methodRecorderImpl{
		recordLatency: latencyRecorder,
		countCalls:    callsCounter,
		attributes:    attrs,
	}
}
