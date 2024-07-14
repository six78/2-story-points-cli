package game

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func TestEventManager(t *testing.T) {
	const subscribersCount = 3
	manager := NewEventManager()

	subs := make([]Subscription, subscribersCount)
	for i := range subs {
		subs[i] = *manager.Subscribe()
	}

	require.Equal(t, subscribersCount, manager.Count())

	event := Event{}
	err := gofakeit.Struct(&event)
	require.NoError(t, err)

	manager.Send(event)

	for i := range subs {
		require.Equal(t, event, <-subs[i].Events)
	}

	manager.Close()
	require.Empty(t, manager.subscriptions)

	for i := range subs {
		_, ok := <-subs[i].Events
		require.False(t, ok)
	}

	require.Zero(t, manager.Count())
}
