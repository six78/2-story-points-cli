package protocol

const Version byte = 1

type PlayerID string
type IssueID string

type Issue struct {
	ID         IssueID    `json:"id"`
	TitleOrURL string     `json:"titleOrUrl"`
	Votes      IssueVotes `json:"votes"`
	Result     *VoteValue `json:"result"` // NOTE: keep pointer. Because "empty string means vote is not revealed"
	Hint       *Hint      `json:"-"`
}

type MessageType string

const (
	MessageTypeState         MessageType = "__state"
	MessageTypePlayerOnline  MessageType = "__player_online"
	MessageTypePlayerVote    MessageType = "__player_vote"
	MessageTypePlayerOffline MessageType = "__player_left"
)

type Message struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"updatedAt"` // WARNING: rename to Timestamp
}

type GameStateMessage struct {
	Message
	State State `json:"state"`
}

type PlayerOnlineMessage struct {
	Message
	Player Player `json:"player,omitempty"`
}

type PlayerOfflineMessage struct {
	Message
	Player Player `json:"player,omitempty"`
}

type PlayerVoteMessage struct {
	Message
	PlayerID   PlayerID   `json:"playerId"`
	Issue      IssueID    `json:"issue"`
	VoteResult VoteResult `json:"vote"`
}

type Deck []VoteValue

type IssueVotes map[PlayerID]VoteResult
