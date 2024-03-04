package protocol

import (
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
)

const Version int = 1

//
//type Player struct {
//	Name   string
//	Dealer bool
//}

type Player string
type PlayerID string
type VoteResult int

type VoteItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

type State struct {
	Players        []Player              `json:"players"`
	VoteItem       VoteItem              `json:"voteItem"`
	TempVoteResult map[Player]VoteResult `json:"tempVoteResults"`
}

type MessageType string

const (
	MessageTypeState        MessageType = "__state"
	MessageTypePlayerOnline             = "__player_online"
	MessageTypePlayerVote               = "__player_vote"
)

type Message struct {
	Type      MessageType `json:"type"`
	Timestamp int64       `json:"timestamp"`
}

type GameStateMessage struct {
	Message
	State State `json:"state"`
}

type PlayerOnlineMessage struct {
	Message
	Name Player `json:"name,omitempty"`
}

type PlayerVote struct {
	Message
	VoteBy     Player     `json:"voteBy"`
	VoteFor    string     `json:"voteFor"`
	VoteResult VoteResult `json:"voteResult"`
}

type Session struct {
	SymmetricKey []byte `json:"SymmetricKey"`
}

func (info *Session) ToSessionID() (string, error) {
	sessionID := base58.Encode(info.SymmetricKey)
	return sessionID, nil
}

func ParseSessionID(sessionID string) (*Session, error) {
	decoded, err := base58.Decode(sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode session ID")
	}

	return &Session{
		SymmetricKey: decoded,
	}, nil
}

func BuildSession(symmetricKey []byte) *Session {
	return &Session{
		SymmetricKey: symmetricKey,
	}
}
