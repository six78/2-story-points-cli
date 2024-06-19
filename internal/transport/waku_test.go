package transport

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/six78/2-story-points-cli/internal/testcommon"
	pp "github.com/six78/2-story-points-cli/pkg/protocol"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
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

func (s *WakuSuite) TestWakuCreate() {
	err := s.node.Initialize()
	s.Require().NoError(err)
}
