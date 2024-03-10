package game

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	wp "github.com/waku-org/go-waku/waku/v2/payload"
	"go.uber.org/zap"
	"net/url"
	"time"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
)

type StateSubscription chan *protocol.State

type Transport interface {
	SubscribeToMessages(session *protocol.Session) (chan []byte, error)
	PublishUnencryptedMessage(session *protocol.Session, payload []byte) error
}

type Game struct {
	logger       *zap.Logger
	ctx          context.Context
	transport    Transport
	leaveSession chan struct{}

	player *protocol.Player

	session               *protocol.Session
	sessionID             string
	currentState          *protocol.State
	currentStateTimestamp int64
	stateSubscribers      []StateSubscription
}

func NewGame(ctx context.Context, transport Transport) *Game {
	playerUuid, err := uuid.NewUUID()
	if err != nil {
		panic(errors.Wrap(err, "failed to generate user uuid"))
	}

	return &Game{
		logger:       config.Logger.Named("game"),
		ctx:          ctx,
		transport:    transport,
		leaveSession: nil,
		player: &protocol.Player{
			ID:       protocol.PlayerID(playerUuid.String()),
			Name:     config.PlayerName(),
			IsDealer: false,
		},
		session: nil,
	}
}

func (g *Game) Start() {
	g.leaveSession = make(chan struct{})

	go g.processIncomingMessages()
	go g.publishOnlineState()

	if g.player.IsDealer {
		go g.publishStatePeriodically()
	}
}

func (g *Game) LeaveSession() {
	if g.leaveSession != nil {
		close(g.leaveSession)
	}
}

func (g *Game) Stop() {
	for _, subscriber := range g.stateSubscribers {
		close(subscriber)
	}
	g.stateSubscribers = nil
	g.LeaveSession()
	// WARNING: wait for all routines to finish
}

func (g *Game) generateSymmetricKey() (*wp.KeyInfo, error) {
	key := make([]byte, config.SymmetricKeyLength)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return &wp.KeyInfo{
		Kind:   wp.Symmetric,
		SymKey: key,
	}, nil
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
		if g.player.IsDealer {
			logger.Warn("dealer should not receive state messages")
			return
		}
		if message.Timestamp < g.currentStateTimestamp {
			logger.Warn("ignoring outdated state message")
			return
		}
		var state protocol.GameStateMessage
		err := json.Unmarshal(payload, &state)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.currentState = &state.State
		g.notifyChangedState()

	case protocol.MessageTypePlayerOnline:
		var playerOnline protocol.PlayerOnlineMessage
		err := json.Unmarshal(payload, &playerOnline)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.logger.Info("player is online", zap.Any("player", playerOnline.Player))
		if g.player.IsDealer {
			if _, ok := g.currentState.Players[playerOnline.Player.ID]; !ok {
				playerOnline.Player.Order = len(g.currentState.Players)
				if g.currentState.Players == nil {
					g.currentState.Players = make(map[protocol.PlayerID]protocol.Player)
				}
				g.currentState.Players[playerOnline.Player.ID] = playerOnline.Player
				g.notifyChangedState()
				go g.PublishState()
			}
		}

	case protocol.MessageTypePlayerVote:
		if !g.player.IsDealer {
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
			g.currentState.TempVoteResult = make(map[protocol.PlayerID]protocol.VoteResult)
		}
		g.currentState.TempVoteResult[playerVote.VoteBy] = playerVote.VoteResult
		g.notifyChangedState()
		go g.PublishState()

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
	return g.currentState
}

func (g *Game) notifyChangedState() {
	g.logger.Debug("notifying state change",
		zap.Int("subscribers", len(g.stateSubscribers)),
		zap.Any("state", g.currentState),
	)
	for _, subscriber := range g.stateSubscribers {
		subscriber <- g.currentState
	}
}

func (g *Game) publishOnlineState() {
	g.PublishUserOnline(g.player)
	for {
		select {
		case <-time.After(config.OnlineMessagePeriod):
			g.PublishUserOnline(g.player)
		case <-g.leaveSession:
			return
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *Game) publishStatePeriodically() {
	for {
		select {
		case <-time.After(config.StateMessagePeriod):
			g.PublishState()
		case <-g.leaveSession:
			return
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *Game) processIncomingMessages() {
	sub, err := g.transport.SubscribeToMessages(g.session)
	// FIXME: defer unsubscribe

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
		case <-g.leaveSession:
			return
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
	err = g.transport.PublishUnencryptedMessage(g.session, payload)
	if err != nil {
		g.logger.Error("failed to publish message", zap.Error(err))
		return
	}
}

func (g *Game) PublishUserOnline(player *protocol.Player) {
	g.logger.Debug("publishing online state")
	g.publishMessage(protocol.PlayerOnlineMessage{
		Message: protocol.Message{
			Type:      protocol.MessageTypePlayerOnline,
			Timestamp: g.timestamp(),
		},
		Player: *player,
	})
}

func (g *Game) PublishVote(vote int) {
	g.logger.Debug("publishing vote", zap.Int("vote", vote))
	g.publishMessage(protocol.PlayerVote{
		Message: protocol.Message{
			Type:      protocol.MessageTypePlayerVote,
			Timestamp: g.timestamp(),
		},
		VoteBy:     g.player.ID,
		VoteFor:    g.currentState.VoteItem.ID,
		VoteResult: protocol.VoteResult(vote),
	})
}

func (g *Game) PublishState() {
	if !g.player.IsDealer {
		g.logger.Warn("only dealer can publish state")
		return
	}

	g.logger.Debug("publishing state")
	g.publishMessage(protocol.GameStateMessage{
		Message: protocol.Message{
			Type:      protocol.MessageTypeState,
			Timestamp: g.timestamp(),
		},
		State: *g.currentState,
	})
}

func (g *Game) timestamp() int64 {
	return time.Now().UnixMilli()
}

func (g *Game) Deal(input string) error {
	if !g.player.IsDealer {
		return errors.New("only dealer can deal")
	}

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
	g.currentState.TempVoteResult = nil
	g.notifyChangedState()
	go g.PublishState()

	return nil
}

func (g *Game) CreateNewSession() error {
	keyInfo, err := g.generateSymmetricKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate symmetric key")
	}

	info := protocol.BuildSession(keyInfo.SymKey)
	sessionID, err := info.ToSessionID()

	if err != nil {
		return errors.Wrap(err, "failed to marshal session info")
	}

	g.logger.Info("new session created", zap.String("sessionID", sessionID))
	g.player.IsDealer = true
	g.session = info
	g.sessionID = sessionID
	g.currentState = &protocol.State{
		Players: make(map[protocol.PlayerID]protocol.Player),
	}
	g.currentStateTimestamp = g.timestamp()

	return nil
}

func (g *Game) JoinSession(sessionID string) error {
	info, err := protocol.ParseSessionID(sessionID)
	if err != nil {
		return errors.Wrap(err, "failed to parse session ID")
	}

	g.player.IsDealer = false
	g.session = info
	g.sessionID = sessionID
	g.currentState = nil
	g.currentStateTimestamp = g.timestamp()
	g.logger.Info("joined session", zap.String("sessionID", sessionID))

	return nil
}

func (g *Game) IsDealer() bool {
	return g.player.IsDealer
}

func (g *Game) SessionID() string {
	return g.sessionID
}

func (g *Game) Player() protocol.Player {
	return *g.player
}

func (g *Game) RenamePlayer(name string) {
	g.player.Name = name
	g.PublishUserOnline(g.player)
}
