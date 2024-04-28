package deckview

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"waku-poker-planning/config"
	"waku-poker-planning/game"
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/messages"
)

func TestDeckView(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	suite.Suite
}

func (s *Suite) SetupSuite() {
	logger, err := zap.NewDevelopment()
	s.Require().NoError(err)
	config.Logger = logger
}

func (s *Suite) TestRenderCard() {
	testCases := []struct {
		value    protocol.VoteValue
		cursor   bool
		voted    bool
		expected string
	}{
		{protocol.VoteValue("1"), false, false, "     \n╭───╮\n│ 1 │\n╰───╯\n     "},
		{protocol.VoteValue("2"), true, false, "     \n╭───╮\n│ 2 │\n╰───╯\n  ^  "},
		{protocol.VoteValue("3"), true, true, "╭───╮\n│ 3 │\n╰───╯\n     \n  ^  "},
		{protocol.VoteValue("4"), false, true, "╭───╮\n│ 4 │\n╰───╯\n     "},
	}

	for _, tc := range testCases {
		result := renderCard(tc.value, tc.cursor, false, tc.voted)
		s.Require().Equal(tc.expected, result)
	}
}

func (s *Suite) TestRenderDeck() {
	model := Model{
		deck:         game.CreateDeck([]string{"1", "2", "3"}),
		voteState:    protocol.VotingState,
		myVote:       protocol.VoteValue("2"),
		focused:      true,
		isDealer:     false,
		commandMode:  false,
		voteCursor:   2,
		finishCursor: 0,
	}

	result := model.View()
	expected := `
      ╭───╮       
╭───╮ │ 2 │ ╭───╮ 
│ 1 │ ╰───╯ │ 3 │ 
╰───╯       ╰───╯ 
              ^   
`
	// Remove leading and trailing newlines
	expected = expected[1 : len(expected)-1]

	s.Require().Equal(expected, result)
}

func (s *Suite) TestModelInit() {
	testCases := []struct {
		name    string
		focused bool
	}{{
		name:    "new model focused",
		focused: true,
	}, {
		name:    "new model blurred",
		focused: false,
	},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			model := New(tc.focused)
			cmd := model.Init()
			s.Require().Nil(cmd)
			s.Require().Equal(tc.focused, model.focused)
			s.Require().Zero(model.VoteCursor())
			s.Require().Zero(model.FinishCursor())
		})
	}
}

func (s *Suite) TestFocus() {
	model := New(false)
	model.Focus()
	s.Require().True(model.focused)
	model.Blur()
	s.Require().False(model.focused)
}

func (s *Suite) TestDecrementCursor() {
	testCases := []struct {
		given    int
		expected int
	}{
		{
			given:    0,
			expected: 0,
		},
		{
			given:    1,
			expected: 0,
		},
		{
			given:    2,
			expected: 1,
		},
	}

	for _, tc := range testCases {
		test := func(*testing.T) {
			model := New(true)
			result := model.decrementCursor(tc.given)
			s.Require().Equal(tc.expected, result)
		}
		testName := fmt.Sprintf("test decrement %d", tc.given)
		s.T().Run(testName, test)
	}
}

func (s *Suite) TestIncrementCursor() {
	testCases := []struct {
		given    int
		expected int
	}{
		{
			given:    0,
			expected: 1,
		},
		{
			given:    1,
			expected: 2,
		},
		{
			given:    2,
			expected: 2,
		},
	}

	for _, tc := range testCases {
		test := func(*testing.T) {
			model := New(true)
			model.deck = game.CreateDeck([]string{"a", "b", "c"})
			result := model.incrementCursor(tc.given)
			s.Require().Equal(tc.expected, result)
		}
		testName := fmt.Sprintf("test increment %d", tc.given)
		s.T().Run(testName, test)
	}
}

func (s *Suite) TestUpdate() {
	deck := make(protocol.Deck, 3)
	gofakeit.Slice(deck)

	model := New(false)
	_ = model.Init()

	model = model.Update(messages.GameStateMessage{
		State: nil,
	})
	
	s.Require().Equal(protocol.Deck{}, model.deck)
	s.Require().Equal(protocol.IdleState, model.voteState)

	model = model.Update(messages.GameStateMessage{
		State: &protocol.State{
			Deck:          deck,
			ActiveIssue:   "1",
			VotesRevealed: false,
		},
	})

	s.Require().Equal(deck, model.deck)
	s.Require().Equal(protocol.VotingState, model.voteState)

	model = model.Update(messages.RoomJoin{IsDealer: true})
	s.Require().True(model.isDealer)

	model = model.Update(messages.RoomJoin{IsDealer: false})
	s.Require().False(model.isDealer)

	model = model.Update(messages.CommandModeChange{CommandMode: true})
	s.Require().True(model.commandMode)

	model = model.Update(messages.CommandModeChange{CommandMode: false})
	s.Require().False(model.commandMode)

}
