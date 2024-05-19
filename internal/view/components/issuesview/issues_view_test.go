package issuesview

import (
	"2sp/internal/testcommon"
	"2sp/internal/view/messages"
	protocol2 "2sp/pkg/protocol"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/stretchr/testify/suite"
	"strings"
	"testing"
)

func TestIssuesView(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	testcommon.Suite
}

func (s *Suite) TestInit() {
	model := New()
	s.Require().False(model.commandMode)
	s.Require().False(model.isDealer)
	s.Require().Empty(model.issues)
	s.Require().Empty(model.activeIssue)
	s.Require().NotNil(model.cursor)
	s.Require().NotNil(model.spinner)
	s.Require().Equal(spinner.MiniDot, model.spinner.Spinner)

	cmd := model.Init()
	s.NotNil(cmd)
}

func (s *Suite) TestUpdateCommandMode() {
	model := New()
	s.False(model.commandMode)

	model, _ = model.Update(messages.CommandModeChange{CommandMode: true})
	s.True(model.commandMode)

	model, _ = model.Update(messages.CommandModeChange{CommandMode: false})
	s.False(model.commandMode)
}

func (s *Suite) TestUpdateRoomJoin() {
	model := New()
	s.False(model.isDealer)

	model, _ = model.Update(messages.RoomJoin{IsDealer: true})
	s.True(model.isDealer)

	model, _ = model.Update(messages.RoomJoin{IsDealer: false})
	s.False(model.isDealer)
}

func (s *Suite) TestUpdateGameState() {
	model := New()

	state := protocol2.State{
		Issues: protocol2.IssuesList{
			&protocol2.Issue{
				ID:         "1",
				TitleOrURL: "issue-1",
			},
			&protocol2.Issue{
				ID:         "2",
				TitleOrURL: "issue-2",
			},
			&protocol2.Issue{
				ID:         "3",
				TitleOrURL: "issue-3",
			},
		},
		ActiveIssue: "2",
	}
	model, _ = model.Update(messages.GameStateMessage{State: &state})
	s.Equal(state.Issues, model.issues)
	s.Equal(state.ActiveIssue, model.activeIssue)
	s.Equal(0, model.cursor.Min())
	s.Equal(2, model.cursor.Max())

	model, _ = model.Update(messages.GameStateMessage{State: nil})
	s.Nil(model.issues)
	s.Empty(model.activeIssue)
	s.Equal(0, model.cursor.Min())
	s.Equal(0, model.cursor.Max())
}

func (s *Suite) TestEmptyView() {
	model := New()
	view := model.View()
	lines := strings.Split(view, "\n")
	s.Require().Len(lines, 3)

	s.Require().Equal("Issues:", lines[0])
	s.Require().Equal("- No issues dealt yet", lines[1])
	s.Require().Empty(lines[2])
}

func (s *Suite) TestView() {
	result8 := protocol2.VoteValue("8")
	result13 := protocol2.VoteValue("13")

	model := New()
	model.issues = protocol2.IssuesList{
		&protocol2.Issue{
			ID:         "1",
			TitleOrURL: "issue-1",
		},
		&protocol2.Issue{
			ID:         "2",
			TitleOrURL: "issue-2",
		},
		&protocol2.Issue{
			ID:         "3",
			TitleOrURL: "issue-3",
			Result:     &result13,
		},
		&protocol2.Issue{
			ID:         "4",
			TitleOrURL: "issue-4",
			Result:     &result8,
		},
		&protocol2.Issue{
			ID:         "5",
			TitleOrURL: "issue-5",
		},
	}
	model.isDealer = true
	model.commandMode = false
	model.activeIssue = "2"
	model.cursor.SetRange(0, 4)
	model.cursor.SetPosition(4)
	model.Focus()

	view := model.View()
	lines := strings.Split(view, "\n")
	s.Require().Len(lines, 7)

	s.Require().Equal("Issues:", lines[0])
	s.Require().Equal("   -   issue-1", lines[1])
	s.Require().Equal("   ⠋   issue-2", lines[2])
	s.Require().Equal("  13   issue-3", lines[3])
	s.Require().Equal("   8   issue-4", lines[4])
	s.Require().Equal(">  -   issue-5", lines[5])
	s.Require().Empty(lines[6])
}