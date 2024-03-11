package protocol

import (
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
)

const Version byte = 1

type PlayerID string
type VoteResult int

type Player struct {
	ID       PlayerID
	Name     string
	IsDealer bool
	Order    int
}

type VoteItem struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	URL         string                  `json:"url"`
	Result      VoteResult              `json:"result"`
	VoteHistory map[PlayerID]VoteResult `json:"voteHistory"`
}

type VoteState string

const (
	IdleState     VoteState = "idle"
	VotingState             = "voting"
	RevealedState           = "revealed"
	FinishedState           = "finished"
)

type State struct {
	Players        map[PlayerID]Player      `json:"players"`
	VoteItem       VoteItem                 `json:"voteItem"`
	TempVoteResult map[PlayerID]*VoteResult `json:"tempVoteResults"`
	VoteState      VoteState                `json:"voteState"`
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
	Player Player `json:"name,omitempty"`
}

type PlayerVote struct {
	Message
	VoteBy     PlayerID   `json:"voteBy"`
	VoteFor    string     `json:"voteFor"`
	VoteResult VoteResult `json:"voteResult"`
}

type Session struct {
	Version      byte   `json:"version"`
	SymmetricKey []byte `json:"symmetricKey"`
}

// SessionID: base58 encoded byte array:
// - byte 0: 	    version
// - byte 1..end: symmetric key
// Total expected length: 17 bytes

func (info *Session) ToSessionID() (string, error) {
	bytes := make([]byte, 0, 1+len(info.SymmetricKey))
	bytes = append(bytes, Version)
	bytes = append(bytes, info.SymmetricKey...)
	sessionID := base58.Encode(bytes)
	return sessionID, nil
}

func ParseSessionID(sessionID string) (*Session, error) {
	decoded, err := base58.Decode(sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode session ID")
	}

	if len(decoded) < 1 {
		return nil, errors.New("session id is too short")
	}

	decodedVersion := decoded[0]

	if decodedVersion != Version {
		return nil, errors.Errorf("unexpected version: %d", decodedVersion)
	}

	return &Session{
		Version:      decodedVersion,
		SymmetricKey: decoded[1:],
	}, nil
}

func BuildSession(symmetricKey []byte) *Session {
	return &Session{
		SymmetricKey: symmetricKey,
	}
}
