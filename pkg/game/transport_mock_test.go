package game

import (
	"2sp/internal/transport"
	"2sp/pkg/protocol"
	"github.com/stretchr/testify/require"
	"testing"
)

type TransportMock struct {
	t             *testing.T
	subscriptions map[protocol.RoomID][]chan []byte
}

func NewTransportMock(t *testing.T) *TransportMock {
	return &TransportMock{
		t:             t,
		subscriptions: make(map[protocol.RoomID][]chan []byte),
	}
}

func (t *TransportMock) SubscribeToMessages(room *protocol.Room) (*transport.MessagesSubscription, error) {
	roomID, err := room.ToRoomID()
	require.NoError(t.t, err)

	channel := make(chan []byte, 10)
	subs, ok := t.subscriptions[roomID]
	if !ok {
		subs = make([]chan []byte, 0, 1)
	}
	subs = append(subs, channel)
	t.subscriptions[roomID] = subs
	return &transport.MessagesSubscription{
		Ch:          channel,
		Unsubscribe: nil,
	}, nil
}

func (t *TransportMock) PublishUnencryptedMessage(room *protocol.Room, payload []byte) error {
	roomID, err := room.ToRoomID()
	require.NoError(t.t, err)

	subs, ok := t.subscriptions[roomID]
	if !ok {
		return nil
	}

	for _, sub := range subs {
		sub <- payload
	}

	return nil
}

func (t *TransportMock) PublishPublicMessage(room *protocol.Room, payload []byte) error {
	return t.PublishUnencryptedMessage(room, payload)
}

func (t *TransportMock) PublishPrivateMessage(room *protocol.Room, payload []byte) error {
	return t.PublishUnencryptedMessage(room, payload)
}

func (t *TransportMock) ConnectionStatus() transport.ConnectionStatus {
	return transport.ConnectionStatus{}
}

func (t *TransportMock) SubscribeToConnectionStatus() transport.ConnectionStatusSubscription {
	return nil
}

func (t *TransportMock) subscribeToAll() {
	for _, subs := range t.subscriptions {
		for _, sub := range subs {
			close(sub)
		}
	}
}
