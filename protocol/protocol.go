package protocol

var Version int = 1

//
//type Player struct {
//	Name   string
//	Dealer bool
//}

type Player string
type PlayerID string
type VoteResult int

type VoteItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
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
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
}

type GameStateMessage struct {
	Message
	State State `json:"state"`
}

type PlayerOnlineMessage struct {
	Message
	Name Player `json:"name,omitempty"`
}

type PlayerVote struct {
	Message
	VoteBy     Player     `json:"voteBy"`
	VoteFor    string     `json:"voteFor"`
	VoteResult VoteResult `json:"voteResult"`
}
