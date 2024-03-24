package game

import (
	"github.com/google/uuid"
	"waku-poker-planning/protocol"
)

func GeneratePlayerID() (protocol.PlayerID, error) {
	playerUUID := uuid.New()
	return protocol.PlayerID(playerUUID.String()), nil
}

func GenerateVoteItemID() (protocol.VoteItemID, error) {
	itemUUID := uuid.New()
	return protocol.VoteItemID(itemUUID.String()), nil
}
