package pgx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/tel-io/instrumentation/plugins/pgx"

// TraceLog implements pgx.QueryTracer, pgx.BatchTracer, pgx.ConnectTracer, and pgx.CopyFromTracer. All fields are
// required.
type TraceLog struct {
	*config
	cb []Callback
}

var _ pgx.BatchTracer = &TraceLog{}
var _ pgx.ConnectTracer = &TraceLog{}
var _ pgx.CopyFromTracer = &TraceLog{}
var _ pgx.PrepareTracer = &TraceLog{}
var _ pgx.QueryTracer = &TraceLog{}

func New(opts ...Option) (*TraceLog, error) {
	cfg := newConfig(opts...)

	tracer := newMethodTracer(
		cfg.tracerProvider.Tracer(instrumentationName,
			trace.WithSchemaURL(semconv.SchemaURL),
		),
		traceWithDefaultAttributes(cfg.defaultAttributes...),
		traceWithSpanNameFormatter(cfg.spanNameFormatter),
	)

	rec, err := newRecorder(cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	logger := &methodLoggerImpl{logger: cfg.loggerProvider, dumpSQL: cfg.dumpSQL}

	return &TraceLog{
		config: cfg,
		cb:     []Callback{tracer, rec, logger},
	}, nil
}

func newRecorder(opts *config) (Callback, error) {
	meter := opts.meterProvider.Meter(instrumentationName)

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

	return newMethodRecorder(latencyMsHistogram.Record, callsCounter.Add, opts.defaultAttributes...), nil
}

type ctxKey int

const (
	_ ctxKey = iota
	tracelogQueryCtxKey
	tracelogBatchCtxKey
	tracelogCopyFromCtxKey
	tracelogConnectCtxKey
	tracelogPrepareCtxKey
)

type traceQueryData struct {
	list []QueryFn
}

func (tl *TraceLog) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	res := &traceQueryData{}
	cxt := ctx

	for _, callback := range tl.cb {
		var ff QueryFn

		cxt, ff = callback.Query(cxt, data)

		res.list = append(res.list, ff)
	}

	return context.WithValue(cxt, tracelogQueryCtxKey, res)
}

func (tl *TraceLog) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	queryData := ctx.Value(tracelogQueryCtxKey).(*traceQueryData)

	for _, fn := range queryData.list {
		fn(conn, data)
	}
}

type traceBatchData struct {
	listA []BatchQueryFn
	listB []BatchEndFn
}

func (tl *TraceLog) TraceBatchStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	res := &traceBatchData{}

	cxt := ctx

	for _, callback := range tl.cb {
		var (
			ff1 BatchQueryFn
			ff2 BatchEndFn
		)

		cxt, ff1, ff2 = callback.Batch(cxt, data)

		res.listA = append(res.listA, ff1)
		res.listB = append(res.listB, ff2)
	}

	return context.WithValue(cxt, tracelogBatchCtxKey, res)
}

func (tl *TraceLog) TraceBatchQuery(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchQueryData) {
	queryData := ctx.Value(tracelogBatchCtxKey).(*traceBatchData)
	for _, fn := range queryData.listA {
		fn(conn, data)
	}
}

func (tl *TraceLog) TraceBatchEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchEndData) {
	queryData := ctx.Value(tracelogBatchCtxKey).(*traceBatchData)
	for _, fn := range queryData.listB {
		fn(conn, data)
	}
}

type traceCopyFromData struct {
	list []CopyFn
}

func (tl *TraceLog) TraceCopyFromStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context {
	res := &traceCopyFromData{}
	cxt := ctx

	for _, callback := range tl.cb {
		var ff CopyFn

		cxt, ff = callback.Copy(cxt, data)

		res.list = append(res.list, ff)
	}

	return context.WithValue(ctx, tracelogCopyFromCtxKey, res)
}

func (tl *TraceLog) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
	copyFromData := ctx.Value(tracelogCopyFromCtxKey).(*traceCopyFromData)

	for _, fn := range copyFromData.list {
		fn(conn, data)
	}
}

type traceConnectData struct {
	list []ConnectFn
}

func (tl *TraceLog) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	res := &traceConnectData{}
	cxt := ctx

	for _, callback := range tl.cb {
		var ff ConnectFn

		cxt, ff = callback.Connect(cxt, data)

		res.list = append(res.list, ff)
	}

	return context.WithValue(ctx, tracelogConnectCtxKey, res)
}

func (tl *TraceLog) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	connectData := ctx.Value(tracelogConnectCtxKey).(*traceConnectData)

	for _, fn := range connectData.list {
		fn(data)
	}
}

type tracePrepareData struct {
	list []PrepareFn
}

func (tl *TraceLog) TracePrepareStart(ctx context.Context, _ *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	res := &tracePrepareData{}
	cxt := ctx

	for _, callback := range tl.cb {
		var ff PrepareFn

		cxt, ff = callback.Prepare(cxt, data)

		res.list = append(res.list, ff)
	}

	return context.WithValue(ctx, tracelogPrepareCtxKey, res)
}

func (tl *TraceLog) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {
	prepareData := ctx.Value(tracelogPrepareCtxKey).(*tracePrepareData)

	for _, fn := range prepareData.list {
		fn(conn, data)
	}
}

func getPID(conn *pgx.Conn) uint32 {
	pgConn := conn.PgConn()
	if pgConn != nil {
		return pgConn.PID()
	}

	return 0
}
