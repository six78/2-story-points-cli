package deckview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"math"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/components/voteview"
	"waku-poker-planning/view/messages"
)

var (
	defaultBorderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	votedBorderStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa"))
	highlightBorderStyle = lipgloss.NewStyle().Foreground(config.UserColor)
)

type Model struct {
	deck         protocol.Deck
	voteState    protocol.VoteState
	myVote       protocol.VoteValue
	focused      bool
	isDealer     bool
	commandMode  bool
	voteCursor   int
	finishCursor int
}

func New(focused bool) Model {
	return Model{
		focused: focused,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) Model {
	switch msg := msg.(type) {
	case messages.GameStateMessage:
		if msg.State == nil {
			m.deck = protocol.Deck{}
			m.voteState = protocol.IdleState
		} else {
			m.deck = msg.State.Deck
			m.voteState = msg.State.VoteState()
		}

	case messages.RoomJoin:
		m.isDealer = msg.IsDealer

	case messages.CommandModeChange:
		m.commandMode = msg.CommandMode

	case messages.MyVote:
		m.myVote = msg.Result.Value

	case tea.KeyMsg:
		if m.commandMode || !m.focused {
			break
		}
		switch msg.Type {
		case tea.KeyLeft:
			switch m.voteState {
			case protocol.VotingState:
				m.voteCursor = m.decrementCursor(m.voteCursor)
			case protocol.RevealedState:
				if m.isDealer {
					m.finishCursor = m.decrementCursor(m.finishCursor)
				}
			default:
			}

		case tea.KeyRight:
			switch m.voteState {
			case protocol.VotingState:
				m.voteCursor = m.incrementCursor(m.voteCursor)
			case protocol.RevealedState:
				if m.isDealer {
					m.finishCursor = m.incrementCursor(m.finishCursor)
				}
			default:
			}
		default:
		}
	}
	return m
}

func (m Model) View() string {
	renderVoteCursor := !m.commandMode && (m.voteState == protocol.VotingState)
	renderFinishCursor := !m.commandMode && (m.voteState == protocol.RevealedState) && m.isDealer
	cards := make([]string, 0, len(m.deck)*2)

	for i, value := range m.deck {
		card := renderCard(value,
			renderVoteCursor && i == m.voteCursor,
			renderFinishCursor && i == m.finishCursor,
			value == m.myVote,
		)
		cards = append(cards, card, " ") // Add a space between cards
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, cards...)
}

func (m *Model) Focus() {
	m.focused = true
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) VoteCursor() int {
	return m.voteCursor
}

func (m *Model) FinishCursor() int {
	return m.finishCursor
}

func renderCard(value protocol.VoteValue, voteCursor bool, finishCursor bool, voted bool) string {
	card := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(*cardBorderStyle(voted, finishCursor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			return *voteview.VoteStyle(value)
		}).
		Rows([]string{string(value)}).
		String()

	var column []string
	column = []string{}

	if !voted {
		column = append(column, "")
	}

	column = append(column, card)

	if voteCursor {
		if voted {
			column = append(column, "")
		}
		column = append(column, "  ^")
	} else {
		column = append(column, "")
	}

	return lipgloss.JoinVertical(lipgloss.Top, column...)
}

func cardBorderStyle(voted bool, highlight bool) *lipgloss.Style {
	if highlight {
		return &highlightBorderStyle
	}
	if voted {
		return &votedBorderStyle
	}
	return &defaultBorderStyle
}

func (m *Model) decrementCursor(cursor int) int {
	return int(math.Max(float64(cursor-1), 0))
}

func (m *Model) incrementCursor(cursor int) int {
	return int(math.Min(float64(cursor+1), float64(len(m.deck)-1)))
}
