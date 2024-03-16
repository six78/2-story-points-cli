package game

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	wp "github.com/waku-org/go-waku/waku/v2/payload"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"net/url"
	"time"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"
)

type StateSubscription chan *protocol.State

type Game struct {
	logger    *zap.Logger
	ctx       context.Context
	transport Transport
	leaveRoom chan struct{}

	player     *protocol.Player
	playerVote protocol.VoteResult

	room             *protocol.Room
	roomID           protocol.RoomID
	state            *protocol.State
	stateTimestamp   int64
	stateSubscribers []StateSubscription
}

func NewGame(ctx context.Context, transport Transport, playerID protocol.PlayerID) *Game {
	return &Game{
		logger:    config.Logger.Named("game"),
		ctx:       ctx,
		transport: transport,
		leaveRoom: nil,
		player: &protocol.Player{
			ID:       playerID,
			Name:     config.PlayerName(),
			IsDealer: false,
		},
		playerVote: 0,
		room:       nil,
	}
}

func (g *Game) Start() {
	g.leaveRoom = make(chan struct{})

	go g.processIncomingMessages()
	go g.publishOnlineState()

	if g.player.IsDealer {
		go g.publishStatePeriodically()
	}
}

func (g *Game) LeaveRoom() {
	if g.leaveRoom != nil {
		close(g.leaveRoom)
	}
}

func (g *Game) Stop() {
	for _, subscriber := range g.stateSubscribers {
		close(subscriber)
	}
	g.stateSubscribers = nil
	g.LeaveRoom()
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
		if message.Timestamp < g.stateTimestamp {
			logger.Warn("ignoring outdated state message")
			return
		}
		var stateMessage protocol.GameStateMessage
		err := json.Unmarshal(payload, &stateMessage)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.state = &stateMessage.State
		g.notifyChangedState(false)

	case protocol.MessageTypePlayerOnline:
		var playerOnline protocol.PlayerOnlineMessage
		err := json.Unmarshal(payload, &playerOnline)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.logger.Info("player is online", zap.Any("player", playerOnline.Player))
		if g.player.IsDealer {
			if _, ok := g.state.Players[playerOnline.Player.ID]; !ok {
				playerOnline.Player.Order = len(g.state.Players)
				if g.state.Players == nil {
					g.state.Players = make(map[protocol.PlayerID]protocol.Player)
				}
				g.state.Players[playerOnline.Player.ID] = playerOnline.Player
				g.notifyChangedState(true)
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
		if !slices.Contains(g.state.Deck, playerVote.VoteResult) {
			logger.Warn("player vote ignored as not found in deck",
				zap.String("playerID", string(playerVote.VoteBy)),
				zap.Any("vote", playerVote.VoteResult),
				zap.Any("deck", g.state.Deck))
			return
		}
		g.logger.Info("player voted",
			zap.String("name", string(playerVote.VoteBy)),
			zap.String("voteFor", playerVote.VoteFor),
			zap.Int("voteResult", int(playerVote.VoteResult)))

		if g.state.TempVoteResult == nil {
			g.state.TempVoteResult = make(map[protocol.PlayerID]*protocol.VoteResult)
		}
		g.state.TempVoteResult[playerVote.VoteBy] = &playerVote.VoteResult
		g.notifyChangedState(true)

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
	return g.state
}

func (g *Game) notifyChangedState(publish bool) {
	state := g.hiddenCurrentState()

	g.logger.Debug("notifying state change",
		zap.Int("subscribers", len(g.stateSubscribers)),
		zap.Any("state", state),
	)

	for _, subscriber := range g.stateSubscribers {
		subscriber <- state
	}

	if publish {
		go g.publishState(state)
	}
}

func (g *Game) publishOnlineState() {
	g.PublishUserOnline(g.player)
	for {
		select {
		case <-time.After(config.OnlineMessagePeriod):
			g.PublishUserOnline(g.player)
		case <-g.leaveRoom:
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
			g.publishState(g.hiddenCurrentState())
		case <-g.leaveRoom:
			return
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *Game) processIncomingMessages() {
	sub, err := g.transport.SubscribeToMessages(g.room)
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
		case <-g.leaveRoom:
			return
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *Game) publishMessage(message any) {
	if g.room == nil {
		g.logger.Warn("no room to publish message")
		return

	}
	payload, err := json.Marshal(message)
	if err != nil {
		g.logger.Error("failed to marshal message", zap.Error(err))
		return
	}
	err = g.transport.PublishUnencryptedMessage(g.room, payload)
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

func (g *Game) PublishVote(vote protocol.VoteResult) error {
	if g.state.VoteState != protocol.VotingState {
		return errors.New("no voting in progress")
	}
	if !slices.Contains(g.state.Deck, vote) {
		return fmt.Errorf("invalid vote")
	}
	g.logger.Debug("publishing vote", zap.Any("vote", vote))
	g.playerVote = vote
	g.publishMessage(protocol.PlayerVote{
		Message: protocol.Message{
			Type:      protocol.MessageTypePlayerVote,
			Timestamp: g.timestamp(),
		},
		VoteBy:     g.player.ID,
		VoteFor:    g.state.VoteItem.ID,
		VoteResult: vote,
	})
	return nil
}

func (g *Game) publishState(state *protocol.State) {
	if !g.player.IsDealer {
		g.logger.Warn("only dealer can publish state")
		return
	}

	if state == nil {
		g.logger.Error("no state to publish")
		return
	}

	g.logger.Debug("publishing state")
	g.publishMessage(protocol.GameStateMessage{
		Message: protocol.Message{
			Type:      protocol.MessageTypeState,
			Timestamp: g.timestamp(),
		},
		State: *state,
	})
}

func (g *Game) timestamp() int64 {
	return time.Now().UnixMilli()
}

func (g *Game) Deal(input string) error {
	if !g.player.IsDealer {
		return errors.New("only dealer can deal")
	}

	if g.state.VoteState != protocol.IdleState {
		return errors.New("cannot deal when voting is in progress")
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

	g.state.VoteItem = item
	g.state.TempVoteResult = nil
	g.state.VoteState = protocol.VotingState
	g.notifyChangedState(true)

	return nil
}

func (g *Game) CreateNewRoom() error {
	keyInfo, err := g.generateSymmetricKey()
	if err != nil {
		return errors.Wrap(err, "failed to generate symmetric key")
	}

	info := protocol.BuildRoom(keyInfo.SymKey)
	roomID, err := info.ToRoomID()

	if err != nil {
		return errors.Wrap(err, "failed to marshal room info")
	}

	g.logger.Info("new room created", zap.String("roomID", roomID.String()))
	g.player.IsDealer = true
	g.room = info
	g.roomID = roomID

	deck, _ := GetDeck(Fibonacci)
	g.state = &protocol.State{
		Players:   make(map[protocol.PlayerID]protocol.Player),
		VoteState: protocol.IdleState,
		Deck:      deck,
	}
	g.stateTimestamp = g.timestamp()

	return nil
}

func (g *Game) JoinRoom(roomID string) error {
	info, err := protocol.ParseRoomID(roomID)
	if err != nil {
		return errors.Wrap(err, "failed to parse room ID")
	}

	g.player.IsDealer = false
	g.room = info
	g.roomID = protocol.NewRoomID(roomID)
	g.state = nil
	g.stateTimestamp = g.timestamp()
	g.logger.Info("joined room", zap.String("roomID", roomID))

	return nil
}

func (g *Game) IsDealer() bool {
	return g.player.IsDealer
}

func (g *Game) RoomID() string {
	return g.roomID.String()
}

func (g *Game) Player() protocol.Player {
	return *g.player
}

func (g *Game) PlayerVote() protocol.VoteResult {
	return g.playerVote
}

func (g *Game) RenamePlayer(name string) {
	g.player.Name = name
	g.PublishUserOnline(g.player)
}

func (g *Game) Reveal() error {
	if !g.player.IsDealer {
		return errors.New("only dealer can reveal cards")
	}

	if g.state.VoteState != protocol.VotingState {
		return errors.New("cannot reveal when voting is not in progress")
	}

	g.state.VoteState = protocol.RevealedState
	g.notifyChangedState(true)
	return nil
}

func (g *Game) hiddenCurrentState() *protocol.State {
	if g.state == nil {
		return nil
	}

	// Create a deep copy of the state
	hiddenState := *g.state

	// Manually copy the map, otherwise it's a modifiable reference
	hiddenState.TempVoteResult = make(map[protocol.PlayerID]*protocol.VoteResult, len(g.state.TempVoteResult))
	for playerID, vote := range g.state.TempVoteResult {
		if hiddenState.VoteState == protocol.VotingState {
			hiddenState.TempVoteResult[playerID] = nil
		} else {
			hiddenState.TempVoteResult[playerID] = vote
		}
	}

	return &hiddenState
}

func (g *Game) SetDeck(deck protocol.Deck) error {
	if !g.player.IsDealer {
		return errors.New("only dealer can set deck")
	}
	if g.state.VoteState != protocol.IdleState && g.state.VoteState != protocol.FinishedState {
		return errors.New("cannot set deck when voting is in progress")
	}
	g.state.Deck = deck
	g.notifyChangedState(true)
	return nil
}
