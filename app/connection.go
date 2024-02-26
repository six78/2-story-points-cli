package app

import (
	"github.com/pkg/errors"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"net"
	"waku-poker-planning/config"
)

var fleets = map[string]string{
	"wakuv2.prod": "enrtree://ANEDLO25QVUGJOUTQFRYKWX6P4Z4GKVESBMHML7DZ6YK4LGS5FC5O@prod.wakuv2.nodes.status.im",
	"wakuv2.test": "enrtree://AO47IDOLBKH72HIZZOXQP6NMRESAN7CHYWIBNXDXWRJRZWLODKII6@test.wakuv2.nodes.status.im",
}

func createWakuNode() (*node.WakuNode, chan node.ConnStatus, error) {
	hostAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to resolve TCP address")
	}

	wakuConnectionStatus := make(chan node.ConnStatus)

	options := []node.WakuNodeOption{
		node.WithWakuRelay(),
		node.WithLightPush(),
		//node.WithLogger(logger),
		//node.WithLogLevel(zap.DebugLevel),
		node.WithHostAddress(hostAddr),
		//node.WithDiscoveryV5(60000, nodes, true),
		node.WithConnectionStatusChannel(wakuConnectionStatus),
	}

	options = append(options, node.DefaultWakuNodeOptions...)

	wakuNode, err := node.New(options...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create waku node")
	}

	return wakuNode, wakuConnectionStatus, nil
}

func (a *App) watchConnectionStatus() {
	for connStatus, more := <-a.wakuConnectionStatus; more; {
		peersCount := len(maps.Keys(connStatus.Peers))
		a.logger.Debug("connection status", zap.Any("peersCount", peersCount))
	}
}

func (a *App) discoverNodes() error {
	enrTree, ok := fleets[config.Fleet]
	if !ok {
		return errors.Errorf("unknown fleet %s", config.Fleet)
	}

	discoveredNodes, err := dnsdisc.RetrieveNodes(a.ctx, enrTree)
	if err != nil {
		return err
	}

	for _, d := range discoveredNodes {
		a.waku.AddDiscoveredPeer(d.PeerID, d.PeerInfo.Addrs, peerstore.DNSDiscovery, nil, true)
	}

	return nil
}
