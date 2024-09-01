package hintview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/six78/2-story-points-cli/internal/view/components/voteview"
	"github.com/six78/2-story-points-cli/internal/view/messages"
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

var (
	headerStyle       = lipgloss.NewStyle() // .Foreground(lipgloss.Color("#FAFAFA"))
	acceptableStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	unacceptableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	textStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
)

type Model struct {
	hint *protocol.Hint
	deck protocol.Deck
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

		m.deck = msg.State.Deck
	}

	return m, nil
}

func (m Model) View() string {
	if m.hint == nil {
		return ""
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		headerStyle.Render("Recommended:")+""+voteview.Render(m.hint.Value, m.deck),
		headerStyle.Render("Acceptable:")+"  "+renderAcceptanceIcon(m.hint.Acceptable),
		headerStyle.Render(">")+" "+textStyle.Render(m.hint.Description),
	)
}

func renderAcceptanceIcon(acceptable bool) string {
	if acceptable {
		return acceptableStyle.Render("✓")
	}
	return unacceptableStyle.Render("x")
}
