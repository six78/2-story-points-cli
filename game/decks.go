package game

import (
	"golang.org/x/exp/maps"
	"waku-poker-planning/protocol"
)

var fibonacciDeck = []protocol.VoteValue{
	"1", "2", "3", "5", "8", "13", "21", "34", "55", "89",
}

const Fibonacci = "fibonacci"

//var TShirtDeck = []protocol.VoteResult{
//	"XS", "S", "M", "L", "XL", "XXL",
//}

var decks = map[string][]protocol.VoteValue{
	Fibonacci: fibonacciDeck,
}

func GetDeck(deckName string) ([]protocol.VoteValue, bool) {
	deck, ok := decks[deckName]
	return deck, ok
}

func AvailableDecks() []string {
	return maps.Keys(decks)
}
