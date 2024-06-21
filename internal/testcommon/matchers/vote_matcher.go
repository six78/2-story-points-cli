package matchers

import (
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

type VoteMatcher struct {
	MessageMatcher
	playerID  protocol.PlayerID
	issueID   protocol.IssueID
	voteValue protocol.VoteValue
}

func NewVoteMatcher(playerID protocol.PlayerID, issue protocol.IssueID, value protocol.VoteValue) *VoteMatcher {
	return &VoteMatcher{
		playerID:  playerID,
		issueID:   issue,
		voteValue: value,
	}
}

func (m *VoteMatcher) Matches(x interface{}) bool {
	if !m.MessageMatcher.Matches(x) {
		return false
	}

	if m.message.Type != protocol.MessageTypePlayerVote {
		return false
	}

	vote, err := protocol.UnmarshalPlayerVote(m.payload)
	if err != nil {
		return false
	}

	return vote.PlayerID == m.playerID &&
		vote.Issue == m.issueID &&
		vote.VoteResult.Value == m.voteValue &&
		vote.Timestamp > 0
}

func (m *VoteMatcher) String() string {
	return "is any state message"
}
