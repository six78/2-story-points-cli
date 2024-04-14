package messages

import (
	"waku-poker-planning/protocol"
	"waku-poker-planning/view/states"
)

type FatalErrorMessage struct {
	Err error
}

type AppStateMessage struct {
	ErrorMessage
	FinishedState states.AppState
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
