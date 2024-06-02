package testcommon

import (
	"2sp/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"reflect"
)

type Suite struct {
	suite.Suite
	Logger *zap.Logger
}

func (s *Suite) SetupSuite() {
	logger, err := zap.NewDevelopment()
	s.Require().NoError(err)
	s.Logger = logger
	config.Logger = logger
}

func (s *Suite) TearDownSuite() {
	_ = config.Logger.Sync()
}

func (s *Suite) SplitBatch(batch tea.Cmd) []tea.Cmd {
	s.Require().Equal(reflect.Func, reflect.TypeOf(batch).Kind())

	result := batch()
	s.Require().NotNil(result)

	batchMessage := result.(tea.BatchMsg)
	s.Require().NotNil(batchMessage)

	return batchMessage
}
