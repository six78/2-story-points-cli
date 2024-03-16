package view

type State int

const (
	Idle State = iota
	Initializing
	InputPlayerName
	WaitingForPeers
	UserAction
	CreatingRoom
	JoiningRoom
)
