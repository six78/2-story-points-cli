package protocol

const Version byte = 1

type PlayerID string
type VoteItemID string

type Player struct {
	ID   PlayerID `json:"id"`
	Name string   `json:"name"`
}

type Issue struct {
	ID         VoteItemID              `json:"id"`
	TitleOrURL string                  `json:"titleOrUrl"`
	Votes      map[PlayerID]VoteResult `json:"votes"`
	Result     *VoteValue              `json:"result"` // NOTE: keep pointer. Because "empty string means vote is not revealed"
}

type VoteState string

const (
	IdleState     VoteState = "idle"
	VotingState   VoteState = "voting"
	RevealedState VoteState = "revealed"
	FinishedState VoteState = "finished"
)

//type VoteStateDidukh struct {
//	Issue          Issue
//	Revealed       bool
//	TempVoteResult map[PlayerID]*VoteResult
//}

// TODO:  Vote -> Estimate ?

type State struct {
	Players   []Player  `json:"players"`
	VoteState VoteState `json:"voteState"`

	ActiveIssue VoteItemID `json:"activeIssue"`
	Issues      IssuesList `json:"issues"`

	Deck Deck `json:"deck"`
}

type MessageType string

const (
	MessageTypeState        MessageType = "__state"
	MessageTypePlayerOnline MessageType = "__player_online"
	MessageTypePlayerVote   MessageType = "__player_vote"
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

type PlayerVoteMessage struct {
	Message
	VoteBy     PlayerID   `json:"voteBy"`
	VoteFor    VoteItemID `json:"voteFor"`
	VoteResult VoteResult `json:"voteResult"` // TODO: rename to `voteValue`
}

type Deck []VoteValue
