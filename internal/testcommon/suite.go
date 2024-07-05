package testcommon

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/brianvoe/gofakeit/v6"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/six78/2-story-points-cli/internal/config"
)

type Suite struct {
	suite.Suite
	Ctx    context.Context
	cancel context.CancelFunc
	Logger *zap.Logger
	Clock  clockwork.FakeClock
}

func (s *Suite) SetupSuite() {
	s.Ctx, s.cancel = context.WithCancel(context.Background())
	s.Clock = clockwork.NewFakeClock()
	logger, err := zap.NewDevelopment()
	s.Require().NoError(err)
	s.Logger = logger
	config.Logger = logger
}

func (s *Suite) TearDownSuite() {
	s.cancel()
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

func (s *Suite) FakePayload() ([]byte, []byte) {
	payload := make([]byte, 10)
	gofakeit.Slice(&payload)

	jsonPayload, err := json.Marshal(payload)
	s.Require().NoError(err)

	return payload, jsonPayload
}
