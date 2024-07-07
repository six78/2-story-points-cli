package game

import (
	"strings"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"

	"github.com/six78/2-story-points-cli/pkg/protocol"
)

func TestMedian(t *testing.T) {
	// Odd number of votes
	votes := []int{1, 1, 1, 1, 2}
	hint := median(votes)
	require.Equal(t, 1, hint)

	// Even number of votes
	votes = []int{1, 1, 1, 2}
	hint = median(votes)
	require.Equal(t, 1, hint)

	// Test round up
	votes = []int{1, 1, 2, 2}
	hint = median(votes)
	require.Equal(t, 2, hint)

	// Empty list
	votes = []int{}
	hint = median(votes)
	require.Equal(t, -1, hint)
}

func TestHint(t *testing.T) {
	deck := protocol.Deck{"1", "2", "3", "5", "8", "13", "21"}

	type Case struct {
		values       []protocol.VoteValue
		measurements hintMeasurements
		expectedHint protocol.Hint
	}

	// NOTE: Some test cases here are double check.
	//		 But the intention was also to see how the algorithm behaves for different scenarios.

	testCases := []Case{
		{
			values:       []protocol.VoteValue{"3", "3", "3", "3", "3"},
			measurements: hintMeasurements{median: 2, meanDeviation: 0, maxDeviation: 0},
			expectedHint: protocol.Hint{Acceptable: true, Value: "3", Description: descriptionBingo},
		},
		{
			values:       []protocol.VoteValue{"3", "3", "3", "3", "5"},
			measurements: hintMeasurements{median: 2, meanDeviation: 0.2, maxDeviation: 1},
			expectedHint: protocol.Hint{Acceptable: true, Value: "3", Description: descriptionGoodJob},
		},
		{
			values:       []protocol.VoteValue{"3", "3", "3", "3", "8"},
			measurements: hintMeasurements{median: 2, meanDeviation: 0.4, maxDeviation: 2},
			expectedHint: protocol.Hint{Acceptable: false, Value: "3", Description: maximumDeviationIsTooHigh},
		},
		{
			values:       []protocol.VoteValue{"3", "3", "3", "3", "13"},
			measurements: hintMeasurements{median: 2, meanDeviation: 0.6, maxDeviation: 3},
			// Test: varietyOfVotesIsTooHigh takes precedence over maximumDeviationIsTooHigh
			expectedHint: protocol.Hint{Acceptable: false, Value: "3", Description: varietyOfVotesIsTooHigh},
		},
		{
			values:       []protocol.VoteValue{"3", "3", "3", "3", "21"},
			measurements: hintMeasurements{median: 2, meanDeviation: 0.8, maxDeviation: 4},
			expectedHint: protocol.Hint{Acceptable: false, Value: "3", Description: varietyOfVotesIsTooHigh},
		},
		{
			values:       []protocol.VoteValue{"3", "3", "3", "5", "5"},
			measurements: hintMeasurements{median: 2, meanDeviation: 0.4, maxDeviation: 1},
			expectedHint: protocol.Hint{Acceptable: true, Value: "3", Description: descriptionNotBad},
		},
		{
			values:       []protocol.VoteValue{"3", "3", "3", "5", "8"},
			measurements: hintMeasurements{median: 2, meanDeviation: 0.6, maxDeviation: 2},
			expectedHint: protocol.Hint{Acceptable: false, Value: "3", Description: varietyOfVotesIsTooHigh},
		},
		{
			values:       []protocol.VoteValue{"2", "3", "3", "3", "3", "3", "5"},
			measurements: hintMeasurements{median: 2, meanDeviation: 2 / 7.0, maxDeviation: 1},
			expectedHint: protocol.Hint{Acceptable: true, Value: "3", Description: descriptionNotBad},
		},
		{
			values:       []protocol.VoteValue{"2", "3", "3", "3", "3", "5"},
			measurements: hintMeasurements{median: 2, meanDeviation: 2 / 6.0, maxDeviation: 1},
			expectedHint: protocol.Hint{Acceptable: true, Value: "3", Description: descriptionNotBad},
		},
		{
			values:       []protocol.VoteValue{"2", "3", "3", "3", "5"},
			measurements: hintMeasurements{median: 2, meanDeviation: 2 / 5.0, maxDeviation: 1},
			expectedHint: protocol.Hint{Acceptable: true, Value: "3", Description: descriptionNotBad},
		},
		{
			values:       []protocol.VoteValue{"2", "3", "3", "5"},
			measurements: hintMeasurements{median: 2, meanDeviation: 2 / 4.0, maxDeviation: 1},
			expectedHint: protocol.Hint{Acceptable: true, Value: "3", Description: descriptionYouCanDoBetter},
		},
		{
			values:       []protocol.VoteValue{"2", "3", "5"},
			measurements: hintMeasurements{median: 2, meanDeviation: 2 / 3.0, maxDeviation: 1},
			expectedHint: protocol.Hint{Acceptable: false, Value: "3", Description: varietyOfVotesIsTooHigh},
		},
		{
			// This also tests round up median when even number of votes
			values:       []protocol.VoteValue{"2", "3", "5", "8"},
			measurements: hintMeasurements{median: 3, meanDeviation: 1, maxDeviation: 2},
			expectedHint: protocol.Hint{Acceptable: false, Value: "5", Description: varietyOfVotesIsTooHigh},
		},
		{
			// This also tests round up median when even number of votes
			values:       []protocol.VoteValue{},
			measurements: hintMeasurements{median: -1, meanDeviation: 0, maxDeviation: 0},
			expectedHint: protocol.Hint{Acceptable: false, Value: "", Description: notEnoughVotes},
		},
	}

	for _, tc := range testCases {
		name := voteValuesString(tc.values)
		t.Run(name, func(t *testing.T) {
			issueVotes := buildIssueVotes(tc.values)
			indexes, err := getVotesAsDeckIndexes(issueVotes, deck)
			require.NoError(t, err)

			// First, check the measures (private API)
			measures := getMeasures(indexes)
			require.Equal(t, tc.measurements, measures)

			// Now check the actual hint (public API)
			hint, err := GetResultHint(deck, issueVotes)
			require.NoError(t, err)
			require.Equal(t, tc.expectedHint, *hint)
		})
	}
}

func TestInvalidVote(t *testing.T) {
	deck := protocol.Deck{"1", "2"}
	issueVotes := buildIssueVotes([]protocol.VoteValue{"1", "X"})
	_, err := GetResultHint(deck, issueVotes)
	require.Error(t, err)
	require.Equal(t, ErrVoteNotFoundInDeck, err)
}

func voteValuesString(values []protocol.VoteValue) string {
	list := make([]string, len(values))
	for i, v := range values {
		list[i] = string(v)
	}
	return strings.Join(list, ",")
}

func buildIssueVotes(votes []protocol.VoteValue) protocol.IssueVotes {
	issueVotes := make(protocol.IssueVotes)
	for _, v := range votes {
		playerID := protocol.PlayerID(gofakeit.UUID())
		issueVotes[playerID] = protocol.VoteResult{Value: v}
	}
	return issueVotes
}
