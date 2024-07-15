package nats

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/mock"
	natsmockery "github.com/tel-io/instrumentation/middleware/nats/v2/mocks"
)

func (s *Suite) TestCommonPublish_RequestMsgWithContext() {
	m := &natsmockery.NatsPublisher{}
	interceptor := MiddlewareChain(
		NewRecovery(),
		NewLogs(defaultOperationFn, true, true),
	)

	type fields struct {
		interceptor Interceptor
		Conn        natsPublisher
	}

	type args struct {
		msg *nats.Msg
		res *nats.Msg
		err error
	}

	tests := []struct {
		name     string
		fields   fields
		args     args
		contains []string
	}{
		{
			//2022-12-04T19:48:12+01:00	debug	nats/logs.go:48	NATS:RESPOND/PING	{"subject": "PING", "kind_of": "RESPOND", "reply": "INBOX_PONG", "payload": "REPLY_DATA", "duration": "57.375µs"}
			//2022-12-04T19:48:12+01:00	debug	nats/logs.go:48	NATS:PUB/PING	{"subject": "PING", "kind_of": "PUB", "reply": "PONG", "payload": "DEMO_DATA", "duration": "287.708µs"}
			"OK",
			fields{
				interceptor: interceptor,
				Conn:        m,
			},
			args{
				msg: &nats.Msg{Subject: "PING", Reply: "PONG", Data: []byte("DEMO_DATA")},
				res: &nats.Msg{Subject: "INBOX_PONG", Data: []byte("REPLY_DATA")},
				err: nil,
			},
			[]string{"PING", "PONG", "DEMO_DATA", "INBOX_PONG", "REPLY_DATA"},
		},
		{
			"err",
			fields{
				interceptor: interceptor,
				Conn:        m,
			},
			args{
				msg: &nats.Msg{Subject: "PING", Reply: "PONG", Data: []byte("DEMO_DATA")},
				res: nil,
				err: fmt.Errorf("some err"),
			},
			[]string{"PING", "PONG", "DEMO_DATA", "some err"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.buf.Reset()

			c := &CommonPublish{
				interceptor: tt.fields.interceptor,
				Conn:        tt.fields.Conn,
			}

			m.On("RequestMsgWithContext", mock.Anything, tt.args.msg).
				Return(tt.args.res, tt.args.err).Once()

			gotRes, err := c.RequestMsgWithContext(s.tel.Copy().Ctx(), tt.args.msg)
			s.Truef(err == tt.args.err, "err %v != err from arg %v", err, tt.args.err)
			s.Equal(tt.args.res, gotRes)

			for _, v := range tt.contains {
				s.Contains(s.buf.String(), v)
			}

			m.AssertExpectations(s.T())
		})
	}
}
