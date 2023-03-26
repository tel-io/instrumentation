package pgx

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
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

	cfg *RecordConfig
}

func newRecorder(cfg *RecordConfig) (Callback, error) {
	meter := cfg.meterProvider.Meter(instrumentationName)

	latencyMsHistogram, err := meter.SyncFloat64().Histogram(dbSQLClientLatencyMs,
		instrument.WithUnit(unit.Milliseconds),
		instrument.WithDescription(`The distribution of latencies of various calls in milliseconds`),
	)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	callsCounter, err := meter.SyncInt64().Counter(dbSQLClientCalls,
		instrument.WithUnit(unit.Dimensionless),
		instrument.WithDescription(`The number of various calls of methods`),
	)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return newMethodRecorder(latencyMsHistogram.Record, callsCounter.Add, cfg), nil
}

func (r *methodRecorderImpl) Record(ctx context.Context) func(method string, err error) {
	startTime := time.Now()

	attrs := make([]attribute.KeyValue, 0, len(r.cfg.DefaultAttributes)+2)

	attrs = append(attrs, r.cfg.DefaultAttributes...)

	return func(method string, err error) {
		elapsedTime := float64(time.Since(startTime).Milliseconds())

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

func (r *methodRecorderImpl) Query(ctx context.Context, start pgx.TraceQueryStartData) (context.Context, QueryFn) {
	cb := r.Record(ctx)

	return ctx, func(conn *pgx.Conn, data pgx.TraceQueryEndData) {
		cb(start.SQL, data.Err)
	}
}

func (r *methodRecorderImpl) Batch(ctx context.Context, start pgx.TraceBatchStartData) (context.Context, BatchQueryFn, BatchEndFn) {
	cb := r.Record(ctx)

	return ctx, func(conn *pgx.Conn, data pgx.TraceBatchQueryData) {},
		func(conn *pgx.Conn, data pgx.TraceBatchEndData) {
			cb("Batch", data.Err)
		}
}

func (r *methodRecorderImpl) Copy(ctx context.Context, data pgx.TraceCopyFromStartData) (context.Context, CopyFn) {
	cb := r.Record(ctx)

	return ctx, func(conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
		cb("CopyFrom", data.Err)
	}
}

func (r *methodRecorderImpl) Connect(ctx context.Context, data pgx.TraceConnectStartData) (context.Context, ConnectFn) {
	cb := r.Record(ctx)

	return ctx, func(data pgx.TraceConnectEndData) {
		cb("Connect", data.Err)
	}
}

func (r *methodRecorderImpl) Prepare(ctx context.Context, start pgx.TracePrepareStartData) (context.Context, PrepareFn) {
	cb := r.Record(ctx)

	return ctx, func(conn *pgx.Conn, data pgx.TracePrepareEndData) {
		method := fmt.Sprintf("Prepare: %s", start.SQL)

		cb(method, data.Err)
	}
}

func newMethodRecorder(
	latencyRecorder float64Recorder,
	callsCounter int64Counter,
	cfg *RecordConfig,
) *methodRecorderImpl {
	return &methodRecorderImpl{
		recordLatency: latencyRecorder,
		countCalls:    callsCounter,
		cfg:           cfg,
	}
}
