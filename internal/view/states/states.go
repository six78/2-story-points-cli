package states

type AppState int

const (
	Idle AppState = iota
	Initializing
	InputPlayerName
	WaitingForPeers
	Playing
)

type RoomView int

const (
	ActiveIssueView RoomView = iota
	IssuesListView
)
