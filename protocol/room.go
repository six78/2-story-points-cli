package protocol

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
	"github.com/pkg/errors"
	"waku-poker-planning/config"
)

type Room struct {
	Version          byte              `json:"version"`
	SymmetricKey     []byte            `json:"symmetricKey"`
	DealerPublicKey  []byte            `json:"dealerPublicKey"`
	DealerPrivateKey *ecdsa.PrivateKey `json:"dealerPrivateKey"`
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

func (room *Room) Bytes() ([]byte, error) {
	dealerPublicKey, err := x509.MarshalPKIXPublicKey(room.DealerPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal dealer public key")
	}

	bytes := make([]byte, 0, 1+len(room.SymmetricKey)+len(dealerPublicKey))
	bytes = append(bytes, room.Version)
	bytes = append(bytes, room.SymmetricKey...)
	bytes = append(bytes, dealerPublicKey...)
	return bytes, nil
}

func (room *Room) ToRoomID() (RoomID, error) {
	bytes, err := room.Bytes()
	if err != nil {
		return RoomID{}, errors.Wrap(err, "failed to convert room to bytes")
	}

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

func NewRoom() (*Room, error) {
	symmetricKey, err := generateSymmetricKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate symmetric key")
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate dealer private key")
	}

	return &Room{
		Version:          Version,
		SymmetricKey:     symmetricKey,
		DealerPublicKey:  crypto.CompressPubkey(&privateKey.PublicKey),
		DealerPrivateKey: privateKey,
	}, nil
}

func generateSymmetricKey() ([]byte, error) {
	key := make([]byte, config.SymmetricKeyLength)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
