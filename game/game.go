package game

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"time"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
)

type StateSubscription chan *protocol.State

type Transport interface {
	SubscribeToMessages() (chan []byte, error)
	PublishMessage(payload []byte) error
}

type Game struct {
	logger    *zap.Logger
	ctx       context.Context
	quit      context.CancelFunc
	transport Transport

	currentState     protocol.State
	stateSubscribers []StateSubscription
}

func NewGame(transport Transport) *Game {
	ctx, quit := context.WithCancel(context.Background())
	return &Game{
		logger:    config.Logger.Named("game"),
		ctx:       ctx,
		quit:      quit,
		transport: transport,
	}
}

func (g *Game) Start() {
	go g.processIncomingMessages()
	go g.publishOnlineState()
}

func (g *Game) Stop() {
	for _, subscriber := range g.stateSubscribers {
		close(subscriber)
	}
	g.quit()
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
		g.publishChangedState(&g.currentState)
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

func (g *Game) CurrentState() *protocol.State {
	return &g.currentState
}

func (g *Game) publishChangedState(state *protocol.State) {
	for _, subscriber := range g.stateSubscribers {
		subscriber <- state
	}
}

func (g *Game) publishOnlineState() {
	g.PublishUserOnline(config.PlayerName)
	for {
		select {
		case <-time.After(config.OnlineMessagePeriod):
			g.PublishUserOnline(config.PlayerName)
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *Game) processIncomingMessages() {
	sub, err := g.transport.SubscribeToMessages()
	// TODO: defer unsubscribe

	if err != nil {
		g.logger.Error("failed to subscribe to messages", zap.Error(err))
		return
	}

	for {
		select {
		case payload, more := <-sub:
			if !more {
				return
			}
			g.processMessage(payload)
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *Game) publishMessage(message protocol.Message) {
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

func (g *Game) PublishUserOnline(user string) {
	g.logger.Debug("publishing online state", zap.String("user", user))
	g.publishMessage(protocol.Message{
		Type: protocol.MessageTypePlayerOnline,
		Name: protocol.Player(user),
	})
}

func (g *Game) PublishVote(vote int) {
	g.logger.Debug("publishing vote", zap.Int("vote", vote))
	g.publishMessage(protocol.Message{
		Type:       protocol.MessageTypePlayerVote,
		Name:       protocol.Player(config.PlayerName),
		VoteFor:    g.currentState.VoteItem.ID,
		VoteResult: protocol.VoteResult(vote),
	})
}
