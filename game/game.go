package game

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"net/url"
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

	dealer           bool
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
		dealer:    config.Dealer,
	}
}

func (g *Game) Start() {
	go g.processIncomingMessages()
	go g.publishOnlineState()

	if g.dealer {
		go g.publishState()
	}
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
		g.logger.Error("failed to unmarshal message", zap.Error(err))
		return
	}
	logger := g.logger.With(zap.String("type", string(message.Type)))
	switch message.Type {
	case protocol.MessageTypeState:
		if g.dealer {
			logger.Warn("dealer should not receive state messages")
			return
		}
		var state protocol.GameStateMessage
		err := json.Unmarshal(payload, &state)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.currentState = state.State
		g.notifyChangedState()
	case protocol.MessageTypePlayerOnline:
		var playerOnline protocol.PlayerOnlineMessage
		err := json.Unmarshal(payload, &playerOnline)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.logger.Info("player is online", zap.String("name", string(playerOnline.Name)))
		if g.dealer {
			//playerID := protocol.PlayerID(playerOnline.Name)
			//if _, ok := g.currentState.Players[playerID]; !ok {

			if !slices.Contains(g.currentState.Players, playerOnline.Name) {
				g.currentState.Players = append(g.currentState.Players, playerOnline.Name)
				g.notifyChangedState()
				go g.publishState()
			}
		}
	case protocol.MessageTypePlayerVote:
		if !g.dealer {
			return
		}
		var playerVote protocol.PlayerVote
		err := json.Unmarshal(payload, &playerVote)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.logger.Info("player voted",
			zap.String("name", string(playerVote.VoteBy)),
			zap.String("voteFor", playerVote.VoteFor),
			zap.Int("voteResult", int(playerVote.VoteResult)))

		if g.currentState.TempVoteResult == nil {
			g.currentState.TempVoteResult = make(map[protocol.Player]protocol.VoteResult)
		}
		g.currentState.TempVoteResult[playerVote.VoteBy] = playerVote.VoteResult
		g.notifyChangedState()
		go g.publishState()
	default:
		logger.Warn("unsupported message type")
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

func (g *Game) notifyChangedState() {
	for _, subscriber := range g.stateSubscribers {
		subscriber <- &g.currentState
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

func (g *Game) publishState() {
	for {
		select {
		case <-time.After(config.StateMessagePeriod):
			g.PublishState()
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

func (g *Game) publishMessage(message any) {
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
	g.publishMessage(protocol.PlayerOnlineMessage{
		Message: protocol.Message{
			Type:      protocol.MessageTypePlayerOnline,
			Timestamp: g.timestamp(),
		},
		Name: protocol.Player(user),
	})
}

func (g *Game) PublishVote(vote int) {
	g.logger.Debug("publishing vote", zap.Int("vote", vote))
	g.publishMessage(protocol.PlayerVote{
		Message: protocol.Message{
			Type:      protocol.MessageTypePlayerVote,
			Timestamp: g.timestamp(),
		},
		VoteBy:     protocol.Player(config.PlayerID),
		VoteFor:    g.currentState.VoteItem.ID,
		VoteResult: protocol.VoteResult(vote),
	})
}

func (g *Game) PublishState() {
	g.logger.Debug("publishing state")
	g.publishMessage(protocol.GameStateMessage{
		Message: protocol.Message{
			Type:      protocol.MessageTypeState,
			Timestamp: g.timestamp(),
		},
		State: g.currentState,
	})
}

func (g *Game) timestamp() int64 {
	return time.Now().UnixMilli()
}

func (g *Game) Deal(input string) error {
	item := protocol.VoteItem{}

	itemUuid, err := uuid.NewUUID()
	if err != nil {
		return errors.New("failed to generate UUID")
	}

	item.ID = itemUuid.String()

	u, err := url.Parse(input)
	if err == nil {
		item.URL = u.String()
		// TODO: fetch title/description
	} else {
		item.Name = input
	}

	g.currentState.VoteItem = item
	g.notifyChangedState()
	go g.publishState()

	return nil
}
