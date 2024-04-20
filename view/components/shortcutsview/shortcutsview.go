package shortcutsview

import (
	"fmt"
	bubblekey "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"waku-poker-planning/view/commands"
	"waku-poker-planning/view/states"
)

const separator = "  "

var (
	keyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
)

type Model struct {
	roomView    states.RoomView
	commandMode bool
	isDealer    bool
}

func New() Model {
	return Model{
		roomView:    states.ActiveIssueView,
		commandMode: false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(view states.RoomView, commandMode bool) Model {
	m.roomView = view
	m.commandMode = commandMode
	return m
}

func (m Model) View() string {
	keys := commands.DefaultKeyMap

	var row1 string

	switch m.roomView {
	case states.ActiveIssueView:
		row1 = text("Use ") + key(keys.PreviousCard) +
			text(" and ") + key(keys.NextCard) +
			text(" arrows to select card and press ") + key(keys.SelectCard) + text(" to vote")

	case states.IssuesListView:
		if !m.isDealer {
			// Selecting issue is only available for dealer
			row1 = "" // Keep empty line for alignment between views
		} else {
			row1 = text("Use ") + key(keys.PreviousIssue) +
				text(" and ") + key(keys.NextIssue) +
				text(" arrows to select issue and press ") + key(keys.SelectCard) + text(" to deal")
		}
	}

	row2 := key(keys.ToggleView)
	switch m.roomView {
	case states.ActiveIssueView:
		row2 += text(" Switch to issues list view")
	case states.IssuesListView:
		row2 += text(" To switch to room view")
	default:
		row2 += text(" Switch to ")
	}

	row2 += separator + key(keys.ToggleInput)
	if m.commandMode {
		row2 += text(" Switch to shortcuts mode")
	} else {
		row2 += text(" Switch to commands mode")
	}

	// TODO: Uncomment when implemented
	//row3 += separator + key(keys.LeaveRoom) + text(" Leave room")

	return lipgloss.JoinVertical(lipgloss.Top, row1, row2)
}

func key(key bubblekey.Binding) string {
	s := fmt.Sprintf("[%s]", key.Help().Key)
	return keyStyle.Render(s)
}

func text(text string) string {
	return textStyle.Render(text)
}
