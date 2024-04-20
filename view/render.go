package view

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"strings"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/components/playervoteview"
	"waku-poker-planning/view/states"
)

/*
	TODO: Refactor this.
		I don't like that on each method we're copying the model, this is bad.
		Probably a good idea would be to have a separate Model (with Init/Update/View methods)
		for each component. This would be a proper way of using bubbletea!
*/

func (m model) renderAppState() string {
	switch m.state {
	case states.Idle:
		return "nothing is happening. boring life."
	case states.Initializing:
		return m.spinner.View() + " Starting Waku..."
	case states.InputPlayerName:
		return m.renderPlayerNameInput()
	case states.WaitingForPeers:
		return m.spinner.View() + " Connecting to Waku peers..."
	case states.InsideRoom:
		return m.renderGame()
	case states.CreatingRoom:
		return m.spinner.View() + " Creating room..."
	case states.JoiningRoom:
		return m.spinner.View() + " Joining room..."
	}

	return "unknown app state"
}

func (m model) renderPlayerNameInput() string {
	return fmt.Sprintf(" \n\n%s\n%s",
		m.input.View(),
		m.errorView.View())
}

func (m model) renderGame() string {
	return lipgloss.JoinVertical(lipgloss.Top,
		fmt.Sprintf("Room: %s", m.roomID),
		m.renderRoomView(),
		m.renderActionInput(),
		m.errorView.View())
}

func (m model) renderRoomView() string {
	switch m.roomView {
	case states.ActiveIssueView:
		return m.renderRoomCurrentIssueView()
	case states.IssuesListView:
		return renderIssuesListView(&m)
	default:
		return fmt.Sprintf("unknown view: %d", m.roomView)
	}
}

func (m model) renderRoomCurrentIssueView() string {
	if m.roomID == "" {
		return " Join a room or create a new one ...\n" + m.input.View()
	}

	if m.gameState == nil {
		return m.spinner.View() + " Waiting for initial game state ..."
	}

	return fmt.Sprintf(`Issue:  %s

%s
%s`,
		renderIssue(m.gameState.Issues.Get(m.gameState.ActiveIssue)),
		m.playersView.View(),
		m.renderDeck(),
	)
}

func (m model) renderActionInput() string {
	if m.commandMode {
		return m.input.View()
	}
	return m.shortcutsView.View()
}

func renderLogPath() string {
	//path := config.LogFilePath
	path := strings.Replace(config.LogFilePath, " ", "%20", -1)
	return fmt.Sprintf("Log: file:///%s", path)
}

func (m model) renderDeck() string {
	if m.gameState == nil {
		return ""
	}

	deck := m.gameState.Deck
	renderCursor := !m.commandMode
	myVote := m.app.Game.MyVote().Value

	cards := make([]string, 0, len(deck)*2)

	for i, value := range deck {
		card := renderCard(value, renderCursor && i == m.deckCursor, value == myVote)
		cards = append(cards, card, " ") // Add a space between cards
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, cards...)
}

func renderCard(value protocol.VoteValue, cursor bool, voted bool) string {
	var borderStyle lipgloss.Style
	if voted {
		borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa"))
	} else {
		borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	}

	card := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			return playervoteview.VoteStyle(value)
		}).
		Rows([]string{string(value)}).
		String()

	var column []string
	column = []string{}

	if !voted {
		column = append(column, "")
		//card = lipgloss.JoinVertical(lipgloss.Top, []string{"", card}...)
	}

	column = append(column, card)

	if cursor {
		if voted {
			column = append(column, "")
		}
		column = append(column, "  ^")
		//card = lipgloss.JoinVertical(lipgloss.Top, []string{card, "  ^"}...)
	}

	return lipgloss.JoinVertical(lipgloss.Top, column...)
}

func renderIssue(item *protocol.Issue) string {
	if item == nil {
		return "-"
	}
	return item.TitleOrURL
}

func renderIssuesListView(m *model) string {
	if m.gameState == nil {
		return ""
	}

	showSelector := !m.commandMode && m.app.IsDealer()
	issues := m.gameState.Issues
	activeIssue := m.gameState.ActiveIssue

	var items []string

	for i, issue := range issues {
		result := "-"
		if issue.Result != nil {
			result = string(*issue.Result)
		} else if issue.ID == activeIssue {
			result = m.spinner.View()
		}

		var item string
		var style lipgloss.Style
		if showSelector && i == m.issueCursor {
			item += "> "
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
		} else {
			item += "  "
		}

		item += fmt.Sprintf("%s  %s", result, issue.TitleOrURL)
		items = append(items, style.Render(item))
	}

	fullBlock := lipgloss.JoinVertical(lipgloss.Top, items...)
	return fmt.Sprintf("Issues:\n%s", fullBlock)
}
