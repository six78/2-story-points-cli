package protocol

import (
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
)

type Session struct {
	Version      byte   `json:"version"`
	SymmetricKey []byte `json:"symmetricKey"`
}

type SessionID struct {
	string
}

func NewSessionID(sessionID string) SessionID {
	return SessionID{sessionID}
}

func (id SessionID) String() string {
	return id.string
}

// SessionID: base58 encoded byte array:
// - byte 0: 	    version
// - byte 1..end: symmetric key
// Total expected length: 17 bytes

func (info *Session) ToSessionID() (SessionID, error) {
	bytes := make([]byte, 0, 1+len(info.SymmetricKey))
	bytes = append(bytes, Version)
	bytes = append(bytes, info.SymmetricKey...)
	sessionID := base58.Encode(bytes)
	return NewSessionID(sessionID), nil
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
