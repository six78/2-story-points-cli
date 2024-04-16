package waku

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/node"
	wp "github.com/waku-org/go-waku/waku/v2/payload"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	wakuenr "github.com/waku-org/go-waku/waku/v2/protocol/enr"
	"github.com/waku-org/go-waku/waku/v2/protocol/lightpush"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/protocol/subscription"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
	"net"
	"strings"
	"time"
	"waku-poker-planning/config"
	pp "waku-poker-planning/protocol"
)

var fleets = map[string]string{
	"wakuv2.prod":  "enrtree://ANEDLO25QVUGJOUTQFRYKWX6P4Z4GKVESBMHML7DZ6YK4LGS5FC5O@prod.wakuv2.nodes.status.im",
	"wakuv2.test":  "enrtree://AO47IDOLBKH72HIZZOXQP6NMRESAN7CHYWIBNXDXWRJRZWLODKII6@test.wakuv2.nodes.status.im",
	"waku.sandbox": "enrtree://AIRVQ5DDA4FFWLRBCHJWUWOO6X6S4ZTZ5B667LQ6AJU6PEYDLRD5O@sandbox.waku.nodes.status.im",
}

type Node struct {
	waku   *node.WakuNode
	ctx    context.Context
	logger *zap.Logger

	pubsubTopic          string
	wakuConnectionStatus chan node.ConnStatus
	roomCache            ContentTopicCache
	stats                *Stats
	lightMode            bool
	statusSubscribers    []ConnectionStatusSubscription
}

type ConnectionStatus struct {
	IsOnline   bool
	HasHistory bool
	PeersCount int
}

type ConnectionStatusSubscription chan ConnectionStatus

func NewNode(ctx context.Context, logger *zap.Logger) (*Node, error) {

	hostAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve TCP address")
	}

	wakuConnectionStatus := make(chan node.ConnStatus)

	options := []node.WakuNodeOption{
		node.WithLogger(logger.Named("waku")),
		//node.WithDNS4Domain(),
		//node.WithLogLevel(zap.DebugLevel),
		node.WithHostAddress(hostAddr),
		//node.WithDiscoveryV5(60000, nodes, true),
		node.WithConnectionStatusChannel(wakuConnectionStatus),
	}

	if config.WakuLightMode() {
		options = append(options,
			node.WithLightPush(),
			node.WithWakuFilterLightNode(),
		)
	} else {
		options = append(options,
			node.WithWakuRelay(),
		)
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
		stats:                NewStats(),
		lightMode:            config.WakuLightMode(),
	}, nil
}

func (n *Node) Start() error {
	go n.watchConnectionStatus()

	err := n.waku.Start(n.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start waku node")
	}

	n.logger.Info("waku started", zap.String("peerID", n.waku.ID()))

	if staticNodes := config.WakuStaticNodes(); len(staticNodes) != 0 {
		err = n.addStaticNodes(staticNodes)
		if err != nil {
			return errors.Wrap(err, "failed to add static nodes")
		}
	} else {
		err = n.discoverNodes()
		if err != nil {
			return errors.Wrap(err, "failed to discover nodes")
		}
	}

	return nil
}

func (n *Node) Stop() {
	n.waku.Stop()
}

func parseEnrProtocols(v wakuenr.WakuEnrBitfield) string {
	var out []string
	if v&(1<<3) == 8 {
		out = append(out, "lightpush")
	}
	if v&(1<<2) == 4 {
		out = append(out, "filter")
	}
	if v&(1<<1) == 2 {
		out = append(out, "store")
	}
	if v&(1<<0) == 1 {
		out = append(out, "relay")
	}
	return strings.Join(out, ",")
}

func (n *Node) discoverNodes() error {
	enrTree, ok := fleets[config.Fleet()]
	if !ok {
		return errors.Errorf("unknown fleet %s", config.Fleet())
	}

	// Otherwise run discovery
	var options []dnsdisc.DNSDiscoveryOption
	if nameserver := config.Nameserver(); nameserver != "" {
		options = append(options, dnsdisc.WithNameserver(nameserver))
	}

	discoveredNodes, err := dnsdisc.RetrieveNodes(n.ctx, enrTree, options...)
	if err != nil {
		return err
	}

	n.logger.Debug("discovered nodes", zap.String("entree", enrTree))

	for _, d := range discoveredNodes {
		enrField := new(wakuenr.WakuEnrBitfield)
		err = d.ENR.Record().Load(enr.WithEntry(wakuenr.WakuENRField, &enrField))
		if err != nil {
			return errors.Wrap(err, "failed to load waku enr field")
		}
		n.logger.Debug("discover node",
			zap.String("peerID", d.PeerID.String()),
			zap.Any("peerInfo", d.PeerInfo),
			zap.Any("protocols", parseEnrProtocols(*enrField)),
		)
	}

	for _, d := range discoveredNodes {
		n.waku.AddDiscoveredPeer(d.PeerID, d.PeerInfo.Addrs, peerstore.DNSDiscovery, nil, true)
	}

	return nil
}

func (n *Node) addStaticNodes(staticNodes []string) error {
	for _, staticNode := range staticNodes {
		n.logger.Info("connecting to a static store node",
			zap.String("address", staticNode),
		)
		addr, err := multiaddr.NewMultiaddr(staticNode)
		if err != nil {
			return errors.Wrap(err, "failed to parse multiaddr")
		}

		err = n.DialPeer(addr)
		if err != nil {
			return errors.Wrap(err, "failed to dial static peer")
		}
	}

	return nil
}

func (n *Node) DialPeer(address multiaddr.Multiaddr) error {
	const dialTimeout = 10 * time.Second

	ctx, cancel := context.WithTimeout(n.ctx, dialTimeout)
	defer cancel()

	return n.waku.DialPeerWithMultiAddress(ctx, address)
}

func (n *Node) PublishUnencryptedMessage(room *pp.Room, payload []byte) error {
	message, err := n.buildWakuMessage(room, payload)
	if err != nil {
		return errors.Wrap(err, "failed to build waku message")
	}
	return n.publishWakuMessage(message)
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
		// PrivKey: Set a privkey if the message requires a signature
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

	return n.publishWakuMessage(message)
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

func (n *Node) publishWakuMessage(message *pb.WakuMessage) error {
	var err error
	var messageID []byte

	if n.lightMode {
		publishOptions := []lightpush.Option{
			lightpush.WithPubSubTopic(n.pubsubTopic),
		}
		messageID, err = n.waku.Lightpush().Publish(n.ctx, message, publishOptions...)
	} else {
		publishOptions := []relay.PublishOption{
			relay.WithPubSubTopic(n.pubsubTopic),
		}
		messageID, err = n.waku.Relay().Publish(n.ctx, message, publishOptions...)
	}

	if err != nil {
		n.logger.Error("failed to publish message", zap.Error(err))
		return errors.Wrap(err, "failed to publish message")
	}

	n.stats.IncrementSentMessages()

	n.logger.Info("message sent",
		zap.String("messageID", hex.EncodeToString(messageID)))

	return nil
}

func (n *Node) watchConnectionStatus() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case connStatus, more := <-n.wakuConnectionStatus:
			n.logger.Debug("waku connection status",
				zap.Any("connStatus", connStatus),
				zap.Bool("more", more),
			)
			if !more {
				return
			}
			n.notifyConnectionStatus(&connStatus)
		}
	}
}

//func (n *Node) WaitForPeersConnected() bool {
//	if n.waku.PeerCount() > 0 {
//		return true
//	}
//	ctx, cancel := context.WithTimeout(n.ctx, 20*time.Second)
//	defer cancel()
//	for {
//		select {
//		case <-ctx.Done():
//			return false
//		case connStatus, more := <-n.wakuConnectionStatus:
//			n.logger.Debug("waku connection status",
//				zap.Any("connStatus", connStatus),
//				zap.Bool("more", more),
//			)
//			if !more {
//				return false
//			}
//			if len(connStatus.Peers) >= 0 {
//				return true
//			}
//		}
//	}
//}

func (n *Node) SubscribeToMessages(room *pp.Room) (chan []byte, error) {
	contentTopic, err := n.roomCache.Get(room)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build content topic")
	}

	contentFilter := protocol.NewContentFilter(relay.DefaultWakuTopic, contentTopic)

	var in chan *protocol.Envelope
	var unsubscribe func()

	if n.lightMode {
		var subs []*subscription.SubscriptionDetails
		subs, err = n.waku.FilterLightnode().Subscribe(n.ctx, contentFilter)

		unsubscribe = func() {
			response, err := n.waku.FilterLightnode().Unsubscribe(n.ctx, contentFilter)
			if err != nil {
				n.logger.Warn("failed to unsubscribe from lightnode", zap.Error(err))
			}
			for _, err := range response.Errors() {
				n.logger.Warn("lightnode unsubscribe response error", zap.Error(err.Err))
			}
		}

		if len(subs) != 1 {
			unsubscribe()
			return nil, errors.Errorf("unexpected number of subscriptions: %d", len(subs))
		}

		in = subs[0].C

	} else {
		var subs []*relay.Subscription
		subs, err = n.waku.Relay().Subscribe(n.ctx, contentFilter)

		unsubscribe = func() {
			err := n.waku.Relay().Unsubscribe(n.ctx, contentFilter)
			if err != nil {
				n.logger.Warn("failed to unsubscribe from relay", zap.Error(err))
			}
		}

		if len(subs) != 1 {
			unsubscribe()
			return nil, errors.Errorf("unexpected number of subscriptions: %d", len(subs))
		}

		in = subs[0].Ch
	}

	if err != nil {
		n.logger.Error("failed to subscribe to content topic", zap.Bool("lightMode", n.lightMode))
		return nil, errors.Wrap(err, "failed to subscribe to content topic")
	}

	out := make(chan []byte, 10)

	go func() {
		defer close(out)
		defer unsubscribe()

		for value := range in {
			n.logger.Info("waku message received (relay)",
				zap.String("payload", string(value.Message().Payload)),
			)
			payload, err := decryptMessage(room, value.Message())
			if err != nil {
				n.logger.Warn("failed to decrypt message payload")
			}

			n.stats.IncrementReceivedMessages()
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

func (n *Node) Stats() Stats {
	return *n.stats
}

func (n *Node) SubscribeToConnectionStatus() ConnectionStatusSubscription {
	channel := make(ConnectionStatusSubscription, 10)
	n.statusSubscribers = append(n.statusSubscribers, channel)
	return channel
}

func (n *Node) notifyConnectionStatus(s *node.ConnStatus) {
	status := ConnectionStatus{
		IsOnline:   s.IsOnline,
		HasHistory: s.HasHistory,
		PeersCount: len(s.Peers),
	}

	for _, subscriber := range n.statusSubscribers {
		subscriber <- status
	}
}
