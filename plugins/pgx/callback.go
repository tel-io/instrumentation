package pgx

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type (
	QueryFn      func(*pgx.Conn, pgx.TraceQueryEndData)
	BatchQueryFn func(*pgx.Conn, pgx.TraceBatchQueryData)
	BatchEndFn   func(*pgx.Conn, pgx.TraceBatchEndData)
	CopyFn       func(*pgx.Conn, pgx.TraceCopyFromEndData)
	ConnectFn    func(pgx.TraceConnectEndData)
	PrepareFn    func(*pgx.Conn, pgx.TracePrepareEndData)
)

type Callback interface {
	Query(context.Context, pgx.TraceQueryStartData) (context.Context, QueryFn)
	Batch(context.Context, pgx.TraceBatchStartData) (context.Context, BatchQueryFn, BatchEndFn)
	Copy(context.Context, pgx.TraceCopyFromStartData) (context.Context, CopyFn)
	Connect(context.Context, pgx.TraceConnectStartData) (context.Context, ConnectFn)
	Prepare(context.Context, pgx.TracePrepareStartData) (context.Context, PrepareFn)
}
