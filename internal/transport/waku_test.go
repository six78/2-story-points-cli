package transport

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/suite"
	"github.com/waku-org/go-waku/waku/v2/node"
	wakuenr "github.com/waku-org/go-waku/waku/v2/protocol/enr"
	"go.uber.org/zap"

	"github.com/six78/2-story-points-cli/internal/testcommon"
	pp "github.com/six78/2-story-points-cli/pkg/protocol"
)

func TestWakuSuite(t *testing.T) {
	suite.Run(t, new(WakuSuite))
}

type WakuSuite struct {
	testcommon.Suite
	node   *Node
	cancel func()
}

func (s *WakuSuite) SetupSuite() {
	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())

	logger, err := zap.NewDevelopment()
	s.Require().NoError(err)

	// Skip initialization, for this test we only need roomCache and logger
	s.node = NewNode(ctx, logger)
}

func (s *WakuSuite) TearDownSuite() {
	s.cancel()
}

func (s *WakuSuite) TestPublicEncryption() {
	room, err := pp.NewRoom()
	s.Require().NoError(err)

	payload := make([]byte, 100)
	gofakeit.Slice(payload)

	message, err := s.node.buildWakuMessage(room, payload)
	s.Require().NoError(err)

	err = s.node.encryptPublicPayload(room, message)
	s.Require().NoError(err)

	decryptedPayload, err := decryptMessage(room, message)
	s.Require().NoError(err)

	s.Require().Equal(payload, decryptedPayload)
}

func (s *WakuSuite) TestWakuInitialize() {
	err := s.node.Initialize()
	s.Require().NoError(err)
}

func (s *WakuSuite) TestParseEnrProtocols() {
	p := parseEnrProtocols(wakuenr.WakuEnrBitfield(0b00000000))
	s.Require().Empty(p)

	p = parseEnrProtocols(wakuenr.WakuEnrBitfield(0b00000001))
	s.Require().Equal("relay", p)

	p = parseEnrProtocols(wakuenr.WakuEnrBitfield(0b00000010))
	s.Require().Equal("store", p)

	p = parseEnrProtocols(wakuenr.WakuEnrBitfield(0b00000011))
	s.Require().Equal("store,relay", p)

	p = parseEnrProtocols(wakuenr.WakuEnrBitfield(0b00000100))
	s.Require().Equal("filter", p)

	p = parseEnrProtocols(wakuenr.WakuEnrBitfield(0b00001000))
	s.Require().Equal("lightpush", p)

	p = parseEnrProtocols(wakuenr.WakuEnrBitfield(0b00001111))
	s.Require().Equal("lightpush,filter,store,relay", p)
}

func (s *WakuSuite) TestWatchConnectionStatus() {
	err := s.node.Initialize()
	s.Require().NoError(err)

	sub := s.node.SubscribeToConnectionStatus()

	finished := make(chan struct{})

	go func() {
		s.node.watchConnectionStatus()
		close(finished)
	}()

	peerID := peer.ID(gofakeit.UUID())

	for _, connected := range []bool{true, false} {
		expectedCount := 0
		if connected {
			expectedCount = 1
		}

		sent := node.PeerConnection{
			PeerID:    peerID,
			Connected: connected,
		}

		s.node.peerConnection <- sent

		select {
		case received := <-sub:
			s.Require().Equal(connected, received.IsOnline)
			s.Require().False(received.HasHistory)
			s.Require().Equal(expectedCount, received.PeersCount)
			s.Require().True(reflect.DeepEqual(received, s.node.ConnectionStatus()))
		case <-time.After(500 * time.Millisecond):
			s.Require().Fail("timeout waiting for connection status")
		}
	}

	close(s.node.peerConnection)

	select {
	case <-finished:
		break
	case <-time.After(500 * time.Millisecond):
		s.Require().Fail("timeout waiting for connection status watch finish")
	}
}
