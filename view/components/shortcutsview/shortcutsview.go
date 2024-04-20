package shortcutsview

import (
	"fmt"
	bubblekey "github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"waku-poker-planning/view/commands"
	"waku-poker-planning/view/messages"
	"waku-poker-planning/view/states"
)

const (
	smallSeparator = " "
	bigSeparator   = "  "
)

var (
	keyStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	textStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
)

type Model struct {
	roomView    states.RoomView
	commandMode bool
	isDealer    bool
	inRoom      bool
}

func New() Model {
	return Model{
		roomView:    states.ActiveIssueView,
		commandMode: false,
		isDealer:    false,
		inRoom:      false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg, view states.RoomView) Model {
	m.roomView = view

	switch msg := msg.(type) {
	case messages.CommandModeChange:
		m.commandMode = msg.CommandMode
	case messages.AppStateFinishedMessage:
		switch msg.State {
		case states.CreatingRoom:
			m.isDealer = true
			m.inRoom = true
		case states.JoiningRoom:
			m.isDealer = false
			m.inRoom = true
		default:
		}
	case messages.RoomChange:
		m.inRoom = msg.RoomID != ""
		m.isDealer = msg.IsDealer
	}

	return m
}

func (m Model) View() string {
	keys := commands.DefaultKeyMap

	var rows []string

	if !m.inRoom {
		row := key(keys.NewRoom) + text(" "+keys.NewRoom.Help().Desc) + bigSeparator +
			key(keys.JoinRoom) + text(" "+keys.JoinRoom.Help().Desc)
		rows = append(rows, row)
	}

	if m.inRoom {
		switch m.roomView { // Row 1
		case states.ActiveIssueView:
			row := text("Use ") + key(keys.PreviousCard) +
				text(" and ") + key(keys.NextCard) +
				text(" arrows to select card and press ") + key(keys.SelectCard) + text(" to vote")
			rows = append(rows, row)

		case states.IssuesListView:
			if !m.isDealer {
				// Selecting issue is only available for dealer
				rows = append(rows, "") // Keep empty line for alignment between views
			} else {
				row := text("Use ") + key(keys.PreviousIssue) +
					text(" and ") + key(keys.NextIssue) +
					text(" arrows to select issue and press ") + key(keys.SelectCard) + text(" to deal")
				rows = append(rows, row)
			}
		}
	}

	if m.inRoom && m.isDealer { // Row 2 (optional, dealer-only)
		row := key(keys.RevealVotes) + text(" Reveal") + bigSeparator +
			keyHelp(keys.FinishVote) + bigSeparator +
			keyHelp(keys.AddIssue) + bigSeparator +
			keyHelp(keys.ExitRoom)
		rows = append(rows, row)
	}

	{ // Row 3
		var row string

		if m.inRoom {
			row = key(keys.ToggleView)
			switch m.roomView {
			case states.ActiveIssueView:
				row += text(" Switch to issues list view")
			case states.IssuesListView:
				row += text(" Switch to room view")
			default:
				row += text(" Toggle room view")
			}
			row += bigSeparator
		}

		row += key(keys.ToggleInput)
		if m.commandMode {
			row += text(" Switch to shortcuts mode")
		} else {
			row += text(" Switch to commands mode")
		}
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Top, rows...)
}

func key(key bubblekey.Binding) string {
	s := fmt.Sprintf("[%s]", key.Help().Key)
	return keyStyle.Render(s)
}

func text(text string) string {
	return textStyle.Render(text)
}

func help(key bubblekey.Binding) string {
	return text(key.Help().Desc)
}

func keyHelp(k bubblekey.Binding) string {
	return key(k) + smallSeparator + help(k)
}
