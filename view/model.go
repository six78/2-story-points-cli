package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
	"math"
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

	// UI components state
	interactiveMode bool
	deckCursor      int
	currentView     ViewType

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
		// UI components state
		interactiveMode: !a.IsDealer(),
		deckCursor:      0,
		// View components
		input:   createTextInput(),
		spinner: createSpinner(),
	}
}

func createTextInput() textinput.Model {
	input := textinput.New()
	input.Placeholder = "Type a command..."
	input.Prompt = "┃ "
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
	var (
		inputCommand   tea.Cmd
		spinnerCommand tea.Cmd
	)

	previousState := m.state

	if !m.interactiveMode || m.roomID == "" || m.state == InputPlayerName {
		m.input.Focus()
	} else {
		m.input.Blur()
	}

	m.interactiveMode = !m.app.IsDealer()
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

			if msg.err != nil {
				m.lastCommandError = msg.err.Error()
			} else {
				m.lastCommandError = ""
			}

			config.Logger.Debug("room created or joined",
				zap.String("roomID", m.roomID),
				zap.Error(msg.err),
			)

			if m.roomID != "" {
				commands = append(commands, waitForGameState(m.app))
			}
		}

	case GameStateMessage:
		m.gameState = msg.state
		commands = append(commands, waitForGameState(m.app))

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			var cmd tea.Cmd
			if m.input.Focused() {
				cmd = processUserInput(&m)
			} else {
				cmd = VoteOnCursor(&m)
			}
			if cmd != nil {
				commands = append(commands, cmd)
			}
		case tea.KeyLeft:
			if m.interactiveMode {
				MoveCursorLeft(&m)
			}
		case tea.KeyRight:
			if m.interactiveMode {
				MoveCursorRight(&m)
			}
		case tea.KeyTab:
			if m.interactiveMode {
				toggleCurrentView(&m)
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

	return m, tea.Batch(commands...)
}

func (m model) View() string {
	if m.fatalError != nil {
		return fmt.Sprintf(" ☠️ fatal error: %s\n%s", m.fatalError, renderLogPath())
	}

	view := "\n"
	if config.Debug() {
		view += fmt.Sprintf("%s\n", renderLogPath())
	}
	view += m.renderAppState()
	return lipgloss.JoinHorizontal(lipgloss.Left, "  ", view)
}

func VoteOnCursor(m *model) tea.Cmd {
	// TODO: Instead of imitating action, return a ready-to-go tea.Cmd
	return processAction(m, fmt.Sprintf("vote %s", m.gameState.Deck[m.deckCursor]))
}

func MoveCursorLeft(m *model) {
	m.deckCursor = int(math.Max(float64(m.deckCursor-1), 0))
}

func MoveCursorRight(m *model) {
	m.deckCursor = int(math.Min(float64(m.deckCursor+1), float64(len(m.gameState.Deck)-1)))
}

func toggleCurrentView(m *model) {
	switch m.currentView {
	case RoomView:
		m.currentView = IssuesListView
	case IssuesListView:
		m.currentView = RoomView
	}
}

// Ensure that model fulfils the tea.Model interface at compile time.
// ref: https://www.inngest.com/blog/interactive-clis-with-bubbletea
var _ tea.Model = (*model)(nil)
