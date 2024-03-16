package view

import (
	"waku-poker-planning/protocol"
)

type FatalErrorMessage struct {
	err error
}

type AppStateMessage struct {
	finishedState State
}

type GameStateMessage struct {
	state *protocol.State
}

type ActionErrorMessage struct {
	err error
}
