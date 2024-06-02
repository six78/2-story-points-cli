package protocol

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestRoomID(t *testing.T) {
	sent, err := NewRoom()
	require.NoError(t, err)

	roomID := sent.ToRoomID()
	require.NotEmpty(t, roomID)

	received, err := ParseRoomID(roomID.String())
	require.NoError(t, err)
	require.NotEmpty(t, received)

	require.True(t, reflect.DeepEqual(sent, received))

	require.Equal(t, sent.Version, received.Version)
	require.Equal(t, sent.SymmetricKey, received.SymmetricKey)
}
