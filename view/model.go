package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
)

type model struct {
	// Keep the app part of model.
	// This is the only way to keep app logic separate off the view. If I decide
	// to replace BubbleTea "frontend" with whatever else, I can keep the app as is.
	// Otherwise, I will have to move everything from app.App here. This would
	// create a logic/view mess.
	// As consequence, some properties are duplicated here and in the app.App. But
	// I find it a good trade-off.
	// All methods of the app should be called from BubbleTea. Therefore, there's
	// a tea.Cmd wrapper for each app method.
	app *app.App

	// Actual nextState that will be rendered in components.
	// This is filled from app during Update stage.
	state            State
	fatalError       error
	lastCommandError string
	gameState        *protocol.State
	roomID           string

	// Components to be rendered
	// This is filled from actual nextState during View stage.
	input   textinput.Model
	spinner spinner.Model
}

func initialModel(a *app.App) model {
	return model{
		app: a,
		// Initial model values
		state:     Initializing,
		gameState: nil,
		roomID:    "",
		// View components
		input:   createTextInput(),
		spinner: createSpinner(),
	}
}

func createTextInput() textinput.Model {
	input := textinput.New()
	input.Placeholder = "Type a command..."
	input.Prompt = "┃ "
	input.Focus()
	return input
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	//s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return s
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick, initializeApp(m.app))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	config.Logger.Debug("model update",
		zap.Any("msg", msg),
	)
	var (
		inputCommand   tea.Cmd
		spinnerCommand tea.Cmd
	)

	previousState := m.state

	m.input, inputCommand = m.input.Update(msg)
	m.spinner, spinnerCommand = m.spinner.Update(msg)

	commands := []tea.Cmd{inputCommand, spinnerCommand}

	switch msg := msg.(type) {

	case FatalErrorMessage:
		m.fatalError = msg.err

	case ActionErrorMessage:
		m.lastCommandError = msg.err.Error()

	case AppStateMessage:
		switch msg.finishedState {
		case Initializing:
			if m.app.Game.Player().Name == "" {
				m.state = InputPlayerName
			} else {
				m.state = WaitingForPeers
				commands = append(commands, waitForWakuPeers(m.app))
			}
		case InputPlayerName:
			m.state = WaitingForPeers
			commands = append(commands, waitForWakuPeers(m.app))
		case WaitingForPeers:
			m.state = UserAction
			if config.InitialAction() != "" {
				m.input.SetValue(config.InitialAction())
				cmd := processInput(&m)
				commands = append(commands, cmd)
			}
		case UserAction:
			break
		case CreatingRoom, JoiningRoom:
			m.state = UserAction
			m.roomID = m.app.Game.RoomID()
			m.gameState = m.app.GameState()
			commands = append(commands, waitForGameState(m.app))
		}

	case GameStateMessage:
		m.gameState = msg.state
		commands = append(commands, waitForGameState(m.app))

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			cmd := processUserInput(&m)
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}
	}

	if m.state != previousState {
		switch m.state {
		case UserAction:
			m.input.Placeholder = "Type a command..."
		case InputPlayerName:
			m.input.Placeholder = "Type your name..."
		}
	}

	config.Logger.Debug("model updated",
		zap.Any("state", m.state),
		zap.Any("previousState", previousState),
	)

	return m, tea.Batch(commands...)
}

func (m model) View() string {
	//config.Logger.Debug("model view")

	var view string
	if m.fatalError != nil {
		view = fmt.Sprintf(" ☠️ FATAL ERROR: %s", m.fatalError)
	} else {
		view = m.renderAppState()
	}

	return fmt.Sprintf("%s\n\n%s", renderLogPath(), view)
}

// Ensure that model fulfils the tea.Model interface at compile time.
// ref: https://www.inngest.com/blog/interactive-clis-with-bubbletea
var _ tea.Model = (*model)(nil)
