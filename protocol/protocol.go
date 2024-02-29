package protocol

var Version int = 1

//
//type Player struct {
//	Name   string
//	Dealer bool
//}

type Player string
type VoteResult int

type State struct {
	Players []Player `json:"players"`
	VoteFor string   `json:"voteFor"`
}

type MessageType string

const (
	MessageTypeState        MessageType = "__state"
	MessageTypeStartVoting              = "__start_voting"
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
