package waku

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	"time"
	"waku-poker-planning/config"
	pp "waku-poker-planning/protocol"
)

var fleets = map[string]string{
	"wakuv2.prod": "enrtree://ANEDLO25QVUGJOUTQFRYKWX6P4Z4GKVESBMHML7DZ6YK4LGS5FC5O@prod.wakuv2.nodes.status.im",
	"wakuv2.test": "enrtree://AO47IDOLBKH72HIZZOXQP6NMRESAN7CHYWIBNXDXWRJRZWLODKII6@test.wakuv2.nodes.status.im",
}

type Node struct {
	waku   *node.WakuNode
	ctx    context.Context
	logger *zap.Logger

	pubsubTopic string

	wakuConnectionStatus  chan node.ConnStatus
	connectionStatus      node.ConnStatus
	connectionStatusMutex sync.Mutex
}

func NewNode(ctx context.Context, logger *zap.Logger) (*Node, error) {

	hostAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve TCP address")
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
		ctx:                  ctx,
		logger:               logger.Named("waku"),
		pubsubTopic:          relay.DefaultWakuTopic,
		wakuConnectionStatus: wakuConnectionStatus,
	}, nil
}

func (n *Node) Start() error {
	err := n.waku.Start(n.ctx)
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
	enrTree, ok := fleets[config.Fleet()]
	if !ok {
		return errors.Errorf("unknown fleet %s", config.Fleet())
	}

	discoveredNodes, err := dnsdisc.RetrieveNodes(n.ctx, enrTree)
	if err != nil {
		return err
	}

	for _, d := range discoveredNodes {
		n.waku.AddDiscoveredPeer(d.PeerID, d.PeerInfo.Addrs, peerstore.DNSDiscovery, nil, true)
	}

	return nil
}

func (n *Node) PublishUnencryptedMessage(session *pp.Session, payload []byte) error {
	message := &pb.WakuMessage{
		Payload: payload,
	}
	return n.publishWakuMessage(session, message)
}

func (n *Node) publishWakuMessage(session *pp.Session, message *pb.WakuMessage) error {
	contentTopic, err := sessionContentTopic(session)
	if err != nil {
		return errors.Wrap(err, "failed to build content topic")
	}

	version := uint32(0)

	message.Version = &version
	message.ContentTopic = contentTopic
	message.Timestamp = utils.GetUnixEpoch()

	publishOptions := []relay.PublishOption{
		relay.WithPubSubTopic(n.pubsubTopic),
	}

	messageID, err := n.waku.Relay().Publish(n.ctx, message, publishOptions...)

	if err != nil {
		n.logger.Error("failed to publish message", zap.Error(err))
		return errors.Wrap(err, "failed to publish message")
	}

	n.logger.Info("message sent", zap.String("messageID", hex.EncodeToString(messageID)))
	return nil
}

func (n *Node) WaitForPeersConnected() bool {
	if n.waku.PeerCount() > 0 {
		return true
	}
	for {
		select {
		case <-time.After(20 * time.Second):
			return false
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

func (n *Node) SubscribeToMessages(session *pp.Session) (chan []byte, error) {
	contentTopic, err := sessionContentTopic(session)
	contentFilter := protocol.NewContentFilter(relay.DefaultWakuTopic, contentTopic)
	subs, err := n.waku.Relay().Subscribe(n.ctx, contentFilter)
	unsubscribe := func() {
		err := n.waku.Relay().Unsubscribe(n.ctx, contentFilter)
		if err != nil {
			n.logger.Warn("failed to unsubscribe from relay", zap.Error(err))
		}
	}

	if err != nil {
		fmt.Println(err)
		return nil, errors.Wrap(err, "failed to subscribe to relay")
	}

	if len(subs) != 1 {
		unsubscribe()
		return nil, errors.Errorf("unexpected number of subscriptions: %d", len(subs))
	}

	in := subs[0].Ch
	out := make(chan []byte, 10)

	go func() {
		defer close(out)
		defer unsubscribe()

		for value := range in {
			n.logger.Info("waku message received",
				zap.String("payload", string(value.Message().Payload)),
			)
			out <- value.Message().Payload
		}
	}()

	return out, nil
}

func sessionContentTopic(info *pp.Session) (string, error) {
	if len(info.SymmetricKey) < 4 {
		return "", errors.New("symmetric key too short")
	}

	version := strconv.Itoa(int(pp.Version))
	contentTopicName := hexutil.Encode(info.SymmetricKey[:4])
	contentTopic, err := protocol.NewContentTopic("six78", version, contentTopicName, "json")
	if err != nil {
		return "", errors.Wrap(err, "failed to create content topic")
	}
	return contentTopic.String(), nil
}

//func (n *Node) SendPublicMessage(payload []byte) error {
//	wp.EncodeWakuMessage()
//}
//
//func (n *Node) SendPrivateMessage(payload []byte) error {
//
//}
