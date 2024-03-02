package view

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"waku-poker-planning/app"
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
		return AppStateMessage{nextState: app.WaitingForPeers}
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
		return AppStateMessage{nextState: app.Playing}
	}
}

func startGame(a *app.App) tea.Cmd {
	return func() tea.Msg {
		a.StartGame()
		return GameStateMessage{state: a.GameState()}
	}
}

func waitForGameState(app *app.App) tea.Cmd {
	return func() tea.Msg {
		state, more := app.WaitForGameState()
		if !more {
			return FatalErrorMessage{
				err: errors.New("game nextState subscription closed unexpectedly"),
			}
		}
		return GameStateMessage{state: state}
	}
}

// initial command: StartWaku
// waku started -> wait for peers
// peers connected -> start game (wait for game nextState)
