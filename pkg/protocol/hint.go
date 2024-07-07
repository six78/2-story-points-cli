package protocol

type Hint struct {
	// Acceptable shows if the voting for given issue can be considered as "acceptable".
	// It will be false if the variety of votes is too high. In this case Advice will contain
	// a suggestion to discuss and re-vote.
	Acceptable bool

	// Value is the recommended value for the issue.
	// It's guaranteed to be one of the values from the deck.
	Value VoteValue

	// Description contains text message for the team.
	// When Acceptable is false, Description explaining the reject reasoning.
	// When Acceptable is true, Description contains some congratulatory message.
	Description string
}
