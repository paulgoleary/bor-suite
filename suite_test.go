package bor_suite

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type testSuite struct {
	BorServerTestSuite
}

func TestTenantTestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestBasic() {
	// TODO !!!
}
