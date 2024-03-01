package view

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
	"waku-poker-planning/config"
	"waku-poker-planning/game"
	"waku-poker-planning/protocol"
)

type EndOfState struct{}

type model struct {
	state        protocol.State
	game         *game.Game
	stateChannel chan protocol.State

	input       textinput.Model
	senderStyle lipgloss.Style
}

func initialModel(g *game.Game) model {
	ta := textinput.New()
	ta.Placeholder = "Type a command..."
	ta.Prompt = "┃ "
	ta.Focus()

	return model{
		game:         g,
		state:        g.CurrentState(),
		stateChannel: g.SubscribeToStateChanges(),
		input:        ta,
	}
}

func (m model) waitForGameState() tea.Msg {
	state, more := <-m.stateChannel
	if !more {
		return EndOfState{}
	}
	return state
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForGameState)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var inputCommand tea.Cmd
	m.input, inputCommand = m.input.Update(msg)

	commands := []tea.Cmd{inputCommand, m.waitForGameState}

	switch msg := msg.(type) {
	case protocol.State:
		m.state = msg

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			//m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.input.Value())
			//m.viewport.SetContent(strings.Join(m.messages, "\n"))
			userCommand := m.input.Value()

			if strings.HasPrefix(userCommand, "online") {
				onlineUser := strings.TrimPrefix(userCommand, "online")
				onlineUser = strings.Trim(onlineUser, " ")
				commands = append(commands, func() tea.Msg {
					m.game.PublishOnline(onlineUser)
					return nil
				})
			}

			if strings.HasPrefix(userCommand, "rename") {
				user := strings.TrimPrefix(userCommand, "rename")
				user = strings.Trim(user, " ")
				commands = append(commands, func() tea.Msg {
					config.PlayerName = user
					m.game.PublishOnline(config.PlayerName)
					return nil
				})
			}

			m.input.Reset()
		}
	}

	return m, tea.Batch(commands...)
}

func (m model) View() string {
	players, err := json.Marshal(m.state.Players)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(`  LOG FILE: %s

  VOTING FOR: %s

  PLAYERS: %s

%s

`,
		"file:///"+config.LogFilePath,
		m.state.VoteFor,
		players,
		m.input.View(),
	)
}
