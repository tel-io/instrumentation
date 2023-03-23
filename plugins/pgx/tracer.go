package pgx

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	sqlattribute "github.com/tel-io/instrumentation/plugins/otelsql/attribute"
)

type SpanNameFormatter func(ctx context.Context, op string) string

type errorToSpanStatus func(err error) (codes.Code, string)

type queryTracer func(ctx context.Context, query string, args []driver.NamedValue) []attribute.KeyValue

// methodTracer traces a sql method.
type methodTracer interface {
	// ShouldTrace checks whether it should trace a method and the given context has a parent span
	ShouldTrace(ctx context.Context) (bool, bool)
	MustTrace(ctx context.Context) (context.Context, func(method string, err error))
	Trace(ctx context.Context) (context.Context, func(method string, err error))
}

type methodTracerImpl struct {
	tracer trace.Tracer

	formatSpanName SpanNameFormatter
	errorToStatus  func(err error) (codes.Code, string)
	allowRoot      bool
	attributes     []attribute.KeyValue
}

func (t *methodTracerImpl) Trace(ctx context.Context) (context.Context, func(method string, err error)) {
	shouldTrace, hasParentSpan := t.ShouldTrace(ctx)

	if !shouldTrace {
		return ctx, func(_ string, _ error) {}
	}

	newCtx, end := t.MustTrace(ctx)

	if !hasParentSpan {
		ctx = newCtx
	}

	return ctx, end
}
func (t *methodTracerImpl) MustTrace(ctx context.Context) (context.Context, func(method string, err error)) {
	ctx, span := t.tracer.Start(ctx, t.formatSpanName(ctx, "in_progress"),
		trace.WithTimestamp(time.Now()),
		trace.WithSpanKind(trace.SpanKindClient),
	)

	attrs := make([]attribute.KeyValue, 0, len(t.attributes)+1)

	attrs = append(attrs, t.attributes...)

	return ctx, func(method string, err error) {
		code, desc := t.errorToStatus(err)

		span.SetName(t.formatSpanName(ctx, method))
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
		f(fmt.Sprintln("SQL Query", data.CommandTag), data.Err)
	}
}

func (t *methodTracerImpl) Batch(ctx context.Context, data pgx.TraceBatchStartData) (context.Context, BatchQueryFn, BatchEndFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(conn *pgx.Conn, data pgx.TraceBatchQueryData) {},
		func(conn *pgx.Conn, data pgx.TraceBatchEndData) {
			f("SQL Batch", data.Err)
		}
}

func (t *methodTracerImpl) Copy(ctx context.Context, data pgx.TraceCopyFromStartData) (context.Context, CopyFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(conn *pgx.Conn, data pgx.TraceCopyFromEndData) {
		f("SQL CopyFrom", data.Err)
	}
}

func (t *methodTracerImpl) Connect(ctx context.Context, data pgx.TraceConnectStartData) (context.Context, ConnectFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(data pgx.TraceConnectEndData) {
		f("SQL Connect", data.Err)
	}
}

func (t *methodTracerImpl) Prepare(ctx context.Context, data pgx.TracePrepareStartData) (context.Context, PrepareFn) {
	cxt, f := t.Trace(ctx)

	return cxt, func(conn *pgx.Conn, data pgx.TracePrepareEndData) {
		f("SQL Prepare", data.Err)
	}
}

func (t *methodTracerImpl) ShouldTrace(ctx context.Context) (bool, bool) {
	hasSpan := trace.SpanContextFromContext(ctx).IsValid()

	return t.allowRoot || hasSpan, hasSpan
}

func newMethodTracer(tracer trace.Tracer, opts ...func(t *methodTracerImpl)) *methodTracerImpl {
	t := &methodTracerImpl{
		tracer:         tracer,
		formatSpanName: formatSpanName,
		errorToStatus:  spanStatusFromError,
	}

	for _, o := range opts {
		o(t)
	}

	return t
}

func tracerOrNil(t methodTracer, shouldTrace bool) methodTracer {
	if shouldTrace {
		return t
	}

	return nil
}

func traceWithAllowRoot(allow bool) func(t *methodTracerImpl) {
	return func(t *methodTracerImpl) {
		t.allowRoot = allow
	}
}

func traceWithDefaultAttributes(attrs ...attribute.KeyValue) func(t *methodTracerImpl) {
	return func(t *methodTracerImpl) {
		t.attributes = append(t.attributes, attrs...)
	}
}

func traceWithSpanNameFormatter(f SpanNameFormatter) func(t *methodTracerImpl) {
	return func(t *methodTracerImpl) {
		t.formatSpanName = f
	}
}

func formatSpanName(_ context.Context, method string) string {
	var sb strings.Builder

	sb.Grow(len(method) + 4)
	sb.WriteString("sql:")
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
