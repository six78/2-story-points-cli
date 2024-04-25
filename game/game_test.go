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
		Players: make([]protocol.Player, 0, playersCount),
		Issues:  make(protocol.IssuesList, 0, issuesCount),
	}

	votes := make(map[protocol.PlayerID]protocol.VoteResult, playersCount)
	deck, deckFound := GetDeck(Fibonacci)
	require.True(t, deckFound)

	state.Deck = deck

	for i := 0; i < playersCount; i++ {
		playerID, err := GeneratePlayerID()
		require.NoError(t, err)

		state.Players = append(state.Players, protocol.Player{
			ID:   playerID,
			Name: fmt.Sprintf("player-%d", i),
		})

		vote := fmt.Sprintf("%d", i%len(deck))
		votes[playerID] = *protocol.NewVoteResult(protocol.VoteValue(vote))
	}

	for i := 0; i < issuesCount; i++ {
		voteItemID, err := GenerateIssueID()
		require.NoError(t, err)

		state.Issues = append(state.Issues, &protocol.Issue{
			ID:         voteItemID,
			TitleOrURL: fmt.Sprintf("https://github.com/six78/waku-poker-planing/issues/%d", i),
			Votes:      votes, // same votes for each issue, whatever
			Result:     &deck[i%len(deck)],
		})
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

	player := &protocol.Player{
		ID:   "player-1",
		Name: "player-1",
	}
	transport := NewTransportMock(t)
	game := NewGame(ctx, transport, player)

	// FIXME: first subscribe to messages. And after creating room wait for initial state.

	err = game.CreateNewRoom()
	require.NoError(t, err)

	room := game.Room()
	sub, err := transport.SubscribeToMessages(&room)
	require.NoError(t, err)

	const firstItemText = "a"
	const dealerVote = protocol.VoteValue("1")

	var firstVoteItemID protocol.IssueID

	checkVoteItems := func(t *testing.T, issuesList protocol.IssuesList) *protocol.Issue {
		require.Len(t, issuesList, 1)
		item := issuesList.Get(firstVoteItemID)
		require.NotNil(t, item)
		require.Equal(t, firstItemText, item.TitleOrURL)
		return item
	}

	{ // Deal first vote item
		firstVoteItemID, err = game.Deal(firstItemText)
		require.NoError(t, err)

		state := expectState(t, sub.Ch, nil)
		item := checkVoteItems(t, state.Issues)
		require.Nil(t, item.Result)
		require.Len(t, item.Votes, 0)
	}

	currentVoteItem := game.CurrentState().Issues.Get(game.CurrentState().ActiveIssue)
	require.NotNil(t, currentVoteItem)
	require.Equal(t, firstItemText, currentVoteItem.TitleOrURL)

	{ // Publish dealer vote
		err = game.PublishVote(dealerVote)
		require.NoError(t, err)

		playerVote := expectPlayerVote(t, sub.Ch)
		require.Equal(t, game.Player().ID, playerVote.PlayerID)
		require.Equal(t, currentVoteItem.ID, playerVote.Issue)
		require.Equal(t, dealerVote, playerVote.VoteResult.Value)
		require.Greater(t, playerVote.VoteResult.Timestamp, int64(0))

		state := expectState(t, sub.Ch, func(state *protocol.State) bool {
			item := state.Issues.Get(firstVoteItemID)
			if item == nil {
				return false
			}
			_, ok := item.Votes[game.Player().ID]
			return ok
		})
		// FIXME: check vote state
		item := checkVoteItems(t, state.Issues)
		require.Nil(t, item.Result)
		require.Len(t, item.Votes, 1)

		vote, ok := item.Votes[game.Player().ID]
		require.True(t, ok)
		require.Empty(t, vote.Value)
		require.Greater(t, vote.Timestamp, int64(0))
	}

	{ // Reveal votes
		err = game.Reveal()
		require.NoError(t, err)

		state := expectState(t, sub.Ch, nil)
		item := checkVoteItems(t, state.Issues)
		require.Nil(t, item.Result)
		require.Len(t, item.Votes, 1)

		vote, ok := item.Votes[game.Player().ID]
		require.True(t, ok)
		require.NotNil(t, vote)
		require.Equal(t, dealerVote, vote.Value)
		require.Greater(t, vote.Timestamp, int64(0))
	}

	const votingResult = protocol.VoteValue("1")

	{ // Finish voting
		err = game.Finish(votingResult)
		require.NoError(t, err)

		state := expectState(t, sub.Ch, nil)
		item := checkVoteItems(t, state.Issues)
		require.NotNil(t, item.Result)
		require.Equal(t, *item.Result, votingResult)
		require.Len(t, item.Votes, 1)

		vote, ok := item.Votes[game.Player().ID]
		require.True(t, ok)
		require.Equal(t, dealerVote, vote.Value)
		require.Greater(t, vote.Timestamp, int64(0))
	}

	const secondItemText = "b"
	var secondVoteItemID protocol.IssueID

	checkVoteItems = func(t *testing.T, voteList protocol.IssuesList) *protocol.Issue {
		require.Len(t, voteList, 2)

		item := voteList.Get(firstVoteItemID)
		require.NotNil(t, item)
		require.Equal(t, firstItemText, item.TitleOrURL)

		item = voteList.Get(secondVoteItemID)
		require.NotNil(t, item)
		require.Equal(t, secondItemText, item.TitleOrURL)

		return item
	}

	{ // Deal another vote item
		secondVoteItemID, err = game.Deal(secondItemText)
		require.NoError(t, err)

		state := expectState(t, sub.Ch, nil)
		item := checkVoteItems(t, state.Issues)
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
