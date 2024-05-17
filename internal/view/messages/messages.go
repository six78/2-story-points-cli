package messages

import (
	"2sp/internal/waku"
	protocol2 "2sp/pkg/protocol"
	"2sp/view/states"
)

type FatalErrorMessage struct {
	Err error
}

type AppStateFinishedMessage struct {
	State states.AppState
}

type AppStateMessage struct {
	State states.AppState
}

type GameStateMessage struct {
	State *protocol2.State
}

type ErrorMessage struct {
	Err error
}

func NewErrorMessage(err error) ErrorMessage {
	return ErrorMessage{Err: err}
}

type PlayerIDMessage struct {
	PlayerID protocol2.PlayerID
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
	RoomID   protocol2.RoomID
	IsDealer bool
}

// TODO: Try to find a better solution, probably game.subscribeToMyVote().
// With this message the logic is duplicated in Game and Model.
type MyVote struct {
	Result protocol2.VoteResult
}

type EnableEnterKey struct {
}
