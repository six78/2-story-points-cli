package view

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"go.uber.org/zap"
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
		return AppStateMessage{finishedState: app.Initializing}
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
		return AppStateMessage{finishedState: app.WaitingForPeers}
	}
}

func createNewSession(a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.CreateNewSession()
		if err != nil {
			return FatalErrorMessage{err}
		}
		return AppStateMessage{finishedState: app.CreatingSession}
	}
}

func joinSession(sessionID string, a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.JoinSession(sessionID)
		config.Logger.Debug("joining session", zap.String("sessionID", sessionID), zap.Error(err))
		if err != nil {
			return FatalErrorMessage{err}
		}
		return AppStateMessage{finishedState: app.JoiningSession}
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
