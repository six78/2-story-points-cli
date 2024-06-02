package game

import (
	"2sp/internal/testcommon"
	"2sp/internal/testcommon/matchers"
	"2sp/internal/transport"
	mocktransport "2sp/internal/transport/mock"
	"2sp/pkg/protocol"
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestGame(t *testing.T) {
	suite.Run(t, new(Suite))
}

type Suite struct {
	testcommon.Suite

	ctx       context.Context
	cancel    context.CancelFunc
	transport *mocktransport.MockService

	game     *Game
	stateSub StateSubscription
}

func (s *Suite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	ctrl := gomock.NewController(s.T())
	s.transport = mocktransport.NewMockService(ctrl)

	s.game = s.newGame([]Option{
		WithEnableSymmetricEncryption(true),
	})

	s.stateSub = s.game.SubscribeToStateChanges()
	s.Require().NotNil(s.stateSub)

	err := s.game.Initialize()
	s.Require().NoError(err)
}

func (s *Suite) TearDownTest() {
	s.cancel()
}

func (s *Suite) TestStateSize() {
	const playersCount = 20
	const issuesCount = 30

	state := protocol.State{
		Players: make([]protocol.Player, 0, playersCount),
		Issues:  make(protocol.IssuesList, 0, issuesCount),
	}

	votes := make(map[protocol.PlayerID]protocol.VoteResult, playersCount)
	deck, deckFound := GetDeck(Fibonacci)
	s.Require().True(deckFound)

	state.Deck = deck

	for i := 0; i < playersCount; i++ {
		playerID, err := GeneratePlayerID()
		s.Require().NoError(err)

		state.Players = append(state.Players, protocol.Player{
			ID:   playerID,
			Name: fmt.Sprintf("player-%d", i),
		})

		vote := fmt.Sprintf("%d", i%len(deck))
		votes[playerID] = *protocol.NewVoteResult(protocol.VoteValue(vote))
	}

	for i := 0; i < issuesCount; i++ {
		issueID, err := GenerateIssueID()
		s.Require().NoError(err)

		state.Issues = append(state.Issues, &protocol.Issue{
			ID:         issueID,
			TitleOrURL: fmt.Sprintf("https://github.com/six78/waku-poker-planing/issues/%d", i),
			Votes:      votes, // same votes for each issue, whatever
			Result:     &deck[i%len(deck)],
		})
	}

	stateMessage, err := json.Marshal(state)
	s.Require().NoError(err)

	fmt.Println("state size", "bytes", len(stateMessage))
	s.Require().True(len(stateMessage) < 100*1024, "state size should be less than 100 kilobytes")
}

func (s *Suite) TestSimpleGame() {
	room, initialState, err := s.game.CreateNewRoom()
	s.Require().NoError(err)
	s.Require().NotNil(room)

	roomID := room.ToRoomID()

	roomMatcher := matchers.NewRoomMatcher(room)
	onlineMatcher := matchers.NewOnlineMatcher()

	// Online state is sent periodically
	s.transport.EXPECT().PublishPublicMessage(roomMatcher, onlineMatcher).AnyTimes()

	// We need to loop the subscription to mock waku behaviour
	// We should probably check the published messages instead of received ones, but it's fine for now.
	subscription := &transport.MessagesSubscription{
		Ch:          make(chan []byte),
		Unsubscribe: func() {},
	}
	loop := func(room *protocol.Room, payload []byte) {
		subscription.Ch <- payload
	}
	s.transport.EXPECT().SubscribeToMessages(roomMatcher).
		Return(subscription, nil).
		Times(1)

	// Join room
	stateMatcher := matchers.NewStateMatcher(nil)
	s.transport.EXPECT().PublishPublicMessage(roomMatcher, stateMatcher).
		Times(1)

	err = s.game.JoinRoom(roomID, initialState)
	s.Require().NoError(err)

	state := stateMatcher.Wait(s.T())
	s.Require().False(state.VotesRevealed)
	s.Require().Empty(state.ActiveIssue)
	s.Require().Len(state.Players, 1)
	s.Logger.Info("match on join room")

	// Deal first vote item

	firstItemText := gofakeit.LetterN(10)
	const dealerVote = protocol.VoteValue("1")

	var firstIssueID protocol.IssueID

	checkIssues := func(issuesList protocol.IssuesList) *protocol.Issue {
		s.Require().Len(issuesList, 1)
		item := issuesList.Get(firstIssueID)
		s.Require().NotNil(item)
		s.Require().Equal(firstItemText, item.TitleOrURL)
		return item
	}

	{ // Deal first vote item
		stateMatcher = matchers.NewStateMatcher(nil)

		s.transport.EXPECT().
			PublishPublicMessage(roomMatcher, stateMatcher).
			Times(1)

		firstIssueID, err = s.game.Deal(firstItemText)
		s.Require().NoError(err)

		state = stateMatcher.Wait(s.T())
		item := checkIssues(state.Issues)
		s.Require().Nil(item.Result)
		s.Require().Len(item.Votes, 0)
		s.Logger.Info("match on deal first item")
	}

	currentIssue := s.game.CurrentState().Issues.Get(s.game.CurrentState().ActiveIssue)
	s.Require().NotNil(currentIssue)
	s.Require().Equal(firstItemText, currentIssue.TitleOrURL)

	{ // Publish dealer vote
		voteMatcher := matchers.NewVoteMatcher(s.game.Player().ID, currentIssue.ID, dealerVote)
		s.transport.EXPECT().
			PublishPublicMessage(roomMatcher, voteMatcher).
			Do(loop). // FIXME: Game should not depend on transport loop/no-loop behaviour
			Times(1)

		stateMatcher = matchers.NewStateMatcher(nil)
		s.transport.EXPECT().
			PublishPublicMessage(roomMatcher, stateMatcher).
			Times(1)

		err = s.game.PublishVote(dealerVote)
		s.Require().NoError(err)

		state = stateMatcher.Wait(s.T())
		item := checkIssues(state.Issues)
		s.Require().NotNil(item)
		s.Require().Nil(item.Result)
		s.Require().Len(item.Votes, 1)

		vote, ok := item.Votes[s.game.Player().ID]
		s.Require().True(ok)
		s.Require().Empty(vote.Value)
		s.Require().Greater(vote.Timestamp, int64(0))
	}

	{ // Reveal votes
		stateMatcher = matchers.NewStateMatcher(nil)
		s.transport.EXPECT().
			PublishPublicMessage(roomMatcher, stateMatcher).
			Times(1)

		err = s.game.Reveal()
		s.Require().NoError(err)

		state = stateMatcher.Wait(s.T())
		item := checkIssues(state.Issues)
		s.Require().Nil(item.Result)
		s.Require().Len(item.Votes, 1)

		vote, ok := item.Votes[s.game.Player().ID]
		s.Require().True(ok)
		s.Require().NotNil(vote)
		s.Require().Equal(dealerVote, vote.Value)
		s.Require().Greater(vote.Timestamp, int64(0))
	}

	const votingResult = protocol.VoteValue("1")

	{ // Finish voting
		stateMatcher = matchers.NewStateMatcher(nil)
		s.transport.EXPECT().
			PublishPublicMessage(roomMatcher, stateMatcher).
			Times(1)

		err = s.game.Finish(votingResult)
		s.Require().NoError(err)

		state = stateMatcher.Wait(s.T())
		item := checkIssues(state.Issues)
		s.Require().NotNil(item.Result)
		s.Require().Equal(*item.Result, votingResult)
		s.Require().Len(item.Votes, 1)

		vote, ok := item.Votes[s.game.Player().ID]
		s.Require().True(ok)
		s.Require().Equal(dealerVote, vote.Value)
		s.Require().Greater(vote.Timestamp, int64(0))
	}

	const secondItemText = "b"
	var secondIssueID protocol.IssueID

	checkIssues = func(issues protocol.IssuesList) *protocol.Issue {
		s.Require().Len(issues, 2)

		item := issues.Get(firstIssueID)
		s.Require().NotNil(item)
		s.Require().Equal(firstItemText, item.TitleOrURL)

		item = issues.Get(secondIssueID)
		s.Require().NotNil(item)
		s.Require().Equal(secondItemText, item.TitleOrURL)

		return item
	}

	{ // Deal another issue
		stateMatcher = matchers.NewStateMatcher(nil)
		s.transport.EXPECT().
			PublishPublicMessage(roomMatcher, stateMatcher).
			Times(1)

		secondIssueID, err = s.game.Deal(secondItemText)
		s.Require().NoError(err)

		state = stateMatcher.Wait(s.T())
		item := checkIssues(state.Issues)
		s.Require().Nil(item.Result)
		s.Require().Len(item.Votes, 0)
	}
}
