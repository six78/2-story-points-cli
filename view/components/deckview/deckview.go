package deckview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"math"
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/components/voteview"
	"waku-poker-planning/view/messages"
)

var (
	defaultBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	votedBorderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa"))
)

type Model struct {
	deck        protocol.Deck
	voteState   protocol.VoteState
	myVote      protocol.VoteValue
	commandMode bool
	cursor      int
	focused     bool
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
	case messages.CommandModeChange:
		m.commandMode = msg.CommandMode
	case messages.MyVote:
		m.myVote = msg.Result.Value

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyLeft:
			if !m.commandMode && m.focused {
				m.cursor = int(math.Max(float64(m.cursor-1), 0))
			}
		case tea.KeyRight:
			if !m.commandMode && m.focused {
				m.cursor = int(math.Min(float64(m.cursor+1), float64(len(m.deck)-1)))
			}
		default:
		}
	}
	return m
}

func (m Model) View() string {
	renderCursor := !m.commandMode && (m.voteState == protocol.VotingState)
	cards := make([]string, 0, len(m.deck)*2)

	for i, value := range m.deck {
		card := renderCard(value, renderCursor && i == m.cursor, value == m.myVote)
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

func (m *Model) Cursor() int {
	return m.cursor
}

func renderCard(value protocol.VoteValue, cursor bool, voted bool) string {
	card := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(*cardBorderStyle(voted)).
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

	if cursor {
		if voted {
			column = append(column, "")
		}
		column = append(column, "  ^")
	} else {
		column = append(column, "")
	}

	return lipgloss.JoinVertical(lipgloss.Top, column...)
}

func cardBorderStyle(voted bool) *lipgloss.Style {
	if voted {
		return &votedBorderStyle
	} else {
		return &defaultBorderStyle
	}
}
