package game

import (
	"encoding/json"
	"go.uber.org/zap"
	"time"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
)

type StateSubscription chan protocol.State

type Transport interface {
	SubscribeToMessages() (chan []byte, error)
	PublishMessage(payload []byte) error
}

type Game struct {
	logger    *zap.Logger
	transport Transport

	currentState     protocol.State
	stateSubscribers []StateSubscription
}

func NewGame(logger *zap.Logger, transport Transport) *Game {
	return &Game{
		logger:    logger.Named("game"),
		transport: transport,
	}
}

func (g *Game) Start() {
	go g.processIncomingMessages()
	go g.publishOnlineState()
}

func (g *Game) processMessage(payload []byte) {
	g.logger.Debug("processing message", zap.String("payload", string(payload)))
	message := protocol.Message{}
	err := json.Unmarshal(payload, &message)
	if err != nil {
		g.logger.Warn("failed to unmarshal message", zap.Error(err))
		return
	}
	switch message.Type {
	case protocol.MessageTypeState:
		g.currentState = *message.State
		g.publishChangedState(g.currentState)
	case protocol.MessageTypePlayerOnline:
		g.logger.Info("player is online", zap.String("name", string(message.Name)))
	default:
		g.logger.Warn("unsupported message type", zap.String("type", string(message.Type)))
	}
}

func (g *Game) SubscribeToStateChanges() StateSubscription {
	channel := make(StateSubscription, 10)
	g.stateSubscribers = append(g.stateSubscribers, channel)
	return channel
}

func (g *Game) publishChangedState(state protocol.State) {
	for _, subscriber := range g.stateSubscribers {
		subscriber <- state
	}
}

func (g *Game) publishOnlineState() {
	for {
		time.Sleep(config.OnlineMessagePeriod)
		g.logger.Debug("publishing online state")
		g.PublishMessage(protocol.Message{
			Type: protocol.MessageTypePlayerOnline,
			Name: protocol.Player(config.PlayerName),
		})
	}
}

func (g *Game) processIncomingMessages() {
	sub, err := g.transport.SubscribeToMessages()
	if err != nil {
		g.logger.Error("failed to subscribe to messages", zap.Error(err))
		return
	}
	for {
		payload, more := <-sub
		if !more {
			return
		}
		g.processMessage(payload)
	}
}

func (g *Game) PublishMessage(message protocol.Message) {
	payload, err := json.Marshal(message)
	if err != nil {
		g.logger.Error("failed to marshal message", zap.Error(err))
		return
	}
	err = g.transport.PublishMessage(payload)
	if err != nil {
		g.logger.Error("failed to publish message", zap.Error(err))
		return
	}
}

//func (g *Game) Vote(vote int) State {
//	return g.currentState
//}
