package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/tel-io/instrumentation/middleware/nats/v2/natsprop"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer for subscribers implementing Middleware
type Tracer struct {
	nameFn NameFn
}

func NewTracer(fn NameFn) *Tracer {
	return &Tracer{nameFn: fn}
}

func (t *Tracer) apply(next MsgHandler) MsgHandler {
	return func(ctx context.Context, msg *nats.Msg) error {
		var (
			kind = extractBaggageKind(ctx)
			opr  = t.nameFn(kind, msg)
			attr = ExtractAttributes(msg, kind, true)
		)

		_, bg, spanContext := natsprop.Extract(ctx, msg)
		ctx = trace.ContextWithRemoteSpanContext(ctx, spanContext)
		ctx = baggage.ContextWithBaggage(ctx, bg)

		span, ctx := tel.StartSpanFromContext(ctx, opr,
			trace.WithSpanKind(convSpanToKind(kind)),
		)
		defer span.End(trace.WithStackTrace(true))

		tel.FromCtx(ctx).PutAttr(attr...)
		tel.UpdateTraceFields(ctx)

		natsprop.Inject(ctx, msg)

		err := next(ctx, msg)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// convert kind_of to tracers span kinds
func convSpanToKind(v string) trace.SpanKind {
	switch v {
	case KindSub:
		return trace.SpanKindConsumer
	case KindPub:
		return trace.SpanKindProducer
	case KindRequest, KindRespond:
		return trace.SpanKindClient
	case KindReply:
		return trace.SpanKindServer
	default:
		return trace.SpanKindUnspecified
	}
}
