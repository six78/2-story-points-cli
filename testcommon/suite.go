package testcommon

import (
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"waku-poker-planning/config"
)

type Suite struct {
	suite.Suite
}

func (s *Suite) SetupSuite() {
	logger, err := zap.NewDevelopment()
	s.Require().NoError(err)
	config.Logger = logger
}
