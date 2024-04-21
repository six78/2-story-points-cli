package messages

import (
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/states"
	"waku-poker-planning/waku"
)

type FatalErrorMessage struct {
	Err error
}

type AppStateFinishedMessage struct {
	ErrorMessage
	State states.AppState
}

type AppStateMessage struct {
	State states.AppState
}

type GameStateMessage struct {
	State *protocol.State
}

type ErrorMessage struct {
	Err error
}

func NewErrorMessage(err error) ErrorMessage {
	return ErrorMessage{Err: err}
}

type PlayerIDMessage struct {
	PlayerID protocol.PlayerID
}

type RoomViewChange struct {
	RoomView states.RoomView
}

type ConnectionStatus struct {
	Status waku.ConnectionStatus
}

type CommandModeChange struct {
	CommandMode bool
}

type RoomJoin struct {
	RoomID   protocol.RoomID
	IsDealer bool
}

// TODO: Try to find a better solution, probably game.subscribeToMyVote().
// With this message the logic is duplicated in Game and Model.
type MyVote struct {
	Result protocol.VoteResult
}

type EnableEnterKey struct {
}
