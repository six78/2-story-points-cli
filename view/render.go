package view

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"golang.org/x/exp/maps"
	"sort"
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
		return m.spinner.View() + " Connecting to Waku peers ..."
	case UserAction:
		return m.renderGame()
	}

	return "unknown app state"
}

func (m model) renderPlayerNameInput() string {
	return fmt.Sprintf(" \n\n%s\n%s", m.input.View(), m.lastCommandError)
}

func (m model) renderGame() string {
	if m.roomID == "" {
		return fmt.Sprintf(
			" Join a room or create a new one ...\n\n%s%s",
			m.input.View(),
			m.lastCommandError,
		)
	}

	if m.gameState == nil {
		return m.spinner.View() + " Waiting for initial game state ..."
	}

	return fmt.Sprintf(`
ROOM ID:      %s
DECK:         %s
ISSUE:        %s

VOTE LIST:
%s

%s

%s
`,
		m.roomID,
		renderDeck(&m.gameState.Deck),
		renderIssue(m.gameState.VoteList[m.gameState.CurrentVoteItemID]),
		renderVoteList(m.gameState.VoteList),
		m.renderPlayers(),
		m.renderActionInput(),
	)
}

type PlayerVoteResult struct {
	Player protocol.Player
	Vote   string
	Style  lipgloss.Style
}

func (m model) renderActionInput() string {
	return fmt.Sprintf("%s\n%s",
		m.input.View(),
		m.renderActionError(),
	)
}

func (m model) renderActionError() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#d78700"))
	return style.Render(m.lastCommandError)
}

func (m model) renderPlayers() string {
	players := make([]PlayerVoteResult, 0, len(m.gameState.Players))

	for playerID, player := range m.gameState.Players {
		vote, style := renderVote(&m, playerID)
		players = append(players, PlayerVoteResult{
			Player: player,
			Vote:   vote,
			Style:  style,
		})
	}

	sort.Slice(players[:], func(i, j int) bool {
		playerI := players[i].Player
		playerJ := players[j].Player
		if playerI.Order != playerJ.Order {
			return playerI.Order < playerJ.Order
		}
		return playerI.Name < playerJ.Name
	})

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
		Border(lipgloss.NormalBorder()).
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

func voteStyle(vote protocol.VoteResult) lipgloss.Style {
	if vote >= 13 {
		return DangerVoteStyle
	}
	if vote >= 5 {
		return MediumVoteStyle
	}
	return LightVoteStyle
}

func renderVote(m *model, playerID protocol.PlayerID) (string, lipgloss.Style) {
	if m.gameState.VoteState == protocol.IdleState {
		return "", CommonVoteStyle
	}
	vote, ok := m.gameState.TempVoteResult[playerID]
	if !ok {
		if m.gameState.VoteState == protocol.RevealedState ||
			m.gameState.VoteState == protocol.FinishedState {
			return "X", NoVoteStyle
		}
		return m.spinner.View(), CommonVoteStyle
	}
	if playerID == m.app.Game.Player().ID {
		playerVote := m.app.Game.PlayerVote()
		vote = &playerVote
	}
	if vote == nil {
		return "✓", ReadyVoteStyle
	}
	return fmt.Sprintf("%d", *vote), voteStyle(*vote)
}

func renderLogPath() string {
	//path := config.LogFilePath
	path := strings.Replace(config.LogFilePath, " ", "%20", -1)
	return fmt.Sprintf("LOG FILE: file:///%s", path)
}

func renderDeck(deck *protocol.Deck) string {
	votes := make([]string, 0, len([]protocol.VoteResult(*deck)))
	for _, vote := range []protocol.VoteResult(*deck) {
		votes = append(votes, fmt.Sprintf("%d", vote))
	}
	return strings.Join(votes, ", ")
}

func renderIssue(item *protocol.VoteItem) string {
	if item == nil {
		return "nil"
	}
	return fmt.Sprintf("%s [%s]", item.Text, item.ID)
}

func renderVoteList(voteList map[protocol.VoteItemID]*protocol.VoteItem) string {
	var itemStrings []string
	items := maps.Values(voteList)

	sort.Slice(items[:], func(i, j int) bool {
		return items[i].Order < items[j].Order
	})

	for _, item := range items {
		var votes []string
		for _, vote := range item.Votes {
			voteString := "nil"
			if vote != nil {
				voteString = fmt.Sprintf("%d", *vote)
			}
			votes = append(votes, voteString)
		}
		resultString := "nil"
		if item.Result != nil {
			resultString = fmt.Sprintf("%d", *item.Result)
		}
		itemString := fmt.Sprintf("%s - result: %s, votes: [%s]",
			item.Text,
			resultString,
			strings.Join(votes, ","),
		)
		itemStrings = append(itemStrings, itemString)
	}
	return strings.Join(itemStrings, "\n")
}
