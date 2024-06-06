package storage

//go:generate mockgen -source=service.go -destination=mock/service.go

import (
	"github.com/six78/2-story-points-cli/pkg/protocol"
)

type Service interface {
	Initialize() error
	PlayerID() protocol.PlayerID
	PlayerName() string
	SetPlayerID(id protocol.PlayerID) error
	SetPlayerName(name string) error
	LoadRoomState(roomID protocol.RoomID) (*protocol.State, error)
	SaveRoomState(roomID protocol.RoomID, state *protocol.State) error
}
