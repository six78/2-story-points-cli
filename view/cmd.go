package view

import (
	"errors"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
	"strings"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
	"waku-poker-planning/view/messages"
	"waku-poker-planning/view/states"
)

// Any command here must:
// 	1. Get App as argument
// 	2. Return tea.Cmd

func initializeApp(a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Initialize()
		if err != nil {
			return messages.FatalErrorMessage{Err: err}
		}
		return messages.AppStateMessage{FinishedState: states.Initializing}
	}
}

func waitForWakuPeers(a *app.App) tea.Cmd {
	return func() tea.Msg {
		ok := a.WaitForPeersConnected()
		if !ok {
			return messages.FatalErrorMessage{
				Err: errors.New("failed to connect to peers"),
			}
		}
		return messages.AppStateMessage{FinishedState: states.WaitingForPeers}
	}
}

func createNewRoom(a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Game.CreateNewRoom()
		return messages.AppStateMessage{
			FinishedState: states.CreatingRoom,
			ErrorMessage: messages.ErrorMessage{
				Err: err,
			},
		}
	}
}

func joinRoom(roomID string, a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Game.JoinRoom(roomID)
		return messages.AppStateMessage{
			FinishedState: states.JoiningRoom,
			ErrorMessage: messages.ErrorMessage{
				Err: err,
			},
		}
	}
}

func waitForGameState(app *app.App) tea.Cmd {
	return func() tea.Msg {
		state, more, err := app.WaitForGameState()
		if err != nil {
			return messages.FatalErrorMessage{err}
		}
		if !more {
			return nil
		}
		return messages.GameStateMessage{State: state}
	}
}

func processUserInput(m *model) tea.Cmd {
	defer m.input.Reset()
	return processInput(m)
}

func processInput(m *model) tea.Cmd {
	if m.state == states.InputPlayerName {
		defer m.input.Reset()
		return processPlayerNameInput(m, m.input.Value())
	}

	if m.state == states.InsideRoom {
		defer m.input.Reset()
		return processAction(m, m.input.Value())
	}

	return nil
}

func processAction(m *model, action string) tea.Cmd {
	defer func() {
		config.Logger.Debug("user action processed",
			zap.Any("state", m.state),
		)
	}()

	args := strings.Fields(action)
	if len(args) == 0 {
		return nil
	}

	commandRoot := Action(args[0])
	commandFn, ok := actions[commandRoot]

	if !ok {
		return func() tea.Msg {
			err := fmt.Errorf("unknown action: %s", commandRoot)
			return messages.NewErrorMessage(err)
		}
	}

	return commandFn(m, args[1:])
}
