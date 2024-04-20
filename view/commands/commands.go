package commands

import (
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	"waku-poker-planning/app"
	"waku-poker-planning/view/messages"
	"waku-poker-planning/view/states"
)

// Any command here must:
// 	1. Get App as argument
// 	2. Return tea.Cmd

func InitializeApp(a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Initialize()
		if err != nil {
			return messages.FatalErrorMessage{Err: err}
		}
		return messages.AppStateFinishedMessage{State: states.Initializing}
	}
}

func CreateNewRoom(a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Game.CreateNewRoom()
		return messages.AppStateFinishedMessage{
			State: states.CreatingRoom,
			ErrorMessage: messages.ErrorMessage{
				Err: err,
			},
		}
	}
}

func JoinRoom(roomID string, a *app.App) tea.Cmd {
	return func() tea.Msg {
		err := a.Game.JoinRoom(roomID)
		return messages.AppStateFinishedMessage{
			State: states.JoiningRoom,
			ErrorMessage: messages.ErrorMessage{
				Err: err,
			},
		}
	}
}

func WaitForGameState(app *app.App) tea.Cmd {
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

func ToggleRoomView(currentRoomView states.RoomView) tea.Cmd {
	return func() tea.Msg {
		var nextRoomView states.RoomView
		switch currentRoomView {
		case states.ActiveIssueView:
			nextRoomView = states.IssuesListView
		case states.IssuesListView:
			nextRoomView = states.ActiveIssueView
		}
		return messages.RoomViewChange{RoomView: nextRoomView}
	}
}

func WaitForConnectionStatus(app *app.App) tea.Cmd {
	return func() tea.Msg {
		status, more, err := app.WaitForConnectionStatus()
		if err != nil {
			return messages.FatalErrorMessage{Err: err}
		}
		if !more {
			return nil
		}
		return messages.ConnectionStatus{
			Status: status,
		}
	}
}

func JoinRoomFromClipboard(app *app.App) tea.Cmd {
	return func() tea.Msg {
		if clipboard.Unsupported {
			err := errors.New("clipboard is unsupported")
			return messages.NewErrorMessage(err)
		}
		roomID, err := clipboard.ReadAll()
		if err != nil {
			err := errors.Wrap(err, "failed to read from clipboard")
			return messages.NewErrorMessage(err)
		}
		return JoinRoom(roomID, app)()
	}
}
