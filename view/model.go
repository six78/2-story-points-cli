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
	"waku-poker-planning/view/components/errorview"
	"waku-poker-planning/view/messages"
	"waku-poker-planning/view/states"
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
	state      states.AppState
	fatalError error
	gameState  *protocol.State
	roomID     string

	// UI components state
	commandMode bool
	deckCursor  int
	roomView    RoomView
	issueCursor int
	errorView   errorview.Model

	// Components to be rendered
	// This is filled from actual nextState during View stage.
	input   textinput.Model
	spinner spinner.Model
}

func initialModel(a *app.App) model {
	return model{
		app: a,
		// Initial model values
		state:     states.Initializing,
		gameState: nil,
		roomID:    "",
		// UI components state
		commandMode: a.IsDealer(),
		deckCursor:  0,
		roomView:    CurrentIssueView,
		issueCursor: 0,
		// View components
		input:     createTextInput(),
		spinner:   createSpinner(),
		errorView: errorview.New(),
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
	return tea.Batch(
		textarea.Blink,
		m.spinner.Tick,
		m.errorView.Init(),
		initializeApp(m.app),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		inputCommand   tea.Cmd
		spinnerCommand tea.Cmd
	)

	previousState := m.state

	if m.commandMode || m.roomID == "" || m.state == states.InputPlayerName {
		m.input.Focus()
	} else {
		m.input.Blur()
	}

	m.commandMode = m.app.IsDealer()
	m.input, inputCommand = m.input.Update(msg)
	m.spinner, spinnerCommand = m.spinner.Update(msg)
	m.errorView = m.errorView.Update(msg)

	commands := []tea.Cmd{inputCommand, spinnerCommand}

	switch msg := msg.(type) {

	case messages.FatalErrorMessage:
		m.fatalError = msg.Err

	case messages.AppStateMessage:
		switch msg.FinishedState {
		case states.Initializing:
			if m.app.Game.Player().Name == "" {
				m.state = states.InputPlayerName
			} else {
				m.state = states.WaitingForPeers
				commands = append(commands, waitForWakuPeers(m.app))
			}
		case states.InputPlayerName:
			m.state = states.WaitingForPeers
			commands = append(commands, waitForWakuPeers(m.app))
		case states.WaitingForPeers:
			m.state = states.InsideRoom
			if config.InitialAction() != "" {
				m.input.SetValue(config.InitialAction())
				cmd := processInput(&m)
				commands = append(commands, cmd)
			}
		case states.InsideRoom:
			break
		case states.CreatingRoom, states.JoiningRoom:
			m.state = states.InsideRoom
			m.roomID = m.app.Game.RoomID()
			m.gameState = m.app.GameState()

			//if msg.Err != nil {
			//	m.lastCommandError = msg.Err.Error()
			//} else {
			//	m.lastCommandError = ""
			//}

			config.Logger.Debug("room created or joined",
				zap.String("roomID", m.roomID),
				zap.Error(msg.Err),
			)

			if m.roomID != "" {
				commands = append(commands, waitForGameState(m.app))
			}
		}

	case messages.GameStateMessage:
		m.gameState = msg.State
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
				switch m.roomView {
				case CurrentIssueView:
					cmd = VoteOnCursor(&m)
				case IssuesListView:
					cmd = DealOnCursor(&m)
				}
			}
			if cmd != nil {
				commands = append(commands, cmd)
			}
		case tea.KeyLeft:
			if !m.commandMode && m.roomView == CurrentIssueView {
				MoveCursorLeft(&m)
			}
		case tea.KeyRight:
			if !m.commandMode && m.roomView == CurrentIssueView {
				MoveCursorRight(&m)
			}
		case tea.KeyUp:
			if !m.commandMode && m.roomView == IssuesListView {
				MoveIssueCursorUp(&m)
			}
		case tea.KeyDown:
			if !m.commandMode && m.roomView == IssuesListView {
				MoveIssueCursorDown(&m)
			}
		case tea.KeyTab:
			toggleCurrentView(&m)
		}
	}

	if m.state != previousState {
		switch m.state {
		case states.InsideRoom:
			m.input.Placeholder = "Type a command..."
		case states.InputPlayerName:
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

func DealOnCursor(m *model) tea.Cmd {
	// TODO: Instead of imitating action, return a ready-to-go tea.Cmd
	return processAction(m, fmt.Sprintf("select %d", m.issueCursor))
}

func MoveCursorLeft(m *model) {
	m.deckCursor = int(math.Max(float64(m.deckCursor-1), 0))
}

func MoveCursorRight(m *model) {
	m.deckCursor = int(math.Min(float64(m.deckCursor+1), float64(len(m.gameState.Deck)-1)))
}

func MoveIssueCursorUp(m *model) {
	m.issueCursor = int(math.Max(float64(m.issueCursor-1), 0))
}

func MoveIssueCursorDown(m *model) {
	m.issueCursor = int(math.Min(float64(m.issueCursor+1), float64(len(m.gameState.Issues)-1)))
}

func toggleCurrentView(m *model) {
	switch m.roomView {
	case CurrentIssueView:
		m.roomView = IssuesListView
	case IssuesListView:
		m.roomView = CurrentIssueView
	}
}

// Ensure that model fulfils the tea.Model interface at compile time.
// ref: https://www.inngest.com/blog/interactive-clis-with-bubbletea
var _ tea.Model = (*model)(nil)
