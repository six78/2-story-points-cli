package view

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
)

type EndOfState struct{}

type model struct {
	state        protocol.State
	stateChannel chan protocol.State

	//viewport    viewport.Model
	input       textinput.Model
	senderStyle lipgloss.Style
}

func initialModel(initialState protocol.State) model {
	ta := textinput.New()
	ta.Placeholder = "Type a command..."
	ta.Prompt = "┃ "
	ta.Focus()

	return model{
		state:        initialState,
		stateChannel: nil,
		input:        ta,
		//viewport:    vp,
	}
}

func (m model) waitForState() tea.Msg {
	state, more := <-m.stateChannel
	if !more {
		return EndOfState{}
	}
	return state
}

func (m model) Init() tea.Cmd {
	config.Logger.Debug("view model init")
	return tea.Batch(textarea.Blink, m.waitForState)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	config.Logger.Debug("view update", zap.Any("msg", msg))

	var (
		tiCmd tea.Cmd
		//vpCmd tea.Cmd
	)

	m.input, tiCmd = m.input.Update(msg)
	//m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {

	case protocol.State:
		m.state = msg

	case tea.KeyMsg:
		// Ctrl+c exits. Even with short running programs it's good to have
		// a quit key, just in case your logic is off. Users will be very
		// annoyed if they can't exit.

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			//m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.input.Value())
			//m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.input.Reset()
			//m.viewport.GotoBottom()
		}

		//// Is it a key press?
		//case tea.KeyMsg:
		//
		//
		//	// Cool, what was the actual key pressed?
		//	switch msg.String() {
		//
		//	// These keys should exit the program.
		//	case "ctrl+c", "q":
		//		return m, tea.Quit
		//
		//	// The "up" and "k" keys move the cursor up
		//	case "up", "k":
		//		if m.cursor > 0 {
		//			m.cursor--
		//		}
		//
		//	// The "down" and "j" keys move the cursor down
		//	case "down", "j":
		//		if m.cursor < len(m.choices)-1 {
		//			m.cursor++
		//		}
		//
		//	// The "enter" key and the spacebar (a literal space) toggle
		//	// the selected state for the item that the cursor is pointing at.
		//	case "enter", " ":
		//		_, ok := m.selected[m.cursor]
		//		if ok {
		//			delete(m.selected, m.cursor)
		//		} else {
		//			m.selected[m.cursor] = struct{}{}
		//		}
		//	}

	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, tea.Batch(tiCmd, m.waitForState)
}

func (m model) View() string {
	config.Logger.Debug("view View")

	players, err := json.Marshal(m.state.Players)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(
		"\n VOTING FOR: %s\n\n PLAYERS: %s\n\n%s\n",
		m.state.VoteFor,
		players,
		//m.viewport.View(),
		m.input.View(),
	)
}
