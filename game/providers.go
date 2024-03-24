package game

import "waku-poker-planning/protocol"

type Transport interface {
	SubscribeToMessages(room *protocol.Room) (chan []byte, error)
	PublishUnencryptedMessage(room *protocol.Room, payload []byte) error
	PublishPublicMessage(room *protocol.Room, payload []byte) error
	PublishPrivateMessage(room *protocol.Room, payload []byte) error
}
