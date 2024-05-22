package commands

import (
	"2sp/internal/app"
	"2sp/internal/view/messages"
	"2sp/internal/view/states"
	"2sp/pkg/protocol"
	tea "github.com/charmbracelet/bubbletea"
	"time"
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
		room, initialState, err := a.Game.CreateNewRoom()
		if err != nil {
			return messages.NewErrorMessage(err)
		}

		roomID, err := room.ToRoomID()
		if err != nil {
			return messages.NewErrorMessage(err)
		}

		err = a.Game.JoinRoom(roomID, initialState)
		if err != nil {
			return messages.NewErrorMessage(err)
		}

		return messages.RoomJoin{
			RoomID:   a.Game.RoomID(),
			IsDealer: a.Game.IsDealer(),
		}
	}
}

func JoinRoom(a *app.App, roomID protocol.RoomID) tea.Cmd {
	return func() tea.Msg {
		err := a.Game.JoinRoom(roomID, nil)
		if err != nil {
			return messages.NewErrorMessage(err)
		}
		return messages.RoomJoin{
			RoomID:   a.Game.RoomID(),
			IsDealer: a.Game.IsDealer(),
		}
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

func PublishVote(app *app.App, vote protocol.VoteValue) tea.Cmd {
	return func() tea.Msg {
		err := app.Game.PublishVote(vote)
		if err != nil {
			return messages.NewErrorMessage(err)
		}
		// TODO: Send err=nil ErrorMessage here
		return messages.MyVote{
			Result: app.Game.MyVote(),
		}
	}
}

func FinishVoting(app *app.App, result protocol.VoteValue) tea.Cmd {
	return func() tea.Msg {
		err := app.Game.Finish(result)
		return messages.NewErrorMessage(err)
	}
}

func AddIssue(app *app.App, urlOrTitle string) tea.Cmd {
	return func() tea.Msg {
		_, err := app.Game.AddIssue(urlOrTitle)
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

func DelayMessage(timeout time.Duration, msg tea.Msg, restart chan struct{}) tea.Cmd {
	return func() tea.Msg {
		for {
			select {
			case <-time.After(timeout):
				return msg
			case <-restart:
			}
		}
	}
}
