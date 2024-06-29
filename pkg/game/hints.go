package game

import (
	"math"
	"sort"

	"github.com/pkg/errors"
	"golang.org/x/exp/slices"

	"github.com/six78/2-story-points-cli/pkg/protocol"
)

type hintMeasurements struct {
	// median is the index median value of given votes
	median int

	// meanDeviation is the mean absolute deviation around the median
	// Measured in cards count.
	meanDeviation float64

	// maxDeviation is the maximum absolute deviation from the median
	// Measured in cards count.
	maxDeviation float64
}

const (
	maxAcceptableMaximumDeviation = 1
	maxAcceptableMeanDeviation    = 0.5
	// Rejection reasons
	varietyOfVotesIsTooHigh   = "Variety of votes is too high"
	maximumDeviationIsTooHigh = "Maximum deviation is too high"
)

var (
	ErrVoteNotFoundInDeck = errors.New("vote not found in deck")
)

func GetResultHint(deck protocol.Deck, issueVotes protocol.IssueVotes) (*protocol.Hint, error) {
	// Get votes as deck indexes.
	// We ignore the actual deck values when calculating the hint.
	indexes, err := getVotesAsDeckIndexes(issueVotes, deck)
	if err != nil {
		return nil, err
	}

	// Calculate measures for the votes
	resultMeasures := getMeasures(indexes)
	medianValueIndex := resultMeasures.median
	medianValue := deck[medianValueIndex]

	// Build the hint based on the measures
	hint := &protocol.Hint{
		Hint:       medianValue,
		Advice:     "",
		Acceptable: true,
	}

	if resultMeasures.maxDeviation > maxAcceptableMaximumDeviation {
		hint.Acceptable = false
		hint.RejectReason = maximumDeviationIsTooHigh
	}

	if resultMeasures.meanDeviation >= maxAcceptableMeanDeviation {
		hint.Acceptable = false
		hint.RejectReason = varietyOfVotesIsTooHigh
	}

	return hint, nil
}

func getVotesAsDeckIndexes(issueVotes protocol.IssueVotes, deck protocol.Deck) ([]int, error) {
	indexes := make([]int, 0, len(issueVotes))
	for _, vote := range issueVotes {
		index := slices.Index(deck, vote.Value)
		if index < 0 {
			return nil, ErrVoteNotFoundInDeck
		}
		indexes = append(indexes, index)
	}
	return indexes, nil
}

// getMeasures returns:
// - median value
// - median absolute deviation
// - maximum absolute deviation
// - error if any occurred
func getMeasures(values []int) hintMeasurements {
	r := hintMeasurements{}

	// median value
	r.median = median(values)

	// Maximum deviation
	r.maxDeviation = 0
	for _, v := range values {
		deviation := math.Abs(float64(r.median) - float64(v))
		r.maxDeviation = math.Max(r.maxDeviation, deviation)
	}

	// Average deviation
	sum := 0
	for _, v := range values {
		sum += int(math.Abs(float64(r.median) - float64(v)))
	}
	r.meanDeviation = float64(sum) / float64(len(values))

	return r

}

func median(values []int) int {
	if len(values) == 0 {
		return -1
	}

	sort.Slice(values, func(i, j int) bool {
		return values[i] < values[j]
	})
	center := len(values) / 2
	return values[center]
}
