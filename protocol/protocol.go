package protocol

var Version int = 1

//
//type Player struct {
//	Name   string
//	Dealer bool
//}

type Player string
type VoteResult int

type VoteItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type State struct {
	Players        []Player              `json:"players"`
	VoteItem       VoteItem              `json:"voteItem"`
	TempVoteResult map[Player]VoteResult `json:"tempVoteResults"`
}

type MessageType string

const (
	MessageTypeState        MessageType = "__state"
	MessageTypePlayerOnline             = "__player_online"
	MessageTypePlayerVote               = "__player_vote"
)

type Message struct {
	Type       MessageType `json:"type"`
	State      *State      `json:"state,omitempty"`
	VoteFor    string      `json:"voteFor,omitempty"`
	Name       Player      `json:"name,omitempty"`
	VoteResult VoteResult  `json:"voteResult,omitempty"`
}

type PlayerVote struct {
	VoteBy     Player     `json:"voteBy"`
	VoteFor    string     `json:"voteFor"`
	VoteResult VoteResult `json:"voteResult"`
}
