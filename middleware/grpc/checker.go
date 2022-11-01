package grpc

import (
	"context"

	health "github.com/tel-io/tel/v2/monitoring/heallth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	state = attribute.Key("state")
)

type grpcClientChecker struct {
	conn *grpc.ClientConn
}

func NewGrpcClientChecker(conn *grpc.ClientConn) health.Checker {
	return &grpcClientChecker{conn: conn}
}

func (ch *grpcClientChecker) Check(ctx context.Context) health.ReportDocument {
	st := ch.conn.GetState()

	if st < connectivity.TransientFailure {
		return health.NewReport(ch.conn.Target(), true)
	}

	return health.NewReport(ch.conn.Target(), false, state.String(st.String()))
}
