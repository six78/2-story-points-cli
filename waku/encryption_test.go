package waku

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"testing"
	pp "waku-poker-planning/protocol"
)

func TestEncryptionSuite(t *testing.T) {
	suite.Run(t, new(EncryptionSuite))
}

type EncryptionSuite struct {
	suite.Suite
	node   *Node
	logger *zap.Logger
}

func (s *EncryptionSuite) SetupSuite() {
	var err error
	s.logger, err = zap.NewDevelopment()
	s.Require().NoError(err)

	// For this test we only need roomCache and logger
	s.node = &Node{
		logger:    s.logger,
		roomCache: NewRoomCache(s.logger),
	}
}

func (s *EncryptionSuite) TestPublicEncryption() {
	room, err := pp.NewRoom()
	s.Require().NoError(err)

	roomID, err := room.ToRoomID()
	s.Require().NoError(err)

	s.logger.Info("room created", zap.Any("roomID", roomID))

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

func (s *EncryptionSuite) TestPrivateEncryption() {
	dealerRoom, err := pp.NewRoom()
	s.Require().NoError(err)
	s.Require().NotEmpty(dealerRoom.SymmetricKey)
	s.Require().NotEmpty(dealerRoom.DealerPublicKey)
	s.Require().NotNil(dealerRoom.DealerPrivateKey)

	roomID, err := dealerRoom.ToRoomID()
	s.Require().NoError(err)

	s.logger.Info("dealerRoom created", zap.Any("roomID", roomID))

	room, err := pp.ParseRoomID(roomID.String())
	s.Require().NoError(err)
	s.Require().NotEmpty(room.SymmetricKey)
	s.Require().NotEmpty(room.DealerPublicKey)
	s.Require().Nil(room.DealerPrivateKey)

	payload := make([]byte, 100)
	gofakeit.Slice(payload)

	/*
			NOTE: We should know how the message is encrypted (symmetric/asymmetric) to decrypt it.
				 There are 2 ways to solve this:
		 			1. Separate content topics for those keys
						- Good to separate message streams, players won't even receive smth they can't decrypt.
						- Multiple topics is a bit of a mess.
					2. Make symmetric encryption on a different level.
						- We can encrypt only the vote value itself. Then everyone can take listen to this event and render
							that the player has voted. But won't know the vote value. The dealer won't have to duplicate state
							on each vote. Only send the revealed votes at the end.
						- How does a player get confirmation that his vote was received? 🤔
	*/

	//message, err := s.node.buildWakuMessage(room, payload)
	//s.Require().NoError(err)
	//
	//err = s.node.encryptPrivatePayload(room, message)
	//s.Require().NoError(err)
	//
	//// Endure usual player can't decrypt the message
	//decryptedPayload, err := decryptMessage(room, message)
	//s.Require().Error(err)
	//
	//// Ensure dealer can decrypt the message
	//decryptedPayload, err = decryptMessage(dealerRoom, message)
	//s.Require().NoError(err)
	//
	//s.Require().Equal(payload, decryptedPayload)
}
