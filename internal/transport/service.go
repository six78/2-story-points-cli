package transport

import "github.com/six78/2-story-points-cli/pkg/protocol"

//go:generate mockgen -source=service.go -destination=mock/service.go

type Service interface {
	Initialize() error
	Start() error

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
