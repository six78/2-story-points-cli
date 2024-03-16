package protocol

import (
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
)

type Room struct {
	Version      byte   `json:"version"`
	SymmetricKey []byte `json:"symmetricKey"`
}

type RoomID struct {
	string
}

func NewRoomID(roomID string) RoomID {
	return RoomID{roomID}
}

func (id RoomID) String() string {
	return id.string
}

// RoomID: base58 encoded byte array:
// - byte 0: 	    version
// - byte 1..end: symmetric key
// Total expected length: 17 bytes

func (info *Room) ToRoomID() (RoomID, error) {
	bytes := make([]byte, 0, 1+len(info.SymmetricKey))
	bytes = append(bytes, Version)
	bytes = append(bytes, info.SymmetricKey...)
	roomID := base58.Encode(bytes)
	return NewRoomID(roomID), nil
}

func ParseRoomID(roomID string) (*Room, error) {
	decoded, err := base58.Decode(roomID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode room ID")
	}

	if len(decoded) < 1 {
		return nil, errors.New("room id is too short")
	}

	decodedVersion := decoded[0]

	if decodedVersion != Version {
		return nil, errors.Errorf("unexpected version: %d", decodedVersion)
	}

	return &Room{
		Version:      decodedVersion,
		SymmetricKey: decoded[1:],
	}, nil
}

func BuildRoom(symmetricKey []byte) *Room {
	return &Room{
		SymmetricKey: symmetricKey,
	}
}
