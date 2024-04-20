package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
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
	"waku-poker-planning/view/components/userinput"
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
	input   userinput.Model
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
		commandMode: false,
		deckCursor:  0,
		roomView:    states.ActiveIssueView,
		issueCursor: 0,
		// View components
		input:          userinput.New(false),
		spinner:        createSpinner(),
		errorView:      errorview.New(),
		playersView:    playersview.New(),
		shortcutsView:  shortcutsview.New(),
		wakuStatusView: wakustatusview.New(),
	}
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return s
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
		m.spinner.Tick,
		m.errorView.Init(),
		m.playersView.Init(),
		m.shortcutsView.Init(),
		m.wakuStatusView.Init(),
		commands.InitializeApp(m.app),
	)
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

	switchToState := func(state states.AppState) {
		m.state = state
		appendMessage(messages.AppStateMessage{State: state})
	}

	switch msg := msg.(type) {

	case messages.FatalErrorMessage:
		m.fatalError = msg.Err

	case messages.AppStateFinishedMessage:
		switch msg.State {
		case states.Initializing:
			// Notify PlayerID generated
			appendMessage(messages.PlayerIDMessage{
				PlayerID: m.app.Game.Player().ID,
			})
			// Determine next state
			if m.app.Game.Player().Name == "" {
				switchToState(states.InputPlayerName)
			} else {
				switchToState(states.WaitingForPeers)
			}
			// Subscribe to connection status when app initialized
			appendCommand(commands.WaitForConnectionStatus(m.app))
		case states.InputPlayerName:
			switchToState(states.WaitingForPeers)

		case states.WaitingForPeers:
			switchToState(states.Playing)
			if config.InitialAction() != "" {
				m.input.SetValue(config.InitialAction())
				cmd := ProcessInput(&m)
				appendCommand(cmd)
			}
		case states.Playing:
			break
		case states.CreatingRoom, states.JoiningRoom:
			// TODO: Drop CreatingRoom and JoiningRoom and use messages.RoomChange instead
			switchToState(states.Playing)
			m.gameState = m.app.GameState()

			appendMessage(messages.RoomChange{
				RoomID:   m.app.Game.RoomID(),
				IsDealer: m.app.Game.IsDealer(),
			})

			appendMessage(messages.ErrorMessage{
				Err: msg.Err,
			})

			appendCommand(commands.WaitForGameState(m.app))

			config.Logger.Debug("room created or joined",
				zap.String("roomID", m.roomID),
				zap.Error(msg.Err),
			)
		}
	case messages.AppStateMessage:
		// Immediately skip to next state if peers already connected
		if m.state == states.WaitingForPeers && m.connectionStatus.PeersCount > 0 {
			appendMessage(messages.AppStateFinishedMessage{
				State: states.WaitingForPeers,
			})
		}

	case messages.ConnectionStatus:
		m.connectionStatus = msg.Status
		if m.state == states.WaitingForPeers && m.connectionStatus.PeersCount > 0 {
			appendMessage(messages.AppStateFinishedMessage{
				State: states.WaitingForPeers,
			})
		}
		appendCommand(commands.WaitForConnectionStatus(m.app))

	case messages.GameStateMessage:
		m.gameState = msg.State
		appendCommand(commands.WaitForGameState(m.app))

	case messages.CommandModeChange:
		m.commandMode = msg.CommandMode

	case messages.RoomChange:
		m.roomID = msg.RoomID
		config.Logger.Debug("room changed",
			zap.String("roomID", msg.RoomID),
			zap.Bool("isDealer", msg.IsDealer))

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			// TODO: publish offline message
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
			config.Logger.Debug("<<< tab")
			toggleCurrentView(&m)

		case tea.KeyShiftTab:
			config.Logger.Debug("<<< shift+tab")
			appendMessage(messages.CommandModeChange{
				CommandMode: !m.commandMode,
			})
		}
		if !m.input.Focused() {
			if m.roomID == "" && key.Matches(msg, commands.DefaultKeyMap.NewRoom) {
				appendCommand(runNewAction(&m, nil))
				//case key.Matches(msg, commands.DefaultKeyMap.JoinRoom):
				//	appendCommand(runJoinAction(&m, nil))
			}
			if m.roomID != "" && key.Matches(msg, commands.DefaultKeyMap.ExitRoom) {
				appendCommand(runExitAction(&m, nil))
			}
		}
	}

	m.input, inputCommand = m.input.Update(msg)
	m.spinner, spinnerCommand = m.spinner.Update(msg)
	m.errorView = m.errorView.Update(msg)
	m.playersView, playersCommand = m.playersView.Update(msg)
	m.shortcutsView = m.shortcutsView.Update(msg, m.roomView)
	m.wakuStatusView = m.wakuStatusView.Update(msg)

	appendCommand(inputCommand)
	appendCommand(spinnerCommand)
	appendCommand(playersCommand)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.fatalError != nil {
		return fmt.Sprintf(" ☠️ fatal error: %s\n%s", m.fatalError, renderLogPath())
	}

	view := "\n"
	if config.Debug() {
		view += fmt.Sprintf("%s\n\n", renderLogPath())
	}
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
