package waku

import (
	pp "2sp/pkg/protocol"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
)

func TestEncryptionSuite(t *testing.T) {
	suite.Run(t, new(EncryptionSuite))
}

type EncryptionSuite struct {
	suite.Suite
	node *Node
}

func (s *EncryptionSuite) SetupSuite() {
	logger, err := zap.NewDevelopment()
	s.Require().NoError(err)

	// For this test we only need roomCache and logger
	s.node = &Node{
		logger:    logger,
		roomCache: NewRoomCache(logger),
	}
}

func (s *EncryptionSuite) TestPublicEncryption() {
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
