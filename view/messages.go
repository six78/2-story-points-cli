package view

import (
	"waku-poker-planning/app"
	"waku-poker-planning/protocol"
)

type FatalErrorMessage struct {
	err error
}

type AppStateMessage struct {
	finishedState app.State
}

type GameStateMessage struct {
	state *protocol.State
}
