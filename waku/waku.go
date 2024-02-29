package waku

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"net"
	"strconv"
	"sync"
	"waku-poker-planning/config"
	pp "waku-poker-planning/protocol"
)

var fleets = map[string]string{
	"wakuv2.prod": "enrtree://ANEDLO25QVUGJOUTQFRYKWX6P4Z4GKVESBMHML7DZ6YK4LGS5FC5O@prod.wakuv2.nodes.status.im",
	"wakuv2.test": "enrtree://AO47IDOLBKH72HIZZOXQP6NMRESAN7CHYWIBNXDXWRJRZWLODKII6@test.wakuv2.nodes.status.im",
}

type Node struct {
	waku   *node.WakuNode
	logger *zap.Logger

	pubsubTopic  string
	contentTopic string

	wakuConnectionStatus  chan node.ConnStatus
	connectionStatus      node.ConnStatus
	connectionStatusMutex sync.Mutex
}

func NewNode(logger *zap.Logger) (*Node, error) {

	hostAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve TCP address")
	}

	contentTopic, err := calculateContentTopic(config.SessionName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to calculate content topic")
	}

	wakuConnectionStatus := make(chan node.ConnStatus)

	options := []node.WakuNodeOption{
		node.WithWakuRelay(),
		node.WithLightPush(),
		node.WithLogger(logger.Named("waku")),
		//node.WithLogLevel(zap.DebugLevel),
		node.WithHostAddress(hostAddr),
		//node.WithDiscoveryV5(60000, nodes, true),
		node.WithConnectionStatusChannel(wakuConnectionStatus),
	}

	options = append(options, node.DefaultWakuNodeOptions...)

	wakuNode, err := node.New(options...)
	if err != nil {

		return nil, errors.Wrap(err, "failed to create waku node")
	}

	return &Node{
		waku:                 wakuNode,
		logger:               logger.Named("waku"),
		pubsubTopic:          relay.DefaultWakuTopic,
		contentTopic:         contentTopic,
		wakuConnectionStatus: wakuConnectionStatus,
	}, nil
}

func (n *Node) Start() error {
	err := n.waku.Start(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to start waku node")
	}

	n.logger.Info("waku started", zap.String("peerID", n.waku.ID()))

	err = n.discoverNodes()
	if err != nil {
		return errors.Wrap(err, "failed to discover nodes")
	}

	go n.watchConnectionStatus()
	//go n.receiveMessages(contentTopic)

	return nil
}

func (n *Node) Stop() {
	n.waku.Stop()
}

func (n *Node) watchConnectionStatus() {
	var more bool
	for {
		n.connectionStatus, more = <-n.wakuConnectionStatus
		if !more {
			return
		}
		peersCount := len(maps.Keys(n.connectionStatus.Peers))
		n.logger.Debug("connection status", zap.Any("peersCount", peersCount))
	}
}

func (n *Node) discoverNodes() error {
	enrTree, ok := fleets[config.Fleet]
	if !ok {
		return errors.Errorf("unknown fleet %s", config.Fleet)
	}

	discoveredNodes, err := dnsdisc.RetrieveNodes(context.TODO(), enrTree)
	if err != nil {
		return err
	}

	for _, d := range discoveredNodes {
		n.waku.AddDiscoveredPeer(d.PeerID, d.PeerInfo.Addrs, peerstore.DNSDiscovery, nil, true)
	}

	return nil
}

func (n *Node) PublishMessage(payload []byte) error {
	version := uint32(0)
	message := &pb.WakuMessage{
		Payload:      payload,
		Version:      &version,
		ContentTopic: n.contentTopic,
		Timestamp:    utils.GetUnixEpoch(),
	}
	publishOptions := []relay.PublishOption{
		relay.WithPubSubTopic(n.pubsubTopic),
	}

	messageID, err := n.waku.Relay().Publish(context.Background(), message, publishOptions...)

	if err != nil {
		n.logger.Error("failed to publish message", zap.Error(err))
		return errors.Wrap(err, "failed to publish message")
	}

	n.logger.Info("message sent",
		zap.String("messageID", hex.EncodeToString(messageID)),
		zap.String("payload", string(payload)))

	return nil
}

func (n *Node) WaitForPeersConnected() bool {
	if n.waku.PeerCount() > 0 {
		return true
	}
	for {
		select {
		case connStatus, more := <-n.wakuConnectionStatus:
			if !more {
				return false
			}
			if len(connStatus.Peers) >= 0 {
				return true
			}
		}
	}
}

func (n *Node) SubscribeToMessages() (chan []byte, error) {
	contentFilter := protocol.NewContentFilter(relay.DefaultWakuTopic, n.contentTopic)
	subs, err := n.waku.Relay().Subscribe(context.Background(), contentFilter)

	if err != nil {
		fmt.Println(err)
		return nil, errors.Wrap(err, "failed to subscribe to relay")
	}

	if len(subs) != 1 {
		return nil, errors.Errorf("unexpected number of subscriptions: %d", len(subs))
	}

	in := subs[0].Ch
	out := make(chan []byte, 10)

	go func() {
		defer close(out)

		for value := range in {
			n.logger.Info("<<< MESSAGE RECEIVED",
				zap.String("payload", string(value.Message().Payload)),
			)
			out <- value.Message().Payload
		}
	}()

	return out, nil
}

func calculateContentTopic(name string) (string, error) {
	contentTopic, err := protocol.NewContentTopic("six78", strconv.Itoa(pp.Version), name, "json")
	if err != nil {
		return "", errors.Wrap(err, "failed to create content topic")
	}
	return contentTopic.String(), nil
}
