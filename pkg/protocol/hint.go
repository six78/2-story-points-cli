package protocol

type Hint struct {
	// Acceptable shows if the voting for given issue can be considered as "acceptable".
	// It will be false if the variety of votes is too high. In this case Advice will contain
	// a suggestion to discuss and re-vote.
	Acceptable bool

	// RejectReason contains an explanation of why the vote is not acceptable.
	// When Acceptable is true, RejectReason is empty.
	RejectReason string

	// Value is the recommended value for the issue.
	// It's guaranteed to be one of the values from the deck.
	Value VoteValue

	// Advice is a text advice for the team about current vote.
	// It might contain players mentions in form "@<id>", where <id> a particular player ID.
	Advice string
}
