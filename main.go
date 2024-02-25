package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/peerstore"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
	"net"
	"strconv"
	"time"

	"github.com/waku-org/go-waku/waku/v2/node"
)

//const contentTopic = "/six78/1/helloworld/json"
//const appVersion = uint32(1)

func main() {
	fmt.Println("Hello, world!")
	logger, err := zap.NewDevelopment()

	if err != nil {
		fmt.Println("failed to configure logging: %w", err)
		return
	}

	appVersion := uint32(1)
	contentTopic, err := protocol.NewContentTopic("six78", strconv.Itoa(int(appVersion)), "helloworld", "json")
	if err != nil {
		logger.Error("failed to create content topic", zap.Error(err))
		return
	}

	hostAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	if err != nil {
		logger.Error("failed to resolve TCP address", zap.Error(err))
		return
	}

	ctx := context.Background()

	prodEnrTre := "enrtree://ANEDLO25QVUGJOUTQFRYKWX6P4Z4GKVESBMHML7DZ6YK4LGS5FC5O@prod.wakuv2.nodes.status.im"
	//discoveryURL := "enrtree://AO47IDOLBKH72HIZZOXQP6NMRESAN7CHYWIBNXDXWRJRZWLODKII6@test.wakuv2.nodes.status.im"
	discoveredNodes, err := dnsdisc.RetrieveNodes(ctx, prodEnrTre)
	if err != nil {
		panic(err)
	}

	logger.Info("retrieved nodes", zap.Any("discoveredNodes", discoveredNodes))

	//var nodes []*enode.Node
	//
	//for _, n := range discoveredNodes {
	//	nodes = append(nodes, n.ENR)
	//}

	//logger.Info("nodes", zap.Any("nodes", nodes))

	connNotifier := make(chan node.PeerConnection, 100)

	go func() {
		for {
			conn := <-connNotifier
			logger.Info("PEER CONNECTION", zap.String("peerID", conn.PeerID.String()), zap.Bool("connected", conn.Connected))
		}
	}()

	options := []node.WakuNodeOption{
		node.WithWakuRelay(),
		node.WithLightPush(),
		//node.WithLogger(logger),
		//node.WithLogLevel(zap.DebugLevel),
		node.WithHostAddress(hostAddr),
		//node.WithDiscoveryV5(60000, nodes, true),
		node.WithConnectionNotification(connNotifier),
	}

	options = append(options, node.DefaultWakuNodeOptions...)

	wakuNode, err := node.New(options...)
	if err != nil {
		logger.Error("failed to create waku node", zap.Error(err))
		return
	}

	if err := wakuNode.Start(ctx); err != nil {
		logger.Error("failed to start waku node", zap.Error(err))
		return
	}

	peerID := wakuNode.ID()
	logger.Info("WAKU NODE STARTED", zap.String("peerID", peerID))

	for _, d := range discoveredNodes {
		wakuNode.AddDiscoveredPeer(d.PeerID, d.PeerInfo.Addrs, peerstore.DNSDiscovery, nil, true)
	}

	//wakuNode.AddPeer(maddr, wps.Static, []string{pubsubTopicStr}, relay.WakuRelayID_v200)

	go func() {
		messageVersion := uint32(0)
		i := 0
		for {
			time.Sleep(5 * time.Second)
			i++

			msg := &pb.WakuMessage{
				Payload:      []byte(fmt.Sprintf("Hello from Go mazafaka (%d)", i)),
				Version:      &messageVersion,
				ContentTopic: contentTopic.String(),
				Timestamp:    utils.GetUnixEpoch(),
			}

			publishOptions := []relay.PublishOption{relay.WithPubSubTopic(relay.DefaultWakuTopic)}
			messageID, err := wakuNode.Relay().Publish(ctx, msg, publishOptions...)
			if err != nil {
				logger.Info("failed to publish message", zap.Error(err))
			} else {
				logger.Info("published message", zap.String("messageID", hex.EncodeToString(messageID)))
			}
		}
	}()

	contentFilter := protocol.NewContentFilter(relay.DefaultWakuTopic, contentTopic.String())
	sub, err := wakuNode.Relay().Subscribe(ctx, contentFilter)
	if err != nil {
		fmt.Println(err)
		return
	}

	//go func() {
	for value := range sub[0].Ch {
		logger.Info("<<< MESSAGE RECEIVED",
			zap.String("text", string(value.Message().Payload)),
			zap.String("string", value.Message().String()),
		)
	}
	//}()
}
