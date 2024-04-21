package commands

import (
	tea "github.com/charmbracelet/bubbletea"
	"waku-poker-planning/app"
	"waku-poker-planning/protocol"
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
		if err != nil {
			return messages.NewErrorMessage(err)
		}
		return messages.RoomJoin{
			RoomID:   a.Game.RoomID(),
			IsDealer: a.Game.IsDealer(),
		}
	}
}

func JoinRoom(a *app.App, roomID protocol.RoomID, state *protocol.State) tea.Cmd {
	return func() tea.Msg {
		var err error
		if state == nil {
			// Check storage for state
			state, err = a.LoadRoomState(roomID)
			if err != nil {
				return messages.NewErrorMessage(err)
			}
		}
		err = a.Game.JoinRoom(roomID, state)
		if err != nil {
			return messages.NewErrorMessage(err)
		}
		return messages.RoomJoin{
			RoomID:   a.Game.RoomID(),
			IsDealer: a.Game.IsDealer(),
		}
	}
}

func WaitForGameState(app *app.App) tea.Cmd {
	return func() tea.Msg {
		state, more, err := app.WaitForGameState()
		if err != nil {
			return messages.FatalErrorMessage{Err: err}
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

func PublishVote(app *app.App, vote protocol.VoteValue) tea.Cmd {
	return func() tea.Msg {
		err := app.Game.PublishVote(vote)
		return messages.NewErrorMessage(err)
	}
}

func SelectIssue(app *app.App, index int) tea.Cmd {
	return func() tea.Msg {
		err := app.Game.SelectIssue(index)
		return messages.NewErrorMessage(err)
	}
}

func QuitApp(app *app.App) tea.Cmd {
	return func() tea.Msg {
		app.Game.LeaveRoom()
		return tea.Quit()
	}
}
