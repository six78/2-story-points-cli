package view

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"go.uber.org/zap"
	"math"
	"sort"
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
	appState         app.State
	fatalError       error
	lastCommandError string
	gameState        *protocol.State
	gameSessionID    string

	// Components to be rendered
	// This is filled from actual nextState during View stage.
	input       textinput.Model
	spinner     spinner.Model
	senderStyle lipgloss.Style
}

func initialModel(a *app.App) model {
	return model{
		app: a,
		// Initial model values
		appState:      app.Initializing,
		gameState:     nil,
		gameSessionID: "",
		// View components
		input:   createTextInput(),
		spinner: createSpinner(),
	}
}

func createTextInput() textinput.Model {
	ta := textinput.New()
	ta.Placeholder = "Type a command..."
	ta.Prompt = "┃ "
	ta.Focus()
	return ta
}

func createSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
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
	m.input, inputCommand = m.input.Update(msg)
	m.spinner, spinnerCommand = m.spinner.Update(msg)

	commands := []tea.Cmd{inputCommand, spinnerCommand}

	switch msg := msg.(type) {

	case FatalErrorMessage:
		m.fatalError = msg.err

	case AppStateMessage:
		switch msg.finishedState {
		case app.Initializing:
			m.appState = app.WaitingForPeers
			commands = append(commands, waitForWakuPeers(m.app))
		case app.WaitingForPeers:
			if config.InitialAction() == "" {
				m.appState = app.UserInput
			} else {
				cmd := runAction(&m, config.InitialAction())
				commands = append(commands, cmd)
			}
		case app.UserInput:
			break
		case app.CreatingSession, app.JoiningSession:
			m.appState = app.UserInput
			m.gameSessionID = m.app.GameSessionID()
			m.gameState = m.app.GameState()
			config.Logger.Debug("STATE FINISHED",
				zap.Any("finishedState", msg.finishedState),
				zap.Any("gameSessionID", m.gameSessionID),
				zap.Any("gameState", m.gameState),
			)
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
			cmd := runAction(&m, m.input.Value())
			m.input.Reset()
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}
	}

	return m, tea.Batch(commands...)
}

func (m model) View() string {
	var view string
	if m.fatalError != nil {
		view = fmt.Sprintf(" ☠️ FATAL ERROR: %s", m.fatalError)
	} else {
		view = m.renderAppState()
	}

	return fmt.Sprintf("%s\n\n%s", renderLogPath(), view)
}

func (m model) renderAppState() string {
	switch m.appState {
	case app.Idle:
		return fmt.Sprintf("nothing is happning. boring life.")
	case app.Initializing:
		return m.spinner.View() + " Starting Waku..."
	case app.WaitingForPeers:
		return m.spinner.View() + " Connecting to Waku peers ..."
	case app.UserInput:
		return m.renderGame()
	}

	return "unknown app state"
}

func (m model) renderGame() string {
	if m.gameSessionID == "" {
		return fmt.Sprintf(
			`  Join a game session or create a new one ...

%s
%s
`,
			m.input.View(),
			m.lastCommandError,
		)
	}

	if m.gameState == nil {
		return m.spinner.View() + " Waiting for initial game state ..."
	}

	return fmt.Sprintf(
		`  SESSION:      %s
  PLAYER:       %s
  VOTE ITEM:    %s

%s

%s
%s
`,
		m.gameSessionID,
		m.app.Game.Player().Name,
		render(&m.gameState.VoteItem),
		m.renderPlayers(),
		m.input.View(),
		m.lastCommandError,
	)
}

type PlayerVoteResult struct {
	Player protocol.Player
	Vote   string
}

func (m model) renderPlayers() string {
	var players []PlayerVoteResult

	for playerID, player := range m.gameState.Players {
		voteResult := PlayerVoteResult{
			Player: player,
			Vote:   "🤔",
		}
		if vote, ok := m.gameState.TempVoteResult[playerID]; ok {
			voteResult.Vote = fmt.Sprintf("%d", vote)
		}
		players = append(players, voteResult)
	}

	for playerID, vote := range m.gameState.TempVoteResult {
		if _, ok := m.gameState.Players[playerID]; ok {
			continue
		}
		players = append(players, PlayerVoteResult{
			Player: protocol.Player{
				ID:    playerID,
				Name:  string(playerID),
				Order: math.MaxInt32,
			},
			Vote: fmt.Sprintf("%d", vote),
		})
	}

	sort.Slice(players[:], func(i, j int) bool {
		playerI := players[i].Player
		playerJ := players[j].Player
		if playerI.Order != playerJ.Order {
			return playerI.Order < playerJ.Order
		}
		return playerI.Name < playerJ.Name
	})

	var votes []string
	for _, player := range players {
		votes = append(votes, player.Vote)
	}

	var playerNames []string
	for _, player := range players {
		playerNames = append(playerNames, player.Player.Name)
	}

	var CommonStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		//Background(lipgloss.Color("#7D56F4")).
		//PaddingTop(2).
		PaddingLeft(1).
		PaddingRight(1).
		//Width(22)
		Align(lipgloss.Center)

	var HeaderStyle = CommonStyle.Copy().
		Bold(true)

	rows := [][]string{
		votes,
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return HeaderStyle
			default:
				return CommonStyle
			}
		}).
		Headers(playerNames...).
		Rows(rows...)

	return t.String()
}

func (m model) getPlayerInfo(playerID protocol.PlayerID) protocol.Player {
	player, playerFound := m.gameState.Players[playerID]
	if playerFound {
		return player
	}
	return protocol.Player{
		ID:    playerID,
		Name:  string(playerID),
		Order: math.MaxInt32,
	}
}

func renderLogPath() string {
	return fmt.Sprintf("LOG FILE: file:///%s", config.LogFilePath)
}

func render(item *protocol.VoteItem) string {
	if item.URL == "" {
		return item.Name
	}
	if item.Name == "" {
		return item.URL
	}
	return fmt.Sprintf("%s (%s)", item.URL, item.Name)
}

// Ensure that model fulfils the tea.Model interface at compile time.
// ref: https://www.inngest.com/blog/interactive-clis-with-bubbletea
var _ tea.Model = (*model)(nil)
