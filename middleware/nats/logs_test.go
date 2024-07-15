package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
)

func (s *Suite) TestNewLogs() {
	// mw constructor
	type args struct {
		fn                 NameFn
		dumpPayloadOnError bool
		dumpRequest        bool
	}

	// callback vars
	type handlers struct {
		msg       *nats.Msg
		kindOf    string
		returnErr error
	}

	tests := []struct {
		name     string
		args     args
		handlers handlers
		contains []string
		exclude  []string
	}{
		{
			"default - no error",
			args{
				fn:                 defaultOperationFn,
				dumpPayloadOnError: true,
				dumpRequest:        false,
			},
			handlers{
				msg:       &nats.Msg{Subject: "XXX", Reply: "YYYY", Data: []byte("DEMO_DATA")},
				kindOf:    KindSub,
				returnErr: nil,
			},
			[]string{"XXX", "YYYY"},
			[]string{"DEMO_DATA"},
		},
		{
			// dump payload
			"default - error",
			args{
				fn:                 defaultOperationFn,
				dumpPayloadOnError: true,
				dumpRequest:        false,
			},
			handlers{
				msg:       &nats.Msg{Subject: "XXX", Reply: "YYYY", Data: []byte("DEMO_DATA")},
				kindOf:    KindSub,
				returnErr: fmt.Errorf("some error"),
			},
			[]string{"XXX", "YYYY", KindSub, "DEMO_DATA", "some error"},
			[]string{},
		},
		{
			// dump payload
			"dump no error",
			args{
				fn:                 defaultOperationFn,
				dumpPayloadOnError: true,
				dumpRequest:        true,
			},
			handlers{
				msg:       &nats.Msg{Subject: "XXX", Reply: "YYYY", Data: []byte("DEMO_DATA")},
				kindOf:    KindSub,
				returnErr: nil,
			},
			[]string{"XXX", "YYYY", KindSub, "DEMO_DATA"},
			[]string{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.buf.Reset()

			// any function should notify about kind of message
			ctx := WrapKindOfContext(s.tel.Ctx(), tt.handlers.kindOf)

			got := NewLogs(tt.args.fn, tt.args.dumpPayloadOnError, tt.args.dumpRequest)
			_ = got.apply(func(ctx context.Context, msg *nats.Msg) error {
				return tt.handlers.returnErr
			})(ctx, tt.handlers.msg)

			for _, v := range tt.contains {
				s.Contains(s.buf.String(), v)
			}

			for _, v := range tt.exclude {
				s.NotContains(s.buf.String(), v)
			}
		})
	}
}
