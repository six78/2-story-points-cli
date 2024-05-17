package game

import (
	"2sp/pkg/protocol"
	"github.com/google/uuid"
)

func GeneratePlayerID() (protocol.PlayerID, error) {
	playerUUID := uuid.New()
	return protocol.PlayerID(playerUUID.String()), nil
}

func GenerateIssueID() (protocol.IssueID, error) {
	itemUUID := uuid.New()
	return protocol.IssueID(itemUUID.String()), nil
}
