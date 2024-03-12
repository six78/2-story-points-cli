package game

import (
	"golang.org/x/exp/maps"
	"waku-poker-planning/protocol"
)

var fibonacciDeck = []protocol.VoteResult{
	1, 2, 3, 5, 8, 13, 21, 34, 55, 89,
}

//var TShirtDeck = []protocol.VoteResult{
//	"XS", "S", "M", "L", "XL", "XXL",
//}

var decks = map[string][]protocol.VoteResult{
	"fibonacci": fibonacciDeck,
}

func GetDeck(deckName string) ([]protocol.VoteResult, bool) {
	deck, ok := decks[deckName]
	return deck, ok
}

func AvailableDecks() []string {
	return maps.Keys(decks)
}
