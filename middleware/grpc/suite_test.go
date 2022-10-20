package grpc

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tel-io/tel/v2"
)

type Suite struct {
	suite.Suite

	byf    *bytes.Buffer
	tel    tel.Telemetry
	closer func()
}

func (s *Suite) TearDownSuite() {
	//	s.closer()
}

func (s *Suite) SetupSuite() {
	cfg := tel.DefaultDebugConfig()
	s.tel, s.closer = tel.New(context.Background(), cfg)
	s.byf = tel.SetLogOutput(&s.tel)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func TestAAA(t *testing.T) {
	mymap := map[string]string{"1": "2"}

	fmt.Println(mymap["1"])

}
