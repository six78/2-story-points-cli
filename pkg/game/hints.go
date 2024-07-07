package game

import (
	"math"
	"sort"

	"github.com/pkg/errors"

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
	// thresholds
	maxAcceptableMaximumDeviation = 1
	maxAcceptableMeanDeviation    = 0.5
	// Rejection reasons
	varietyOfVotesIsTooHigh   = "No strong consensus among the players"
	maximumDeviationIsTooHigh = "Maximum deviation threshold exceeded"
	notEnoughVotes            = "Nobody voted 🤷‍"
	// Advices
	descriptionBingo          = "BINGO! 🎉💃"
	descriptionGoodJob        = "Good job 😎"
	descriptionNotBad         = "Not bad 🤞"
	descriptionYouCanDoBetter = "You can do better 💪"
	// internal consts
	float64Epsilon = 1e-9
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

	if len(indexes) == 0 {
		return &protocol.Hint{
			Acceptable:  false,
			Value:       "",
			Description: notEnoughVotes,
		}, nil
	}

	// Calculate measures for the votes
	resultMeasures := getMeasures(indexes)
	medianValueIndex := resultMeasures.median
	medianValue := deck[medianValueIndex]

	// Build the hint based on the measures
	hint := &protocol.Hint{
		Value:      medianValue,
		Acceptable: true,
	}

	if resultMeasures.maxDeviation > maxAcceptableMaximumDeviation {
		hint.Acceptable = false
		hint.Description = maximumDeviationIsTooHigh
	}

	if resultMeasures.meanDeviation > maxAcceptableMeanDeviation {
		hint.Acceptable = false
		hint.Description = varietyOfVotesIsTooHigh
	}

	if hint.Acceptable {
		switch {
		case resultMeasures.meanDeviation == 0:
			hint.Description = descriptionBingo
		case resultMeasures.meanDeviation < maxAcceptableMeanDeviation/2:
			hint.Description = descriptionGoodJob
		case resultMeasures.meanDeviation < maxAcceptableMeanDeviation:
			hint.Description = descriptionNotBad
		case compareFloats(resultMeasures.meanDeviation, maxAcceptableMeanDeviation):
			hint.Description = descriptionYouCanDoBetter
		}
	}

	return hint, nil
}

func getVotesAsDeckIndexes(issueVotes protocol.IssueVotes, deck protocol.Deck) ([]int, error) {
	indexes := make([]int, 0, len(issueVotes))
	for _, vote := range issueVotes {
		if vote.Value == protocol.UncertaintyCard {
			continue
		}
		index := deck.Index(vote.Value)
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

	if r.median < 0 {
		r.maxDeviation = 0
		r.meanDeviation = 0
		return r
	}

	// Maximum deviation
	r.maxDeviation = 0
	for _, v := range values {
		r.maxDeviation = math.Max(r.maxDeviation, deviation(v, r.median))
	}

	// Average deviation
	sum := 0
	for _, v := range values {
		sum += int(deviation(v, r.median))
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

func deviation(value int, median int) float64 {
	return math.Abs(float64(median) - float64(value))
}

func compareFloats(a, b float64) bool {
	return math.Abs(a-b) < float64Epsilon
}
