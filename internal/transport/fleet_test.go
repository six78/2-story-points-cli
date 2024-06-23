package transport

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/require"
)

func TestFleets(t *testing.T) {
	enr, ok := FleetENRTree(WakuSandbox)
	require.True(t, ok)
	require.Equal(t, fleets[WakuSandbox], enr)

	enr, ok = FleetENRTree(ShardsTest)
	require.True(t, ok)
	require.Equal(t, fleets[ShardsTest], enr)

	_, ok = FleetENRTree(WakuTest)
	require.False(t, ok) // We know this fleet, but it's not supported
}

func TestFleetSharded(t *testing.T) {
	require.True(t, ShardsTest.IsSharded())
	require.True(t, ShardsStaging.IsSharded())
	require.False(t, WakuSandbox.IsSharded())
	require.False(t, WakuTest.IsSharded())
	require.False(t, FleetName(gofakeit.LetterN(5)).IsSharded())
}

func TestFleetDefaultPubsubTopic(t *testing.T) {
	require.Equal(t, "/waku/2/default-waku/proto", WakuSandbox.DefaultPubsubTopic())
	require.Equal(t, "/waku/2/rs/16/64", ShardsTest.DefaultPubsubTopic())
}
