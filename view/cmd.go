package view

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	"strings"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
)

// Any command here must:
// 	1. Get App as argument
// 	2. Return tea.Cmd

func initializeApp(a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Initialize()
		if err != nil {
			return FatalErrorMessage{err}
		}
		return AppStateMessage{finishedState: Initializing}
	}
}

func waitForWakuPeers(a *app.App) tea.Cmd {
	return func() tea.Msg {
		ok := a.WaitForPeersConnected()
		if !ok {
			return FatalErrorMessage{
				err: errors.New("failed to connect to peers"),
			}
		}
		return AppStateMessage{finishedState: WaitingForPeers}
	}
}

func createNewRoom(a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Game.CreateNewRoom()
		return AppStateMessage{
			finishedState: CreatingRoom,
			ActionErrorMessage: ActionErrorMessage{
				err: err,
			},
		}
	}
}

func joinRoom(roomID string, a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Game.JoinRoom(roomID)
		return AppStateMessage{
			finishedState: JoiningRoom,
			ActionErrorMessage: ActionErrorMessage{
				err: err,
			},
		}
	}
}

func waitForGameState(app *app.App) tea.Cmd {
	return func() tea.Msg {
		state, more, err := app.WaitForGameState()
		if err != nil {
			return FatalErrorMessage{err}
		}
		if !more {
			return nil
		}
		return GameStateMessage{state: state}
	}
}

func processUserInput(m *model) tea.Cmd {
	defer m.input.Reset()
	return processInput(m)
}

func processInput(m *model) tea.Cmd {
	if m.state == InputPlayerName {
		defer m.input.Reset()
		cmd, _ := processPlayerNameInput(m, m.input.Value())
		return cmd
	}

	if m.state == UserAction {
		defer m.input.Reset()
		return processAction(m, m.input.Value())
	}

	return nil
}

func processAction(m *model, action string) tea.Cmd {
	var err error
	defer func() {
		config.Logger.Debug("user action processed",
			zap.Error(err),
			zap.Any("state", m.state),
		)

		if err != nil {
			m.lastCommandError = err.Error()
		} else {
			m.lastCommandError = ""
		}
	}()

	args := strings.Fields(action)
	if len(args) == 0 {
		err = errors.New("empty action")
		return nil
	}

	commandRoot := Action(args[0])
	commandFn, ok := actions[commandRoot]
	if !ok {
		err = fmt.Errorf("unknown action: %s", commandRoot)
		return nil
	}

	cmd, err := commandFn(m, args[1:])

	return cmd
}
