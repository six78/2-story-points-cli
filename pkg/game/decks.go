package game

import (
	protocol2 "2sp/pkg/protocol"
	"golang.org/x/exp/maps"
)

var fibonacciDeck = protocol2.Deck{"1", "2", "3", "5", "8", "13", "21", "?"}

const Fibonacci = "fibonacci"

//var TShirtDeck = []protocol.VoteResult{
//	"XS", "S", "M", "L", "XL", "XXL",
//}

var decks = map[string]protocol2.Deck{
	Fibonacci: fibonacciDeck,
}

func GetDeck(deckName string) (protocol2.Deck, bool) {
	deck, ok := decks[deckName]
	return deck, ok
}

func AvailableDecks() []string {
	return maps.Keys(decks)
}

func CreateDeck(votes []string) protocol2.Deck {
	result := protocol2.Deck{}
	for _, value := range votes {
		result = append(result, protocol2.VoteValue(value))
	}
	return result
}
