package protocol

import (
	"golang.org/x/exp/maps"

	"github.com/six78/2-story-points-cli/internal/config"
)

type State struct {
	Players       PlayersList `json:"players"`
	Issues        IssuesList  `json:"issues"`
	ActiveIssue   IssueID     `json:"activeIssue"`
	VotesRevealed bool        `json:"votesRevealed"`
	Timestamp     int64       `json:"-"`    // TODO: Fix conflict with Message.Timestamp. Change type to time.Time.
	Deck          Deck        `json:"deck"` // NOTE: This field is experimental and not supported by web client
}

type VoteState string

const (
	IdleState     VoteState = "idle"     // ActiveIssue == ""
	VotingState   VoteState = "voting"   // ActiveIssue != "", Revealed == false
	RevealedState VoteState = "revealed" // ActiveIssue != "", Revealed == true, Issues[ActiveIssue].Result == nil
	FinishedState VoteState = "finished" // ActiveIssue != "", Revealed == true, Issues[ActiveIssue].Result != nil
)

func (s *State) VoteState() VoteState {
	if s.ActiveIssue == "" {
		return IdleState
	}
	if !s.VotesRevealed {
		return VotingState
	}
	issue := s.Issues.Get(s.ActiveIssue)
	if issue == nil {
		config.Logger.Error("active issue not found when calculating vote state")
		return IdleState
	}
	if issue.Result == nil {
		return RevealedState
	}
	return FinishedState
}

func (s *State) GetActiveIssue() *Issue {
	return s.Issues.Get(s.ActiveIssue)
}

func (s *State) ActiveIssueHintDeckIndex() int {
	issue := s.GetActiveIssue()
	if issue == nil {
		return -1
	}
	if issue.Hint == nil {
		return -1
	}
	return s.Deck.Index(issue.Hint.Value)
}

func (s *State) AllPlayersVoted() bool {
	issue := s.GetActiveIssue()
	return issue != nil && len(maps.Keys(issue.Votes)) == len(s.Players)
}
