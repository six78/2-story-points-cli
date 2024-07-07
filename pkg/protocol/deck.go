package protocol

import "golang.org/x/exp/slices"

// UncertaintyCard is a special value that means the player wasn't sure about the vote.
// NOTE: For now I've chosen the simplest way of implementing this.
// - Option 1. Deck flag --with-uncertainty-card.
// It will automatically work for any Deck and will always be at the end of the deck.
// It would require changing the Deck type itself, simple []VoteValue inheritance won't work anymore.
// - Option 2 (chosen at least for now). Simply treat some special symbol as UncertaintyCard.
// User can put it anywhere in the deck.
// To make it more customizable, we can give user an option to specify the special symbol.
const UncertaintyCard = VoteValue("?")

type Deck []VoteValue

func (d Deck) Index(value VoteValue) int {
	return slices.Index(d, value)
}
