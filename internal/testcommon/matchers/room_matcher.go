package matchers

import (
	"fmt"

	"github.com/six78/2-story-points-cli/pkg/protocol"
)

type RoomMatcher struct {
	room *protocol.Room
}

func NewRoomMatcher(room *protocol.Room) *RoomMatcher {
	return &RoomMatcher{
		room: room,
	}
}

func (m *RoomMatcher) Matches(x interface{}) bool {
	switch room := x.(type) {
	case *protocol.Room:
		return m.room.ToRoomID() == room.ToRoomID()
	case protocol.Room:
		return m.room.ToRoomID() == room.ToRoomID()
	case protocol.RoomID:
		return m.room.ToRoomID() == room
	}
	return false
}

func (m *RoomMatcher) String() string {
	return fmt.Sprintf("is equal to room %s", m.room.ToRoomID())
}
