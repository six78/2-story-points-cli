package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"math"
	"net/url"
	"strings"
	"time"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/commands"
	"waku-poker-planning/view/components/deckview"
	"waku-poker-planning/view/components/errorview"
	"waku-poker-planning/view/components/issuesview"
	"waku-poker-planning/view/components/issueview"
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
	roomID           protocol.RoomID
	connectionStatus waku.ConnectionStatus

	// UI components state
	commandMode    bool
	roomViewState  states.RoomView
	issueCursor    int
	errorView      errorview.Model
	playersView    playersview.Model
	shortcutsView  shortcutsview.Model
	wakuStatusView wakustatusview.Model
	deckView       deckview.Model
	issueView      issueview.Model
	issuesListView issuesview.Model

	// Workaround: Used to allow pasting multiline text (list of issues)
	disableEnterKey     bool
	disableEnterRestart chan struct{}

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
		roomID:    protocol.RoomID{},
		// UI components state
		commandMode:   false,
		roomViewState: states.ActiveIssueView,
		issueCursor:   0,
		// View components
		input:          userinput.New(false),
		spinner:        createSpinner(),
		errorView:      errorview.New(),
		playersView:    playersview.New(),
		shortcutsView:  shortcutsview.New(),
		wakuStatusView: wakustatusview.New(),
		deckView:       deckview.New(),
		issueView:      issueview.New(),
		issuesListView: issuesview.New(),
		// Other
		disableEnterKey:     false,
		disableEnterRestart: nil,
	}
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return s
}

func (m model) Init() tea.Cmd {
	m.deckView.Focus()
	return tea.Batch(
		m.input.Init(),
		m.spinner.Tick,
		m.errorView.Init(),
		m.playersView.Init(),
		m.shortcutsView.Init(),
		m.wakuStatusView.Init(),
		m.deckView.Init(),
		m.issueView.Init(),
		m.issuesListView.Init(),
		commands.InitializeApp(m.app),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		inputCommand          tea.Cmd
		spinnerCommand        tea.Cmd
		playersCommand        tea.Cmd
		issueViewCommand      tea.Cmd
		issuesListViewCommand tea.Cmd
	)

	// TODO: Rendering could be cached inside components in most cases.
	//		 This will require to only call Update when needed (because the object is copied).
	//		 On Update() we would set `cache.valid = false`
	//		 and then in View() the string could be cached and set `cache.valid = true`.
	//		 Look shortcutsview for example.

	cmds := make([]tea.Cmd, 0, 8)

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
			// Subscribe to states when app initialized
			appendCommand(commands.WaitForConnectionStatus(m.app))
			appendCommand(commands.WaitForGameState(m.app))

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
		if m.gameState != nil && msg.State != nil && msg.State.ActiveIssue != m.gameState.ActiveIssue {
			appendMessage(messages.MyVote{Result: m.app.Game.MyVote()})
		}
		m.gameState = msg.State
		appendCommand(commands.WaitForGameState(m.app))

	case messages.CommandModeChange:
		m.commandMode = msg.CommandMode

	case messages.RoomJoin:
		m.roomID = msg.RoomID
		config.Logger.Debug("room joined",
			zap.String("roomID", msg.RoomID.String()),
			zap.Bool("isDealer", msg.IsDealer))
		appendMessage(messages.MyVote{Result: m.app.Game.MyVote()})

	case messages.EnableEnterKey:
		m.disableEnterKey = false
		m.disableEnterRestart = nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			appendCommand(commands.QuitApp(m.app))
		case tea.KeyEnter:
			var cmd tea.Cmd
			if m.disableEnterKey {
				break
			}
			if m.input.Focused() {
				cmd = ProcessUserInput(&m)
				appendCommand(cmd)
				break
			}
			switch m.roomViewState {
			case states.ActiveIssueView:
				if m.gameState.VoteState() == protocol.VotingState {
					cmd = VoteOnCursor(&m)
				} else if m.gameState.VoteState() == protocol.RevealedState {
					cmd = FinishOnCursor(&m)
				}
			case states.IssuesListView:
				cmd = commands.SelectIssue(m.app, m.issueCursor)
				toggleRoomView(&m)
			}
			appendCommand(cmd)
		case tea.KeyUp:
			if !m.commandMode && m.roomViewState == states.IssuesListView {
				MoveIssueCursorUp(&m)
			}
		case tea.KeyDown:
			if !m.commandMode && m.roomViewState == states.IssuesListView {
				MoveIssueCursorDown(&m)
			}
		case tea.KeyTab:
			if !m.roomID.Empty() {
				toggleRoomView(&m)
			}
		case tea.KeyShiftTab:
			appendMessage(messages.CommandModeChange{
				CommandMode: !m.commandMode,
			})
		}

		if m.input.Focused() {
			break
		}

		if !m.roomID.Empty() {
			switch {
			case key.Matches(msg, commands.DefaultKeyMap.ExitRoom):
				appendCommand(runExitAction(&m, nil))
			case key.Matches(msg, commands.DefaultKeyMap.RevealVotes):
				appendCommand(runRevealAction(&m, nil))
			case key.Matches(msg, commands.DefaultKeyMap.FinishVote):
				appendCommand(runFinishAction(&m, nil))
			case key.Matches(msg, commands.DefaultKeyMap.RevokeVote):
				appendCommand(commands.PublishVote(m.app, ""))
			}
		} else {
			switch {
			case key.Matches(msg, commands.DefaultKeyMap.NewRoom):
				appendCommand(runNewAction(&m, nil))
			}
		}

		message, command := m.handlePastedText(msg.String())
		appendMessage(message)
		appendCommand(command)
	}

	m.input, inputCommand = m.input.Update(msg)
	m.spinner, spinnerCommand = m.spinner.Update(msg)
	m.errorView = m.errorView.Update(msg)
	m.playersView, playersCommand = m.playersView.Update(msg)
	m.shortcutsView = m.shortcutsView.Update(msg, m.roomViewState)
	m.wakuStatusView = m.wakuStatusView.Update(msg)
	m.deckView = m.deckView.Update(msg)
	m.issueView, issueViewCommand = m.issueView.Update(msg)
	m.issuesListView, issuesListViewCommand = m.issuesListView.Update(msg)

	appendCommand(inputCommand)
	appendCommand(spinnerCommand)
	appendCommand(playersCommand)
	appendCommand(issueViewCommand)
	appendCommand(issuesListViewCommand)

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
	return cursorCommand(m, m.deckView.VoteCursor(), commands.PublishVote)
}

func FinishOnCursor(m *model) tea.Cmd {
	return cursorCommand(m, m.deckView.FinishCursor(), commands.FinishVoting)
}

func cursorCommand(m *model, cursor int, command func(*app.App, protocol.VoteValue) tea.Cmd) tea.Cmd {
	if m.gameState == nil {
		return nil
	}
	if cursor < 0 || cursor > len(m.gameState.Deck) {
		return nil
	}
	cursorValue := m.gameState.Deck[cursor]
	return command(m.app, cursorValue)
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

func toggleRoomView(m *model) {
	switch m.roomViewState {
	case states.ActiveIssueView:
		m.roomViewState = states.IssuesListView
		m.deckView.Blur()
	case states.IssuesListView:
		m.roomViewState = states.ActiveIssueView
		m.deckView.Focus()
	}
}

func (m *model) handlePastedText(text string) (tea.Msg, tea.Cmd) {
	if len(text) < 16 {
		return nil, nil
	}

	config.Logger.Debug("handlePastedText", zap.String("text", text))

	// Try to parse as room id
	room, err := protocol.ParseRoomID(text)
	if err == nil {
		if !room.VersionSupported() {
			err = errors.Wrap(err, fmt.Sprintf("this room has unsupported version %d", room.Version))
			return messages.NewErrorMessage(err), nil
		}
		roomID := protocol.NewRoomID(text)
		return nil, commands.JoinRoom(m.app, roomID, nil)
	}

	// Try to parse as issues list
	lines := strings.Split(text, "\n")
	cmds := make([]tea.Cmd, 0, len(lines))

	for _, line := range lines {
		u, err := url.Parse(line)
		if err != nil {
			config.Logger.Warn("failed to parse issue url", zap.Error(err))
			continue
		}
		cmds = append(cmds, commands.AddIssue(m.app, u.String()))
	}

	m.disableEnterKey = true

	if m.disableEnterRestart != nil {
		m.disableEnterRestart <- struct{}{}
	} else {
		m.disableEnterRestart = make(chan struct{})
		cmd := commands.DelayMessage(
			100*time.Millisecond,
			messages.EnableEnterKey{},
			m.disableEnterRestart,
		)
		cmds = append(cmds, cmd)
	}

	return nil, tea.Batch(cmds...)
}

// Ensure that model fulfils the tea.Model interface at compile time.
// ref: https://www.inngest.com/blog/interactive-clis-with-bubbletea
var _ tea.Model = (*model)(nil)
