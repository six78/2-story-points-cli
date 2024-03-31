package view

import (
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	"waku-poker-planning/config"
	"waku-poker-planning/game"
	"waku-poker-planning/protocol"
)

func TestRenderSuite(t *testing.T) {
	suite.Run(t, new(RenderSuite))
}

type RenderSuite struct {
	suite.Suite
}

func (s *RenderSuite) SetupSuite() {
	logger, err := zap.NewDevelopment()
	s.Require().NoError(err)
	config.Logger = logger
}

func (s *RenderSuite) TestRenderCard() {
	testCases := []struct {
		value    protocol.VoteValue
		cursor   bool
		voted    bool
		expected string
	}{
		{protocol.VoteValue("1"), false, false, "     \n╭───╮\n│ 1 │\n╰───╯"},
		{protocol.VoteValue("2"), true, false, "     \n╭───╮\n│ 2 │\n╰───╯\n  ^  "},
		{protocol.VoteValue("3"), true, true, "╭───╮\n│ 3 │\n╰───╯\n     \n  ^  "},
		{protocol.VoteValue("4"), false, true, "╭───╮\n│ 4 │\n╰───╯"},
	}

	for _, tc := range testCases {
		result := renderCard(tc.value, tc.cursor, tc.voted)
		s.Require().Equal(tc.expected, result)
	}
}

func (s *RenderSuite) TestRenderDeck() {
	deck := game.CreateDeck([]string{"1", "2", "3"})
	result := renderDeck(deck, 2, true, "2")
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
