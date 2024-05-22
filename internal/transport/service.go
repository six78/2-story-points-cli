package transport

import "2sp/pkg/protocol"

type Service interface {
	SubscribeToMessages(room *protocol.Room) (*MessagesSubscription, error)
	PublishUnencryptedMessage(room *protocol.Room, payload []byte) error
	PublishPublicMessage(room *protocol.Room, payload []byte) error
	PublishPrivateMessage(room *protocol.Room, payload []byte) error

	ConnectionStatus() ConnectionStatus
	SubscribeToConnectionStatus() ConnectionStatusSubscription
}

type MessagesSubscription struct {
	Ch          chan []byte
	Unsubscribe func()
}

type ConnectionStatus struct {
	IsOnline   bool
	HasHistory bool
	PeersCount int
}

type ConnectionStatusSubscription chan ConnectionStatus
