package view

type State int

const (
	Idle State = iota
	Initializing
	InputPlayerName
	WaitingForPeers
	UserAction
	CreatingSession
	JoiningSession
)
