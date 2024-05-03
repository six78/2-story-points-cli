package view

import (
	"github.com/stretchr/testify/suite"
	"testing"
	"waku-poker-planning/app"
	"waku-poker-planning/testcommon"
	"waku-poker-planning/view/states"
)

func TestModel(t *testing.T) {
	suite.Run(t, new(ModelSuite))
}

type ModelSuite struct {
	testcommon.Suite
}

func (s *ModelSuite) TestInitialModel() {
	a := app.NewApp()
	m := initialModel(a)

	s.Require().Equal(a, m.app)
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
