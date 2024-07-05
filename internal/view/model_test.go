package view

import (
	"reflect"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/six78/2-story-points-cli/internal/testcommon"
	mocktransport "github.com/six78/2-story-points-cli/internal/transport/mock"
	"github.com/six78/2-story-points-cli/internal/view/messages"
	"github.com/six78/2-story-points-cli/internal/view/states"
	"github.com/six78/2-story-points-cli/pkg/game"
)

func TestModel(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	testcommon.Suite
	game      *game.Game
	transport mocktransport.MockService
}

func (s *ModelSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.transport = mocktransport.NewMockService(ctrl)

	options := []game.Option{
		game.WithContext(s.Ctx),
		game.WithTransport(s.transport),
		game.WithClock(s.Clock),
		game.WithLogger(s.Logger),
		game.WithPlayerName(gofakeit.Username()),
		game.WithPublishStateLoop(false),
	}

	s.game = game.NewGame(options)
	s.Require().NotNil(s.game)

	err := s.game.Initialize()
	s.Require().NoError(err)
}

func (s *ModelSuite) TeardownTest() {
	s.game.Stop()
	s.game = nil
}

func (s *ModelSuite) TestInitialModel() {
	m := InitialModel(s.game, s.transport)

	s.Require().Equal(s.game, m.game)
	s.Require().Equal(s.transport, m.transport)
	s.Require().Equal(states.Initializing, m.state)
	s.Require().Nil(m.gameState)
	s.Require().Empty(m.roomID)
	s.Require().False(m.commandMode)
	s.Require().Equal(states.ActiveIssueView, m.roomViewState)
	s.Require().NotNil(m.input)
	s.Require().NotNil(m.spinner)
	s.Require().NotNil(m.errorView)
	s.Require().NotNil(m.playersView)
	s.Require().NotNil(m.shortcutsView)
	s.Require().NotNil(m.wakuStatusView)
	s.Require().NotNil(m.deckView)
	s.Require().NotNil(m.issueView)
	s.Require().NotNil(m.issuesListView)
	s.Require().False(m.disableEnterKey)
	s.Require().Nil(m.disableEnterRestart)
}

func (s *ModelSuite) TestUpdateEmpty() {
	m := tea.Model(InitialModel(s.game, s.transport))
	_ = m.Init()

	m2, cmd := m.Update(nil)
	s.Require().Nil(cmd)
	s.Require().NotNil(m2.(model))

	eq := reflect.DeepEqual(m, m2)
	s.Require().True(eq)
}

func (s *ModelSuite) TestUpdateFatalErrorMessage() {
	m := InitialModel(s.game, s.transport)
	_ = m.Init()

	err := gofakeit.Error()
	msg := messages.FatalErrorMessage{Err: err}

	m2, cmd := m.Update(msg)
	s.Require().Nil(cmd)
	s.Require().Equal(err, m2.(model).fatalError)
}

//func (s *ModelSuite) TestUpdateInitializingFinishedMessage() {
//	m := InitialModel(s.game, s.transport)
//	_ = m.Init()
//
//	s.transport.EXPECT().SubscribeToConnectionStatus().Times()
//
//	msg := messages.AppStateFinishedMessage{
//		State: states.Initializing,
//	}
//
//	m2, cmd := m.Update(msg)
//
//	s.Require().Empty(s.game.Player().Name)
//	s.Require().Equal(states.InputPlayerName, m2.(model).state)
//
//	s.Require().NotNil(cmd)
//	batchMessage := s.SplitBatch(cmd)
//
//	s.Require().Len(batchMessage, 4)
//
//	// TODO: Check messages content
//}
