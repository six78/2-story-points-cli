package game

import (
	"2sp/pkg/protocol"
	"go.uber.org/zap"
)

func (g *Game) handlePlayerOnlineMessage(message *protocol.PlayerOnlineMessage) {
	g.logger.Info("player online message received", zap.Any("player", message.Player))

	message.Player.ApplyDeprecatedPatchOnReceive()

	// TODO: Store player pointers in a map

	index := g.playerIndex(message.Player.ID)
	if index < 0 {
		message.Player.Online = true
		message.Player.OnlineTimestampMilliseconds = g.timestamp()
		g.state.Players = append(g.state.Players, message.Player)
		g.notifyChangedState(true)
		g.logger.Info("player joined", zap.Any("player", message.Player))
		return
	}

	playerChanged := !g.state.Players[index].Online ||
		g.state.Players[index].Name != message.Player.Name

	g.state.Players[index].OnlineTimestampMilliseconds = g.timestamp()

	if !playerChanged {
		return
	}

	g.state.Players[index].Online = true
	g.state.Players[index].Name = message.Player.Name
	g.notifyChangedState(true)
}

func (g *Game) handlePlayerOfflineMessage(message *protocol.PlayerOfflineMessage) {
	g.logger.Info("player is offline", zap.Any("player", message.Player))

	index := g.playerIndex(message.Player.ID)
	if index < 0 {
		return
	}

	g.state.Players[index].Online = false
	g.notifyChangedState(true)
}
