package pgx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

const (
	inxTrace  = 0
	idxLog    = 1
	idxRecord = 2
	idxMax    = 3
)

var (
	listStart = []int{inxTrace, idxLog, idxRecord}
	listEnd   = []int{idxLog, idxRecord, inxTrace}
)

const instrumentationName = "github.com/tel-io/instrumentation/plugins/pgx"

// TraceLog implements pgx.QueryTracer, pgx.BatchTracer, pgx.ConnectTracer, and pgx.CopyFromTracer. All fields are
// required.
type TraceLog struct {
	*config
	cb [idxMax]Callback
}

var _ pgx.BatchTracer = &TraceLog{}
var _ pgx.ConnectTracer = &TraceLog{}
var _ pgx.CopyFromTracer = &TraceLog{}
var _ pgx.PrepareTracer = &TraceLog{}
var _ pgx.QueryTracer = &TraceLog{}

func New(opts ...Option) (*TraceLog, error) {
	cfg := newConfig(opts...)

	tracer := newMethodTracer(&cfg.TraceConfig)

	rec, err := newRecorder(&cfg.RecordConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	logger := &methodLoggerImpl{cfg: &cfg.LoggerConfig}

	return &TraceLog{
		config: cfg,
		cb:     [idxMax]Callback{tracer, logger, rec}, // inxTrace, idxLog, idxRecord
	}, nil
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

	for _, idx := range listStart {
		var ff QueryFn

		cxt, ff = tl.cb[idx].Query(cxt, data)

		res.list = append(res.list, ff)
	}

	return context.WithValue(cxt, tracelogQueryCtxKey, res)
}

func (tl *TraceLog) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	queryData := ctx.Value(tracelogQueryCtxKey).(*traceQueryData)

	for _, idx := range listEnd {
		queryData.list[idx](conn, data)
	}
}

type traceBatchData struct {
	listA []BatchQueryFn
	listB []BatchEndFn
}

func (tl *TraceLog) TraceBatchStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceBatchStartData) context.Context {
	res := &traceBatchData{}

	cxt := ctx

	for _, idx := range listStart {
		var (
			ff1 BatchQueryFn
			ff2 BatchEndFn
		)

		cxt, ff1, ff2 = tl.cb[idx].Batch(cxt, data)

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

	for _, idx := range listEnd {
		queryData.listB[idx](conn, data)
	}
}

type traceCopyFromData struct {
	list []CopyFn
}

func (tl *TraceLog) TraceCopyFromStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromStartData) context.Context {
	res := &traceCopyFromData{}
	cxt := ctx

	for _, idx := range listStart {
		var ff CopyFn

		cxt, ff = tl.cb[idx].Copy(cxt, data)

		res.list = append(res.list, ff)
	}

	return context.WithValue(ctx, tracelogCopyFromCtxKey, res)
}

func (tl *TraceLog) TraceCopyFromEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
	copyFromData := ctx.Value(tracelogCopyFromCtxKey).(*traceCopyFromData)

	for _, idx := range listEnd {
		copyFromData.list[idx](conn, data)
	}
}

type traceConnectData struct {
	list []ConnectFn
}

func (tl *TraceLog) TraceConnectStart(ctx context.Context, data pgx.TraceConnectStartData) context.Context {
	res := &traceConnectData{}
	cxt := ctx

	for _, idx := range listStart {
		var ff ConnectFn

		cxt, ff = tl.cb[idx].Connect(cxt, data)

		res.list = append(res.list, ff)
	}

	return context.WithValue(ctx, tracelogConnectCtxKey, res)
}

func (tl *TraceLog) TraceConnectEnd(ctx context.Context, data pgx.TraceConnectEndData) {
	connectData := ctx.Value(tracelogConnectCtxKey).(*traceConnectData)

	for _, idx := range listEnd {
		connectData.list[idx](data)
	}
}

type tracePrepareData struct {
	list []PrepareFn
}

func (tl *TraceLog) TracePrepareStart(ctx context.Context, _ *pgx.Conn, data pgx.TracePrepareStartData) context.Context {
	res := &tracePrepareData{}
	cxt := ctx

	for _, idx := range listStart {
		var ff PrepareFn

		cxt, ff = tl.cb[idx].Prepare(cxt, data)

		res.list = append(res.list, ff)
	}

	return context.WithValue(ctx, tracelogPrepareCtxKey, res)
}

func (tl *TraceLog) TracePrepareEnd(ctx context.Context, conn *pgx.Conn, data pgx.TracePrepareEndData) {
	prepareData := ctx.Value(tracelogPrepareCtxKey).(*tracePrepareData)

	for _, idx := range listEnd {
		prepareData.list[idx](conn, data)
	}
}

func getPID(conn *pgx.Conn) uint32 {
	pgConn := conn.PgConn()
	if pgConn != nil {
		return pgConn.PID()
	}

	return 0
}
