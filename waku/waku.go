package waku

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/node"
	wp "github.com/waku-org/go-waku/waku/v2/payload"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
	"net"
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

	wakuConnectionStatus chan node.ConnStatus

	roomCache ContentTopicCache
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
		//node.WithDNS4Domain(),
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
		roomCache:            NewRoomCache(logger),
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

	//go n.watchConnectionStatus()
	//go n.receiveMessages(contentTopic)

	return nil
}

func (n *Node) Stop() {
	n.waku.Stop()
}

func (n *Node) discoverNodes() error {
	enrTree, ok := fleets[config.Fleet()]
	if !ok {
		return errors.Errorf("unknown fleet %s", config.Fleet())
	}

	var options []dnsdisc.DNSDiscoveryOption
	if nameserver := config.Nameserver(); nameserver != "" {
		options = append(options, dnsdisc.WithNameserver(nameserver))
	}

	discoveredNodes, err := dnsdisc.RetrieveNodes(n.ctx, enrTree, options...)
	if err != nil {
		return err
	}

	for _, d := range discoveredNodes {
		n.waku.AddDiscoveredPeer(d.PeerID, d.PeerInfo.Addrs, peerstore.DNSDiscovery, nil, true)
	}

	return nil
}

func (n *Node) PublishUnencryptedMessage(room *pp.Room, payload []byte) error {
	message, err := n.buildWakuMessage(room, payload)
	if err != nil {
		return errors.Wrap(err, "failed to build waku message")
	}
	return n.publishWakuMessage(room, message)
}

/*
	NOTE: Waku built-in encryption was a simple start, but it has a few disadvantages:
		1. It's fixed to 32-bytes key size
		   This makes RoomID too big even with a single SymmetricKey: "NrhbXYhn49Zo7LeKLQGQVRSjoBSLhLD6zSXKwqb3Podf"
		2. Because of this we have to pass pp.Room to this waku package.
		   I'm not sure if this is a good architecture decision.
*/

func (n *Node) encryptPublicPayload(room *pp.Room, message *pb.WakuMessage) error {
	keyInfo := &wp.KeyInfo{
		Kind:   wp.Symmetric,
		SymKey: room.SymmetricKey,
		//PrivKey: room.DealerPrivateKey, // TODO: Implement dealer signature support
	}

	return wp.EncodeWakuMessage(message, keyInfo)
}

func (n *Node) encryptPrivatePayload(room *pp.Room, message *pb.WakuMessage) error {
	keyInfo := &wp.KeyInfo{
		Kind:   wp.Asymmetric,
		PubKey: room.DealerPublicKey,
		SymKey: room.SymmetricKey,
	}

	return wp.EncodeWakuMessage(message, keyInfo)
}

func (n *Node) PublishPublicMessage(room *pp.Room, payload []byte) error {
	message, err := n.buildWakuMessage(room, payload)
	if err != nil {
		return errors.Wrap(err, "failed to build waku message")
	}

	err = n.encryptPublicPayload(room, message)
	if err != nil {
		return errors.Wrap(err, "failed to encrypt message")
	}

	return n.publishWakuMessage(room, message)
}

func (n *Node) PublishPrivateMessage(room *pp.Room, payload []byte) error {
	n.logger.Error("PublishPrivateMessage not implemented")
	return errors.New("PublishPrivateMessage not implemented")
}

func (n *Node) buildWakuMessage(room *pp.Room, payload []byte) (*pb.WakuMessage, error) {
	version := uint32(0)
	if config.EnableSymmetricEncryption {
		version = 1
	}

	contentTopic, err := n.roomCache.Get(room)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build content topic")
	}

	return &pb.WakuMessage{
		Payload:      payload,
		Version:      &version,
		ContentTopic: contentTopic,
		Timestamp:    utils.GetUnixEpoch(),
	}, nil
}

func (n *Node) publishWakuMessage(room *pp.Room, message *pb.WakuMessage) error {
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
	ctx, cancel := context.WithTimeout(n.ctx, 20*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return false
		case connStatus, more := <-n.wakuConnectionStatus:
			n.logger.Debug("<<< waku connection status",
				zap.Any("connStatus", connStatus),
				zap.Bool("more", more),
			)
			if !more {
				return false
			}
			if len(connStatus.Peers) >= 0 {
				return true
			}
		}
	}
}

func (n *Node) SubscribeToMessages(room *pp.Room) (chan []byte, error) {
	contentTopic, err := n.roomCache.Get(room)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build content topic")
	}
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
			payload, err := decryptMessage(room, value.Message())
			if err != nil {
				n.logger.Warn("failed to decrypt message payload")
			}
			out <- payload
		}
	}()

	return out, nil
}

func decryptMessage(room *pp.Room, message *pb.WakuMessage) ([]byte, error) {
	// NOTE: waku automatically decide to decrypt or not based on message.Version (0/1)
	//if !config.EnableSymmetricEncryption {
	//	return payload, nil
	//}

	keyInfo := &wp.KeyInfo{
		Kind:   wp.Symmetric,
		SymKey: room.SymmetricKey,
	}

	err := wp.DecodeWakuMessage(message, keyInfo)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode waku message")
	}

	return message.Payload, nil
}
