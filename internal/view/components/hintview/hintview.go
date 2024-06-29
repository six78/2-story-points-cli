package hintview

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/six78/2-story-points-cli/internal/config"
	"github.com/six78/2-story-points-cli/internal/view/components/voteview"
	"github.com/six78/2-story-points-cli/internal/view/messages"
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

var (
	headerStyle       = lipgloss.NewStyle() // .Foreground(lipgloss.Color("#FAFAFA"))
	acceptableStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	unacceptableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	textStyle         = lipgloss.NewStyle() // .Foreground(lipgloss.Color("#FAFAFA"))
	MentionStyle      = textStyle.Copy().Italic(true).Foreground(config.UserColor)
)

type Model struct {
	hint *protocol.Hint
}

func New() Model {
	return Model{
		hint: nil,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case messages.GameStateMessage:
		if msg.State == nil || !msg.State.VotesRevealed {
			m.hint = nil
			break
		}

		issue := msg.State.Issues.Get(msg.State.ActiveIssue)
		if issue != nil {
			m.hint = issue.Hint
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.hint == nil {
		return ""
	}

	verdictStyle := unacceptableStyle
	verdictText := "x"
	if m.hint.Acceptable {
		verdictStyle = acceptableStyle
		verdictText = "✓"
	}

	rejectionReason := ""
	if !m.hint.Acceptable {
		rejectionReason = fmt.Sprintf(" (%s)", textStyle.Render(m.hint.RejectReason))
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		"",
		headerStyle.Render("Recommended:")+" "+voteview.Render(m.hint.Hint),
		headerStyle.Render("Acceptable:")+"   "+verdictStyle.Render(verdictText)+rejectionReason,
		headerStyle.Render("What to do:")+"   "+textStyle.Render(m.hint.Advice),
		"",
	)
}
