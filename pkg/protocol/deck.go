package protocol

import "golang.org/x/exp/slices"

type Deck []VoteValue

func (d Deck) Index(value VoteValue) int {
	return slices.Index(d, value)
}
