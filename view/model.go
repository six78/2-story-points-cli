package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
	"math"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/commands"
	"waku-poker-planning/view/components/errorview"
	"waku-poker-planning/view/components/playersview"
	"waku-poker-planning/view/components/shortcutsview"
	"waku-poker-planning/view/components/wakustatusview"
	"waku-poker-planning/view/messages"
	"waku-poker-planning/view/states"
	"waku-poker-planning/waku"
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
	state            states.AppState
	fatalError       error
	gameState        *protocol.State
	roomID           string
	connectionStatus waku.ConnectionStatus

	// UI components state
	commandMode    bool
	deckCursor     int
	roomView       states.RoomView
	issueCursor    int
	errorView      errorview.Model
	playersView    playersview.Model
	shortcutsView  shortcutsview.Model
	wakuStatusView wakustatusview.Model

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
		roomView:    states.ActiveIssueView,
		issueCursor: 0,
		// View components
		input:          createTextInput(a.IsDealer()),
		spinner:        createSpinner(),
		errorView:      errorview.New(),
		playersView:    playersview.New(),
		shortcutsView:  shortcutsview.New(),
		wakuStatusView: wakustatusview.New(),
	}
}

func createTextInput(focus bool) textinput.Model {
	input := textinput.New()
	input.Placeholder = "Type a command..."
	input.Prompt = "┃ "
	input.Cursor.SetMode(cursor.CursorBlink)
	input.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
	if focus {
		input.Focus()
	}
	return input
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return s
}

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		m.errorView.Init(),
		m.playersView.Init(),
		m.shortcutsView.Init(),
		m.wakuStatusView.Init(),
		commands.InitializeApp(m.app),
	}
	//if m.app.IsDealer() {
	cmds = append(cmds, textinput.Blink)
	//}
	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		inputCommand   tea.Cmd
		spinnerCommand tea.Cmd
		playersCommand tea.Cmd
	)

	// TODO: Rendering could be cached inside components in most cases.
	//		 This will require to only call Update when needed (because the object is copied).
	//		 On Update() we would set `cache.valid = false`
	//		 and then in View() the string could be cached and set `cache.valid = true`.
	//		 Look shortcutsview for example.

	var cmds []tea.Cmd

	appendCommand := func(command tea.Cmd) {
		cmds = append(cmds, command)
	}

	appendMessage := func(injectedMessage tea.Msg) {
		cmds = append(cmds, func() tea.Msg {
			return injectedMessage
		})
	}

	m.commandMode = m.app.IsDealer()
	previousState := m.state

	switch msg := msg.(type) {

	case messages.FatalErrorMessage:
		m.fatalError = msg.Err

	case messages.AppStateMessage:
		switch msg.FinishedState {
		case states.Initializing:
			// Notify PlayerID generated
			appendMessage(messages.PlayerIDMessage{
				PlayerID: m.app.Game.Player().ID,
			})
			// Determine next state
			if m.app.Game.Player().Name == "" {
				m.state = states.InputPlayerName
				cmd := m.input.Focus()
				appendCommand(cmd)
			} else {
				m.state = states.WaitingForPeers
				//appendCommand(commands.WaitForWakuPeers(m.app))
				appendCommand(commands.WaitForConnectionStatus(m.app))
			}
		case states.InputPlayerName:
			m.state = states.WaitingForPeers
			//appendCommand(commands.WaitForWakuPeers(m.app))
			appendCommand(commands.WaitForConnectionStatus(m.app))

			if m.commandMode {
				cmd := m.input.Focus()
				appendCommand(cmd)
			} else {
				m.input.Blur()
			}

		case states.WaitingForPeers:
			m.state = states.InsideRoom
			if config.InitialAction() != "" {
				m.input.SetValue(config.InitialAction())
				cmd := ProcessInput(&m)
				appendCommand(cmd)
			}
		case states.InsideRoom:
			break
		case states.CreatingRoom, states.JoiningRoom:
			m.state = states.InsideRoom
			m.roomID = m.app.Game.RoomID()
			m.gameState = m.app.GameState()

			appendMessage(messages.ErrorMessage{
				Err: msg.Err,
			})

			if m.roomID != "" {
				appendCommand(commands.WaitForGameState(m.app))
			}

			config.Logger.Debug("room created or joined",
				zap.String("roomID", m.roomID),
				zap.Error(msg.Err),
			)
		}
	case messages.ConnectionStatus:
		m.connectionStatus = msg.Status
		if m.state == states.WaitingForPeers && m.connectionStatus.PeersCount > 0 {
			appendMessage(messages.AppStateMessage{
				FinishedState: states.WaitingForPeers,
			})
		}
		appendCommand(commands.WaitForConnectionStatus(m.app))

	case messages.GameStateMessage:
		m.gameState = msg.State
		appendCommand(commands.WaitForGameState(m.app))

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			var cmd tea.Cmd
			if m.input.Focused() {
				cmd = ProcessUserInput(&m)
			} else {
				switch m.roomView {
				case states.ActiveIssueView:
					cmd = VoteOnCursor(&m)
				case states.IssuesListView:
					cmd = DealOnCursor(&m)
				}
			}
			if cmd != nil {
				appendCommand(cmd)
			}
		case tea.KeyLeft:
			if !m.commandMode && m.roomView == states.ActiveIssueView {
				MoveCursorLeft(&m)
			}
		case tea.KeyRight:
			if !m.commandMode && m.roomView == states.ActiveIssueView {
				MoveCursorRight(&m)
			}
		case tea.KeyUp:
			if !m.commandMode && m.roomView == states.IssuesListView {
				MoveIssueCursorUp(&m)
			}
		case tea.KeyDown:
			if !m.commandMode && m.roomView == states.IssuesListView {
				MoveIssueCursorDown(&m)
			}
		case tea.KeyTab:
			//appendCommand(cmds.ToggleRoomView(m.roomView))
			toggleCurrentView(&m)
		}
	}

	if m.state == states.WaitingForPeers && m.connectionStatus.PeersCount > 0 {
		appendMessage(messages.AppStateMessage{
			FinishedState: states.WaitingForPeers,
		})
	}

	m.input, inputCommand = m.input.Update(msg)
	m.spinner, spinnerCommand = m.spinner.Update(msg)
	m.errorView = m.errorView.Update(msg)
	m.playersView, playersCommand = m.playersView.Update(msg)
	m.shortcutsView = m.shortcutsView.Update(m.roomView, m.commandMode)
	m.wakuStatusView = m.wakuStatusView.Update(msg)

	if !m.input.Focused() {
		appendCommand(m.input.Focus())
	}

	appendCommand(inputCommand)
	appendCommand(spinnerCommand)
	appendCommand(playersCommand)

	if m.state != previousState {
		switch m.state {
		case states.InsideRoom:
			m.input.Placeholder = "Type a command..."
		case states.InputPlayerName:
			m.input.Placeholder = "Type your name..."
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.fatalError != nil {
		return fmt.Sprintf(" ☠️ fatal error: %s\n%s", m.fatalError, renderLogPath())
	}

	view := "\n"
	if config.Debug() {
		view += fmt.Sprintf("%s\n", renderLogPath())
	}
	view += m.wakuStatusView.View() + "\n\n"
	view += m.renderAppState()

	return lipgloss.JoinHorizontal(lipgloss.Left, "  ", view)
}

func VoteOnCursor(m *model) tea.Cmd {
	if m.gameState == nil {
		return nil
	}
	// TODO: Instead of imitating action, return a ready-to-go tea.Cmd
	return ProcessAction(m, fmt.Sprintf("vote %s", m.gameState.Deck[m.deckCursor]))
}

func DealOnCursor(m *model) tea.Cmd {
	// TODO: Instead of imitating action, return a ready-to-go tea.Cmd
	return ProcessAction(m, fmt.Sprintf("select %d", m.issueCursor))
}

func MoveCursorLeft(m *model) {
	m.deckCursor = int(math.Max(float64(m.deckCursor-1), 0))
}

func MoveCursorRight(m *model) {
	if m.gameState == nil {
		return
	}
	m.deckCursor = int(math.Min(float64(m.deckCursor+1), float64(len(m.gameState.Deck)-1)))
}

func MoveIssueCursorUp(m *model) {
	m.issueCursor = int(math.Max(float64(m.issueCursor-1), 0))
}

func MoveIssueCursorDown(m *model) {
	if m.gameState == nil {
		return
	}
	m.issueCursor = int(math.Min(float64(m.issueCursor+1), float64(len(m.gameState.Issues)-1)))
}

func toggleCurrentView(m *model) {
	switch m.roomView {
	case states.ActiveIssueView:
		m.roomView = states.IssuesListView
	case states.IssuesListView:
		m.roomView = states.ActiveIssueView
	}
}

// Ensure that model fulfils the tea.Model interface at compile time.
// ref: https://www.inngest.com/blog/interactive-clis-with-bubbletea
var _ tea.Model = (*model)(nil)
