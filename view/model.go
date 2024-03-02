package view

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.uber.org/zap"
	"strconv"
	"strings"
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
	appState   app.State
	fatalError error
	gameState  protocol.State

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
		appState:  app.Initializing,
		gameState: a.GameState(),
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
		m.appState = msg.nextState
		switch msg.nextState {
		case app.WaitingForPeers:
			commands = append(commands, waitForWakuPeers(m.app))
		case app.Playing:
			commands = append(commands, startGame(m.app))
		}

	case GameStateMessage:
		m.gameState = msg.state
		commands = append(commands, waitForGameState(m.app))

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
					m.app.PublishUserOnline(onlineUser)
					return nil
				})
			}

			if strings.HasPrefix(userCommand, "rename") {
				user := strings.TrimPrefix(userCommand, "rename")
				user = strings.Trim(user, " ")
				commands = append(commands, func() tea.Msg {
					config.PlayerName = user
					m.app.PublishUserOnline(config.PlayerName)
					return nil
				})
			}

			if strings.HasPrefix(userCommand, "vote") {
				voteString := strings.TrimPrefix(userCommand, "vote")
				voteString = strings.Trim(voteString, " ")
				vote, err := strconv.Atoi(voteString)
				if err != nil {
					//m.commandError
					config.Logger.Error("failed to parse vote", zap.Error(err))
				} else {
					commands = append(commands, func() tea.Msg {
						m.app.PublishVote(vote)
						return nil
					})
				}
			}

			m.input.Reset()
		}
	}

	return m, tea.Batch(commands...)
}

func (m model) View() string {
	if m.fatalError != nil {
		return fmt.Sprintf(" ❌FATAL ERROR: %s", m.fatalError)
	}

	switch m.appState {
	case app.Idle:
		return fmt.Sprintf("nothing is happning. boring life.")
	case app.Initializing:
		return m.spinner.View() + " Starting Waku..."
	case app.WaitingForPeers:
		return m.spinner.View() + " Connecting to Waku peers ..."
	case app.Playing:
		return m.renderGame()
	}

	return "unknown app state"
}

func (m model) renderGame() string {
	players, err := json.Marshal(m.gameState.Players)
	if err != nil {
		panic(err)
	}

	voteItem, err := json.Marshal(m.gameState.VoteItem)
	if err != nil {
		panic(err)
	}

	voteResult, err := json.Marshal(m.gameState.TempVoteResult)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf(`
  LOG FILE:     %s

  PLAYERS:      %s

  VOTE ITEM:    %s

  VOTE RESULT:  %s

%s

`,
		"file:///"+config.LogFilePath,
		players,
		voteItem,
		voteResult,
		m.input.View(),
	)
}

// Ensure that model fulfils the tea.Model interface at compile time.
// ref: https://www.inngest.com/blog/interactive-clis-with-bubbletea
var _ tea.Model = (*model)(nil)
