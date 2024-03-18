package game

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
	"time"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
)

func TestStateSize(t *testing.T) {
	const playersCount = 20
	const issuesCount = 30

	state := protocol.State{
		Players:  make(map[protocol.PlayerID]protocol.Player, playersCount),
		VoteList: make(map[protocol.VoteItemID]*protocol.VoteItem, issuesCount),
	}

	votes := make(map[protocol.PlayerID]*protocol.VoteResult, playersCount)
	deck, deckFound := GetDeck(Fibonacci)
	require.True(t, deckFound)

	state.Deck = deck

	for i := 0; i < playersCount; i++ {
		playerID, err := config.GeneratePlayerID()
		require.NoError(t, err)

		state.Players[playerID] = protocol.Player{
			ID:       playerID,
			Name:     fmt.Sprintf("player-%d", i),
			IsDealer: i == 0,
			Order:    i,
		}

		result := protocol.VoteResult(i % len(deck))
		votes[playerID] = &result
	}

	for i := 0; i < issuesCount; i++ {
		voteItemID, err := config.GenerateVoteItemID()
		require.NoError(t, err)

		state.VoteList[voteItemID] = &protocol.VoteItem{
			ID:     voteItemID,
			Text:   fmt.Sprintf("https://github.com/six78/waku-poker-planing/issues/%d", i),
			Votes:  votes, // same votes for each issue, whatever
			Result: &deck[i%len(deck)],
		}
	}

	stateMessage, err := json.Marshal(state)
	require.NoError(t, err)

	fmt.Println("state size", "bytes", len(stateMessage))
	require.True(t, len(stateMessage) < 100*1024, "state size should be less than 100 kilobytes")
}

func TestSimpleGame(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var err error
	config.Logger, err = zap.NewDevelopment()
	require.NoError(t, err)

	transport := NewTransportMock(t)
	game := NewGame(ctx, transport, "player-1")

	err = game.CreateNewRoom()
	require.NoError(t, err)

	room := game.Room()
	sub, err := transport.SubscribeToMessages(&room)
	require.NoError(t, err)

	game.Start()

	const firstItemText = "a"
	const dealerVote = protocol.VoteResult(1)

	var fisrtVoteItemID protocol.VoteItemID

	checkVoteItems := func(t *testing.T, voteList map[protocol.VoteItemID]*protocol.VoteItem) *protocol.VoteItem {
		require.Len(t, voteList, 1)
		item, ok := voteList[fisrtVoteItemID]
		require.True(t, ok)
		require.NotNil(t, item)
		require.Equal(t, firstItemText, item.Text)
		return item
	}

	{ // Deal first vote item
		fisrtVoteItemID, err = game.Deal(firstItemText)
		require.NoError(t, err)

		state := expectState(t, sub, nil)
		item := checkVoteItems(t, state.VoteList)
		require.Nil(t, item.Result)
		require.Len(t, item.Votes, 0)
	}

	currentVoteItem, ok := game.CurrentState().VoteList[game.CurrentState().CurrentVoteItemID]
	require.True(t, ok)
	require.NotNil(t, currentVoteItem)
	require.Equal(t, firstItemText, currentVoteItem.Text)

	{ // Publish dealer vote
		err = game.PublishVote(dealerVote)
		require.NoError(t, err)

		playerVote := expectPlayerVote(t, sub)
		require.Equal(t, game.Player().ID, playerVote.VoteBy)
		require.Equal(t, currentVoteItem.ID, playerVote.VoteFor)
		require.Equal(t, dealerVote, playerVote.VoteResult)

		state := expectState(t, sub, func(state *protocol.State) bool {
			item, ok := state.VoteList[fisrtVoteItemID]
			if !ok {
				return false
			}
			_, ok = item.Votes[game.Player().ID]
			return ok
		})
		// FIXME: check vote state
		item := checkVoteItems(t, state.VoteList)
		require.Nil(t, item.Result)
		require.Len(t, item.Votes, 1)

		vote, ok := item.Votes[game.Player().ID]
		require.True(t, ok)
		require.Nil(t, vote)
	}

	{ // Reveal votes
		err = game.Reveal()
		require.NoError(t, err)

		state := expectState(t, sub, nil)
		item := checkVoteItems(t, state.VoteList)
		require.Nil(t, item.Result)
		require.Len(t, item.Votes, 1)

		vote, ok := item.Votes[game.Player().ID]
		require.True(t, ok)
		require.NotNil(t, vote)
		require.Equal(t, dealerVote, *vote)
	}

	const votingResult = protocol.VoteResult(1)

	{ // Finish voting
		err = game.Finish(votingResult)
		require.NoError(t, err)

		state := expectState(t, sub, nil)
		item := checkVoteItems(t, state.VoteList)
		require.NotNil(t, item.Result)
		require.Equal(t, *item.Result, votingResult)
		require.Len(t, item.Votes, 1)

		vote, ok := item.Votes[game.Player().ID]
		require.True(t, ok)
		require.Equal(t, dealerVote, *vote)
	}

	const secondItemText = "b"
	var secondVoteItemID protocol.VoteItemID

	checkVoteItems = func(t *testing.T, voteList map[protocol.VoteItemID]*protocol.VoteItem) *protocol.VoteItem {
		require.Len(t, voteList, 2)

		item, ok := voteList[fisrtVoteItemID]
		require.True(t, ok)
		require.NotNil(t, item)
		require.Equal(t, firstItemText, item.Text)

		item, ok = voteList[secondVoteItemID]
		require.True(t, ok)
		require.NotNil(t, item)
		require.Equal(t, secondItemText, item.Text)

		return item
	}

	{ // Deal another vote item
		secondVoteItemID, err = game.Deal(secondItemText)
		require.NoError(t, err)

		state := expectState(t, sub, nil)
		item := checkVoteItems(t, state.VoteList)
		require.Nil(t, item.Result)
		require.Len(t, item.Votes, 0)
	}
}

func expectState(t *testing.T, sub chan []byte, cb func(*protocol.State) bool) *protocol.State {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "timeout waiting for state message")

		case payload, more := <-sub:
			require.True(t, more)
			require.NotNil(t, payload)

			message, err := protocol.UnmarshalMessage(payload)
			require.NoError(t, err)

			if message.Type != protocol.MessageTypeState {
				continue
			}

			state, err := protocol.UnmarshalState(payload)
			require.NoError(t, err)

			if cb == nil {
				return state
			}

			if cb(state) {
				return state
			}
		}
	}
}

func expectPlayerVote(t *testing.T, sub chan []byte) *protocol.PlayerVoteMessage {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "timeout waiting for state message")

		case payload, more := <-sub:
			require.True(t, more)
			require.NotNil(t, payload)

			message, err := protocol.UnmarshalMessage(payload)
			require.NoError(t, err)

			if message.Type != protocol.MessageTypePlayerVote {
				continue
			}

			vote, err := protocol.UnmarshalPlayerVote(payload)
			require.NoError(t, err)

			return vote
		}
	}
}
