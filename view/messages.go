package view

import (
	"waku-poker-planning/app"
	"waku-poker-planning/protocol"
)

type FatalErrorMessage struct {
	err error
}

type AppStateMessage struct {
	nextState app.State
}

type GameStateMessage struct {
	state protocol.State
}
