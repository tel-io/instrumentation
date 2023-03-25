package pgx

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	sqlattribute "github.com/tel-io/instrumentation/plugins/otelsql/attribute"
)

type NameFormatter func(ctx context.Context, op string) string

type ErrorToSpanStatus func(err error) (codes.Code, string)

// methodTracer traces a sql method.
type methodTracer interface {
	// ShouldTrace checks whether it should trace a method and the given context has a parent span
	ShouldTrace(ctx context.Context) bool
	MustTrace(ctx context.Context) (context.Context, func(method string, err error))
	Trace(ctx context.Context) (context.Context, func(method string, err error))
}

type methodTracerImpl struct {
	*TraceConfig
}

func (t *methodTracerImpl) Trace(ctx context.Context) (context.Context, func(method string, err error)) {
	if !t.ShouldTrace(ctx) {
		return ctx, func(_ string, _ error) {}
	}

	return t.MustTrace(ctx)
}

func (t *methodTracerImpl) MustTrace(ctx context.Context) (context.Context, func(method string, err error)) {
	span, ctx := tel.StartSpanFromContext(ctx, t.NameFormatter(ctx, "in_progress"),
		trace.WithTimestamp(time.Now()),
		trace.WithSpanKind(trace.SpanKindClient),
	)

	attrs := make([]attribute.KeyValue, 0, len(t.DefaultAttributes)+1)

	attrs = append(attrs, t.DefaultAttributes...)

	return ctx, func(method string, err error) {
		code, desc := t.ErrorToStatus(err)

		span.SetName(t.NameFormatter(ctx, method))
		span.SetAttributes(attrs...)
		span.SetStatus(code, desc)

		attrs = append(attrs, semconv.DBOperationKey.String(method))

		if code == codes.Error {
			span.RecordError(err)
		}

		span.End(
			trace.WithTimestamp(time.Now()),
		)
	}
}

func (t *methodTracerImpl) Query(ctx context.Context, data pgx.TraceQueryStartData) (context.Context, QueryFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(conn *pgx.Conn, data pgx.TraceQueryEndData) {
		f(fmt.Sprintln("Query", data.CommandTag), data.Err)
	}
}

func (t *methodTracerImpl) Batch(ctx context.Context, data pgx.TraceBatchStartData) (context.Context, BatchQueryFn, BatchEndFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(conn *pgx.Conn, data pgx.TraceBatchQueryData) {},
		func(conn *pgx.Conn, data pgx.TraceBatchEndData) {
			f("Batch", data.Err)
		}
}

func (t *methodTracerImpl) Copy(ctx context.Context, data pgx.TraceCopyFromStartData) (context.Context, CopyFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
		f("CopyFrom", data.Err)
	}
}

func (t *methodTracerImpl) Connect(ctx context.Context, data pgx.TraceConnectStartData) (context.Context, ConnectFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(data pgx.TraceConnectEndData) {
		f("Connect", data.Err)
	}
}

func (t *methodTracerImpl) Prepare(ctx context.Context, data pgx.TracePrepareStartData) (context.Context, PrepareFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(conn *pgx.Conn, data pgx.TracePrepareEndData) {
		f("Prepare", data.Err)
	}
}

func (t *methodTracerImpl) ShouldTrace(ctx context.Context) bool {
	hasSpan := trace.SpanContextFromContext(ctx).IsValid() ||
		(tel.FromCtx(ctx).Span() != nil && tel.FromCtx(ctx).Span().IsRecording())

	return t.AllowRootTrace || hasSpan
}

func newMethodTracer(cfg *TraceConfig) *methodTracerImpl {
	return &methodTracerImpl{
		TraceConfig: cfg,
	}
}

func formatSpanName(_ context.Context, method string) string {
	var sb strings.Builder

	sb.Grow(len(method) + 4)
	sb.WriteString("pgx:")
	sb.WriteString(method)

	return sb.String()
}

func spanStatusFromError(err error) (codes.Code, string) {
	if err == nil {
		return codes.Ok, ""
	}

	return codes.Error, err.Error()
}

func spanStatusFromErrorIgnoreErrSkip(err error) (codes.Code, string) {
	if err == nil || errors.Is(err, driver.ErrSkip) {
		return codes.Ok, ""
	}

	return codes.Error, err.Error()
}

func traceNoQuery(context.Context, string, []driver.NamedValue) []attribute.KeyValue {
	return nil
}

func traceQueryWithoutArgs(_ context.Context, sql string, _ []driver.NamedValue) []attribute.KeyValue {
	return []attribute.KeyValue{
		semconv.DBStatementKey.String(sql),
	}
}

func traceQueryWithArgs(_ context.Context, sql string, args []driver.NamedValue) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 1+len(args))
	attrs = append(attrs, semconv.DBStatementKey.String(sql))

	for _, arg := range args {
		attrs = append(attrs, sqlattribute.FromNamedValue(arg))
	}

	return attrs
}
