package playervoteview

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/messages"
)

var (
	CommonVoteStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			Align(lipgloss.Center)
	NoVoteStyle     = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#444444"))
	ReadyVoteStyle  = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#5fd700"))
	LightVoteStyle  = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#00d7ff"))
	MediumVoteStyle = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#ffd787"))
	DangerVoteStyle = CommonVoteStyle.Copy().Foreground(lipgloss.Color("#ff005f"))
)

type voteValueState int

const (
	voteValueDefault voteValueState = iota
	voteValueEmpty
	voteValueInProgress
	voteValueX
	voteValueHidden
)

type Model struct {
	vote  protocol.VoteValue
	style lipgloss.Style

	voteValueState voteValueState

	playerID protocol.PlayerID

	spinner spinner.Model
}

func New(playerID protocol.PlayerID) Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot

	return Model{
		playerID: playerID,
		spinner:  s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var spinnerCmd tea.Cmd
	m.spinner, spinnerCmd = m.spinner.Update(msg)

	switch msg := msg.(type) {
	case messages.GameStateMessage:
		state := msg.State
		m.vote = ""
		m.style = CommonVoteStyle

		if state.VoteState() == protocol.IdleState {
			m.voteValueState = voteValueEmpty
			break
		}

		activeIssue := msg.State.Issues.Get(msg.State.ActiveIssue)
		if activeIssue == nil {
			m.voteValueState = voteValueEmpty
			break
		}

		vote, ok := activeIssue.Votes[m.playerID]
		if !ok {
			if state.VoteState() == protocol.VotingState {
				m.voteValueState = voteValueInProgress
			} else {
				m.voteValueState = voteValueX
				m.style = NoVoteStyle
			}
			break
		}

		m.vote = vote.Value
		if m.vote == "" {
			m.voteValueState = voteValueHidden
			m.style = ReadyVoteStyle
			break
		}

		m.voteValueState = voteValueDefault
		m.style = VoteStyle(m.vote)
	}

	return m, spinnerCmd
}

func (m Model) View() string {
	switch m.voteValueState {
	case voteValueDefault:
		return string(m.vote)
	case voteValueEmpty:
		return ""
	case voteValueInProgress:
		return m.spinner.View()
	case voteValueX:
		return "X"
	case voteValueHidden:
		return "✓"
	}
	return ""
}

// Style returns a standard style for given vote value.
// We don't render the style in View, as this component is frequently used in a table cell.
func (m Model) Style() lipgloss.Style {
	return m.style
}

func VoteStyle(vote protocol.VoteValue) lipgloss.Style {
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
