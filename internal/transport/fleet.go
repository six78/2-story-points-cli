package transport

import (
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
)

type FleetName string

const (
	ShardsStaging FleetName = "shards.staging"
	ShardsTest    FleetName = "shards.test"
	WakuSandbox   FleetName = "waku.sandbox"
	WakuTest      FleetName = "waku.test"
)

const (
	DefaultClusterID = 16
	DefaultShardID   = 64
)

var fleets = map[FleetName]string{
	WakuSandbox: "enrtree://AIRVQ5DDA4FFWLRBCHJWUWOO6X6S4ZTZ5B667LQ6AJU6PEYDLRD5O@sandbox.waku.nodes.status.im",
	ShardsTest:  "enrtree://AMOJVZX4V6EXP7NTJPMAYJYST2QP6AJXYW76IU6VGJS7UVSNDYZG4@boot.test.shards.nodes.status.im",
}

func FleetENRTree(fleet FleetName) (string, bool) {
	enr, ok := fleets[fleet]
	return enr, ok
}

func (f FleetName) IsSharded() bool {
	return f == ShardsStaging || f == ShardsTest
}

func (f FleetName) DefaultPubsubTopic() string {
	if f.IsSharded() {
		return protocol.NewStaticShardingPubsubTopic(DefaultClusterID, DefaultShardID).String()
	}

	return relay.DefaultWakuTopic
}
