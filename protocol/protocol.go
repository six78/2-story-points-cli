package protocol

const Version byte = 1

type PlayerID string
type VoteResult int

type Player struct {
	ID       PlayerID
	Name     string
	IsDealer bool
	Order    int
}

type VoteItem struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	URL         string                  `json:"url"`
	Result      VoteResult              `json:"result"`
	VoteHistory map[PlayerID]VoteResult `json:"voteHistory"`
}

type VoteState string

const (
	IdleState     VoteState = "idle"
	VotingState             = "voting"
	RevealedState           = "revealed"
	FinishedState           = "finished"
)

type State struct {
	Players        map[PlayerID]Player      `json:"players"`
	VoteItem       VoteItem                 `json:"voteItem"`
	TempVoteResult map[PlayerID]*VoteResult `json:"tempVoteResults"`
	VoteState      VoteState                `json:"voteState"`
	Deck           Deck                     `json:"deck"`
}

type MessageType string

const (
	MessageTypeState        MessageType = "__state"
	MessageTypePlayerOnline             = "__player_online"
	MessageTypePlayerVote               = "__player_vote"
)

type Message struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
}

type GameStateMessage struct {
	Message
	State State `json:"state"`
}

type PlayerOnlineMessage struct {
	Message
	Player Player `json:"name,omitempty"`
}

type PlayerVote struct {
	Message
	VoteBy     PlayerID   `json:"voteBy"`
	VoteFor    string     `json:"voteFor"`
	VoteResult VoteResult `json:"voteResult"`
}

type Deck []VoteResult
