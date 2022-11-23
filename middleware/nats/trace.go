package nats

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/tel-io/instrumentation/middleware/nats/natsprop"
	"github.com/tel-io/tel/v2"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

// Tracer for subscribers implementing Middleware
type Tracer struct {
	subNameFn NameFn
}

func NewTracer(fn NameFn) *Tracer {
	return &Tracer{subNameFn: fn}
}

func (t *Tracer) apply(next MsgHandler) MsgHandler {
	return func(cxt context.Context, msg *nats.Msg) error {
		opr := t.subNameFn(msg)

		extract, bg, spanContext := natsprop.Extract(cxt, msg)
		cxt = trace.ContextWithRemoteSpanContext(cxt, spanContext)
		cxt = baggage.ContextWithBaggage(cxt, bg)

		span, ctx := tel.StartSpanFromContext(cxt, opr)
		defer span.End()

		tel.FromCtx(ctx).PutAttr(extract...)
		tel.UpdateTraceFields(ctx)

		err := next(ctx, msg)
		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}

// PubTrace handle trace handling implementing PubMiddleware
type PubTrace struct {
	pubNameFn NameFn
	root      Publish
}

var _ PubMiddleware = &PubTrace{}

func NewPubTrace(fn NameFn) *PubTrace {
	return &PubTrace{
		pubNameFn: fn,
	}
}

func (p *PubTrace) apply(in PubMiddleware) PubMiddleware {
	p.root = in

	return p
}

func (p *PubTrace) PublishMsgWithContext(cxt context.Context, msg *nats.Msg) (err error) {
	opr := p.pubNameFn(msg)

	extract, bg, spanContext := natsprop.Extract(cxt, msg)
	cxt = trace.ContextWithRemoteSpanContext(cxt, spanContext)
	cxt = baggage.ContextWithBaggage(cxt, bg)

	span, ctx := tel.StartSpanFromContext(cxt, opr)

	tel.FromCtx(ctx).PutAttr(extract...)

	natsprop.Inject(ctx, msg)

	defer func() {
		if err != nil {
			span.RecordError(err)
		}

		span.End()
	}()

	return p.root.PublishMsgWithContext(ctx, msg)
}

func (p *PubTrace) PublishWithContext(cxt context.Context, subj string, data []byte) (err error) {
	msg := &nats.Msg{Data: data, Subject: subj}

	return p.PublishMsgWithContext(cxt, msg)
}

func (p *PubTrace) RequestMsgWithContext(cxt context.Context, msg *nats.Msg) (resp *nats.Msg, err error) {
	opr := p.pubNameFn(msg)

	extract, bg, spanContext := natsprop.Extract(cxt, msg)
	cxt = trace.ContextWithRemoteSpanContext(cxt, spanContext)
	cxt = baggage.ContextWithBaggage(cxt, bg)

	span, ctx := tel.StartSpanFromContext(cxt, opr)

	tel.FromCtx(ctx).PutAttr(extract...)

	natsprop.Inject(ctx, msg)

	defer func() {
		if err != nil {
			span.RecordError(err)
		}

		span.End()
	}()

	return p.root.RequestMsgWithContext(ctx, msg)
}

func (p *PubTrace) RequestWithContext(ctx context.Context, subj string, data []byte) (*nats.Msg, error) {
	msg := &nats.Msg{Data: data, Subject: subj}

	return p.RequestMsgWithContext(ctx, msg)
}

// PublishRequestWithContext we just use PublishMsgWithContext to handle publish with reply option
func (p *PubTrace) PublishRequestWithContext(ctx context.Context, subj, reply string, data []byte) error {
	msg := &nats.Msg{Data: data, Subject: subj, Reply: reply}

	return p.PublishMsgWithContext(ctx, msg)
}
