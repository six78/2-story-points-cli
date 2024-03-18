package protocol

const Version byte = 1

type PlayerID string
type VoteItemID string
type VoteResult int // WARNING: change type to string. Empty string means vote is not revealed.

// TODO: Come up with a common approach for pointers.
// 		 One solution is to make VoteResult a string.

type Player struct {
	ID       PlayerID
	Name     string
	IsDealer bool
	Order    int
}

type VoteItem struct {
	ID     VoteItemID               `json:"id"`
	Text   string                   `json:"url"` // In most cases text will be a URL
	Votes  map[PlayerID]*VoteResult `json:"votes"`
	Result *VoteResult              `json:"result"`
	Order  int                      `json:"order"`
}

type VoteState string

const (
	IdleState     VoteState = "idle"
	VotingState   VoteState = "voting"
	RevealedState VoteState = "revealed"
	FinishedState VoteState = "finished"
)

type State struct {
	Players   map[PlayerID]Player `json:"players"`
	VoteState VoteState           `json:"voteState"`

	// Deprecated: TempVoteResults is deprecated. Use VoteList[CurrentVoteItemID] instead.
	TempVoteResult map[PlayerID]*VoteResult `json:"tempVoteResults"`

	// Deprecated: VoteItem is deprecated. Use CurrentVoteItemID and VoteList instead.
	VoteItem VoteItem `json:"voteItem"`

	CurrentVoteItemID VoteItemID               `json:"currentVoteItemID"`
	VoteList          map[VoteItemID]*VoteItem `json:"voteList"` // add order to voteitem

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

type Deck []VoteResult
