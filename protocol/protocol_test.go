package protocol

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestSessionID(t *testing.T) {
	key := []byte{1, 2, 3, 4}
	sent := BuildSession(key)

	sessionID, err := sent.ToSessionID()
	require.NoError(t, err)
	require.NotEmpty(t, sessionID)

	received, err := ParseSessionID(sessionID)
	require.NoError(t, err)
	require.NotEmpty(t, received)

	require.True(t, reflect.DeepEqual(sent, received))

	require.Equal(t, sent.Version, received.Version)
	require.Equal(t, sent.SymmetricKey, received.SymmetricKey)
}
