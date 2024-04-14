package states

type AppState int

const (
	Idle AppState = iota
	Initializing
	InputPlayerName
	WaitingForPeers
	CreatingRoom
	JoiningRoom
	InsideRoom
)
