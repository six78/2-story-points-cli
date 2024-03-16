package game

import "waku-poker-planning/protocol"

type Transport interface {
	SubscribeToMessages(session *protocol.Session) (chan []byte, error)
	PublishUnencryptedMessage(session *protocol.Session, payload []byte) error
}
