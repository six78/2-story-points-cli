package storage

import (
	"2sp/pkg/protocol"
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
