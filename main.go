package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/waku-org/go-waku/waku/v2/dnsdisc"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/pb"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
	"net"
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
	contentTopic, err := protocol.NewContentTopic("six78", "1", "helloworld", "json")
	if err != nil {
		logger.Error("failed to create content topic", zap.Error(err))
		return
	}

	hostAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", 30303))
	if err != nil {
		logger.Error("failed to resolve TCP address", zap.Error(err))
		return
	}

	ctx := context.Background()

	discoveryURL := "enrtree://AO47IDOLBKH72HIZZOXQP6NMRESAN7CHYWIBNXDXWRJRZWLODKII6@test.wakuv2.nodes.status.im"
	discoveryNodes, err := dnsdisc.RetrieveNodes(context.Background(), discoveryURL)
	if err != nil {
		panic(err)
	}

	logger.Info("retrieved nodes", zap.Any("discoveryNodes", discoveryNodes))

	var nodes []*enode.Node

	for _, n := range discoveryNodes {
		nodes = append(nodes, n.ENR)
	}

	logger.Info("nodes", zap.Any("nodes", nodes))

	options := []node.WakuNodeOption{
		node.WithWakuRelay(),
		node.WithLightPush(),
		//node.WithLogger(logger),
		node.WithHostAddress(hostAddr),
		node.WithDiscoveryV5(12345, nodes, true),
	}

	wakuNode, err := node.New(options...)
	if err != nil {
		logger.Error("failed to create waku node", zap.Error(err))
		return
	}

	if err := wakuNode.Start(context.Background()); err != nil {
		logger.Error("failed to start waku node", zap.Error(err))
		return
	}

	contentFilter := protocol.NewContentFilter(relay.DefaultWakuTopic, contentTopic.String())
	sub, err := wakuNode.Relay().Subscribe(ctx, contentFilter)
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		i := 0
		for {
			time.Sleep(5 * time.Second)
			i++
			msg := &pb.WakuMessage{
				Payload:      []byte(fmt.Sprintf("Hello, world! %d", i)),
				Version:      &appVersion,
				ContentTopic: contentTopic.String(),
				Timestamp:    utils.GetUnixEpoch(),
			}

			messageID, err := wakuNode.Relay().Publish(context.Background(), msg)
			if err != nil {
				logger.Info("failed to publish message", zap.Error(err))
			} else {
				logger.Info("published message", zap.String("messageID", hex.EncodeToString(messageID)))
			}
		}
	}()

	//go func() {
	for value := range sub[0].Ch {
		fmt.Println("Received msg:", string(value.Message().Payload))
	}
	//}()
}
