package protocol

var Version int = 1

//
//type Player struct {
//	Name   string
//	Dealer bool
//}

/*
TODO:
	{
		"players": [
			"hoho",
			"player-1709404531"
		],
		"voteItem": {
			"id": "ef068cc3-8a2b-42bf-ac51-8ca46123d4c0",
			"name": "aa"
		},
		"tempVoteResults": {
			"hoho": 1
		}
	}
*/

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

type PlayerVote struct {
	voteBy     Player     `json:"voteBy"`
	voteFor    string     `json:"voteFor"`
	voteResult VoteResult `json:"voteResult"`
}
