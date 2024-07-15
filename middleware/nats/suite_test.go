package nats

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tel-io/tel/v2"
)

type Suite struct {
	suite.Suite

	tel   tel.Telemetry
	close func()

	buf *bytes.Buffer
}

func (s *Suite) SetupSuite() {
	c := tel.DefaultDebugConfig()
	c.LogLevel = "debug"
	c.OtelConfig.Enable = false

	s.tel, s.close = tel.New(context.Background(), c)
	s.buf = tel.SetLogOutput(&s.tel)
}

func (s *Suite) TearDownSuite() {
	s.close()
}

func (s *Suite) TearDownTest() {
	s.buf.Reset()
}

func TestInit(t *testing.T) {
	suite.Run(t, new(Suite))
}
