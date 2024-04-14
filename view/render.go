package view

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"strconv"
	"strings"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
)

func (m model) renderAppState() string {
	switch m.state {
	case Idle:
		return "nothing is happening. boring life."
	case Initializing:
		return m.spinner.View() + " Starting Waku..."
	case InputPlayerName:
		return m.renderPlayerNameInput()
	case WaitingForPeers:
		return m.spinner.View() + " Connecting to Waku peers..."
	case InsideRoom:
		return m.renderGame()
	case CreatingRoom:
		return m.spinner.View() + " Creating room..."
	case JoiningRoom:
		return m.spinner.View() + " Joining room..."
	}

	return "unknown app state"
}

func (m model) renderPlayerNameInput() string {
	return fmt.Sprintf(" \n\n%s\n%s", m.input.View(), m.renderActionError())
}

func (m model) renderGame() string {
	return fmt.Sprintf("%s\n%s\n%s",
		m.renderRoom(),
		m.renderActionInput(),
		m.renderActionError())
}

func (m model) renderRoom() string {
	return fmt.Sprintf("%s\n%s",
		fmt.Sprintf("Room: %s", m.roomID),
		m.renderRoomView())
}

func (m model) renderRoomView() string {
	switch m.roomView {
	case CurrentIssueView:
		return m.renderRoomCurrentIssueView()
	case IssuesListView:
		return renderIssuesListView(&m)
	default:
		return fmt.Sprintf("unknown view: %d", m.roomView)
	}
}

func (m model) renderRoomCurrentIssueView() string {
	if m.roomID == "" {
		return " Join a room or create a new one ..."
	}

	if m.gameState == nil {
		return m.spinner.View() + " Waiting for initial game state ..."
	}

	return fmt.Sprintf(`Issue:  %s

%s
%s`,
		renderIssue(m.gameState.Issues.Get(m.gameState.ActiveIssue)),
		m.renderPlayers(),
		m.renderDeck(),
	)
}

type PlayerVoteResult struct {
	Player protocol.Player
	Vote   string
	Style  lipgloss.Style
}

func (m model) renderActionInput() string {
	if m.commandMode {
		return m.input.View()
	}
	return "TBD: key shortcuts"
}

func (m model) renderActionError() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d78700"))
	return style.Render(m.lastCommandError)
}

func (m model) renderPlayers() string {
	players := make([]PlayerVoteResult, 0, len(m.gameState.Players))

	for _, player := range m.gameState.Players {
		vote, style := renderVote(&m, player.ID)
		players = append(players, PlayerVoteResult{
			Player: player,
			Vote:   vote,
			Style:  style,
		})
	}

	var votes []string
	var playerNames []string
	playerColumn := -1

	for i, player := range players {
		votes = append(votes, player.Vote)
		playerNames = append(playerNames, player.Player.Name)
		if player.Player.ID == m.app.Game.Player().ID {
			playerColumn = i
		}
	}

	var CommonStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		//Background(lipgloss.Color("#7D56F4")).
		//PaddingTop(2).
		PaddingLeft(1).
		PaddingRight(1).
		Align(lipgloss.Center)

	var HeaderStyle = CommonStyle.Copy().Bold(true)

	rows := [][]string{
		votes,
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				if col == playerColumn {
					return HeaderStyle
				} else {
					return CommonStyle
				}
			default:
				return players[col].Style
			}
		}).
		Headers(playerNames...).
		Rows(rows...)

	return t.String()
}

var CommonVoteStyle = lipgloss.NewStyle().
	PaddingLeft(1).
	PaddingRight(1).
	Align(lipgloss.Center)

var NoVoteStyle = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#444444"))
var ReadyVoteStyle = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#5fd700"))
var LightVoteStyle = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#00d7ff"))
var MediumVoteStyle = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#ffd787"))
var DangerVoteStyle = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#ff005f"))

func voteStyle(vote protocol.VoteValue) lipgloss.Style {
	voteNumber, err := strconv.Atoi(string(vote))
	if err != nil {
		return CommonVoteStyle
	}
	if voteNumber >= 13 {
		return DangerVoteStyle
	}
	if voteNumber >= 5 {
		return MediumVoteStyle
	}
	return LightVoteStyle
}

func renderVote(m *model, playerID protocol.PlayerID) (string, lipgloss.Style) {
	if m.gameState.VoteState() == protocol.IdleState {
		return "", CommonVoteStyle
	}
	issue := m.gameState.Issues.Get(m.gameState.ActiveIssue)
	if issue == nil {
		config.Logger.Error("active issue not found")
		return "nil", CommonVoteStyle
	}
	vote, ok := issue.Votes[playerID]
	if !ok {
		if m.gameState.VoteState() == protocol.RevealedState ||
			m.gameState.VoteState() == protocol.FinishedState {
			return "X", NoVoteStyle
		}
		return m.spinner.View(), CommonVoteStyle
	}
	if vote.Value == "" {
		return "✓", ReadyVoteStyle
	}
	return string(vote.Value), voteStyle(vote.Value)
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
			return voteStyle(value)
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
