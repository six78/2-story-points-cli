package game

import (
	"github.com/six78/2-story-points-cli/pkg/protocol"
	"golang.org/x/exp/maps"
)

const (
	DefaultDeck   = FibonacciDeck
	FibonacciDeck = "fibonacci"
	PriorityDeck  = "priority"
)

//var TShirtDeck = []protocol.VoteResult{
//	"XS", "S", "M", "L", "XL", "XXL",
//}

var decks = map[string]protocol.Deck{
	FibonacciDeck: {"1", "2", "3", "5", "8", "13", "21", "?"},
	PriorityDeck:  {"4", "3", "2", "1", "0", "?"},
}

func GetDeck(deckName string) (protocol.Deck, bool) {
	deck, ok := decks[deckName]
	return deck, ok
}

func AvailableDecks() []string {
	return maps.Keys(decks)
}

func CreateDeck(votes []string) protocol.Deck {
	result := protocol.Deck{}
	for _, value := range votes {
		result = append(result, protocol.VoteValue(value))
	}
	return result
}
