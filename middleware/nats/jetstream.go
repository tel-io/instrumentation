package nats

import "github.com/nats-io/nats.go"

type JetStreamContext struct {
	js nats.JetStreamContext

	*Core
}

// JS unwrap
func (j *JetStreamContext) JS() nats.JetStreamContext {
	return j.js
}

// Subscribe creates an async Subscription for JetStream.
// The stream and consumer names can be provided with the nats.Bind() option.
// For creating an ephemeral (where the consumer name is picked by the server),
// you can provide the stream name with nats.BindStream().
// If no stream name is specified, the library will attempt to figure out which
// stream the subscription is for. See important notes below for more details.
//
// IMPORTANT NOTES:
// * If none of the options Bind() nor Durable() are specified, the library will
// send a request to the server to create an ephemeral JetStream consumer,
// which will be deleted after an Unsubscribe() or Drain(), or automatically
// by the server after a short period of time after the NATS subscription is
// gone.
// * If Durable() option is specified, the library will attempt to lookup a JetStream
// consumer with this name, and if found, will bind to it and not attempt to
// delete it. However, if not found, the library will send a request to
// create such durable JetStream consumer. Note that the library will delete
// the JetStream consumer after an Unsubscribe() or Drain() only if it
// created the durable consumer while subscribing. If the durable consumer
// already existed prior to subscribing it won't be deleted.
// * If Bind() option is provided, the library will attempt to lookup the
// consumer with the given name, and if successful, bind to it. If the lookup fails,
// then the Subscribe() call will return an error.
func (j *JetStreamContext) Subscribe(subj string, cb MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return j.subMeter.Hook(
		j.js.Subscribe(subj, j.subWrap(cb), opts...),
	)
}

// QueueSubscribe creates a Subscription with a queue group.
// If no optional durable name nor binding options are specified, the queue name will be used as a durable name.
// See important note in Subscribe()
func (j *JetStreamContext) QueueSubscribe(subj, queue string, cb MsgHandler, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return j.subMeter.Hook(
		j.js.QueueSubscribe(subj, queue, j.subWrap(cb), opts...),
	)
}

// PullSubscribe creates a Subscription that can fetch messages.
// See important note in Subscribe(). Additionally, for an ephemeral pull consumer, the "durable" value must be
// set to an empty string.
//
// Only wrap Subscription for gather stats
func (j *JetStreamContext) PullSubscribe(subj, durable string, opts ...nats.SubOpt) (*nats.Subscription, error) {
	return j.subMeter.Hook(
		j.js.PullSubscribe(subj, durable, opts...),
	)
}
