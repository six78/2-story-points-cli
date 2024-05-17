package game

import (
	"2sp/config"
	"2sp/protocol"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"time"
)

type StateSubscription chan *protocol.State

type Game struct {
	logger    *zap.Logger
	ctx       context.Context
	transport Transport
	exitRoom  chan struct{}
	features  FeatureFlags

	isDealer bool
	player   *protocol.Player
	myVote   protocol.VoteResult // We save our vote to show it in UI

	room             *protocol.Room
	roomID           protocol.RoomID
	state            *protocol.State
	stateTimestamp   int64
	stateSubscribers []StateSubscription
}

func NewGame(ctx context.Context, transport Transport, player *protocol.Player) *Game {
	return &Game{
		logger:    config.Logger.Named("game"),
		ctx:       ctx,
		transport: transport,
		exitRoom:  nil,
		features:  defaultFeatureFlags(),
		isDealer:  false,
		player: &protocol.Player{
			ID:     player.ID,
			Name:   player.Name,
			Online: true,
		},
		myVote: protocol.VoteResult{
			Value:     "",
			Timestamp: 0,
		},
		room:           nil,
		stateTimestamp: 0,
	}
}

func (g *Game) LeaveRoom() {
	if g.room != nil {
		g.publishUserOffline()
	}

	if g.exitRoom != nil {
		close(g.exitRoom)
	}

	g.logger.Info("left room", zap.String("roomID", g.roomID.String()))

	g.exitRoom = nil
	g.isDealer = false
	g.room = nil
	g.roomID = protocol.NewRoomID("")
	g.state = nil
	g.stateTimestamp = 0
	g.notifyChangedState(false)
}

func (g *Game) Stop() {
	for _, subscriber := range g.stateSubscribers {
		close(subscriber)
	}
	g.stateSubscribers = nil
	g.LeaveRoom()
	// WARNING: wait for all routines to finish
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
		if g.isDealer {
			return
		}
		g.logger.Debug("state message received",
			zap.Int64("timestamp", message.Timestamp),
			zap.Int64("localTimestamp", g.stateTimestamp),
		)
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
		if g.state != nil && stateMessage.State.ActiveIssue != g.state.ActiveIssue {
			// Voting finished or new issue dealt. Reset our vote.
			g.resetMyVote()
		}
		g.state = &stateMessage.State
		g.state.Deck, _ = GetDeck(Fibonacci) // FIXME: remove hardcoded deck
		g.notifyChangedState(false)

	case protocol.MessageTypePlayerOnline:
		if !g.isDealer {
			return
		}
		var playerOnline protocol.PlayerOnlineMessage
		err := json.Unmarshal(payload, &playerOnline)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.logger.Info("player online message received", zap.Any("player", playerOnline.Player))

		// TODO: Store player pointers in a map

		index := g.playerIndex(playerOnline.Player.ID)
		if index < 0 {
			playerOnline.Player.Online = true
			playerOnline.Player.OnlineTimestamp = time.Now()
			g.state.Players = append(g.state.Players, playerOnline.Player)
			g.notifyChangedState(true)
			return
		}

		playerChanged := !g.state.Players[index].Online ||
			g.state.Players[index].Name != playerOnline.Player.Name

		if !playerChanged {
			return
		}

		g.state.Players[index].Online = true
		g.state.Players[index].OnlineTimestamp = time.Now()
		g.state.Players[index].Name = playerOnline.Player.Name
		g.notifyChangedState(true)

	case protocol.MessageTypePlayerOffline:
		if !g.isDealer {
			return
		}
		var playerOffline protocol.PlayerOfflineMessage
		err := json.Unmarshal(payload, &playerOffline)
		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}
		g.logger.Info("player is offline", zap.Any("player", playerOffline.Player))

		index := g.playerIndex(playerOffline.Player.ID)
		if index < 0 {
			return
		}

		g.state.Players[index].Online = false
		g.notifyChangedState(true)

	case protocol.MessageTypePlayerVote:
		if !g.isDealer {
			return
		}
		var playerVote protocol.PlayerVoteMessage
		err := json.Unmarshal(payload, &playerVote)

		if err != nil {
			logger.Error("failed to unmarshal message", zap.Error(err))
			return
		}

		if g.state.VoteState() != protocol.VotingState {
			g.logger.Warn("player vote ignored as not in voting state",
				zap.Any("playerID", playerVote.PlayerID),
			)
			return
		}

		if playerVote.VoteResult.Value != "" && !slices.Contains(g.state.Deck, playerVote.VoteResult.Value) {
			logger.Warn("player vote ignored as not found in deck",
				zap.Any("playerID", playerVote.PlayerID),
				zap.Any("vote", playerVote.VoteResult),
				zap.Any("deck", g.state.Deck))
			return
		}

		if g.state.ActiveIssue != playerVote.Issue {
			g.logger.Warn("player vote ignored as not for the current vote item",
				zap.Any("playerID", playerVote.PlayerID),
				zap.Any("voteFor", playerVote.Issue),
				zap.Any("currentVoteItemID", g.state.ActiveIssue),
			)
			return
		}

		item := g.state.Issues.Get(playerVote.Issue)
		if item == nil {
			logger.Error("vote item not found", zap.Any("voteFor", playerVote.Issue))
			return
		}

		currentVote, voteExist := item.Votes[playerVote.PlayerID]
		if voteExist && currentVote.Timestamp >= playerVote.Timestamp {
			logger.Warn("player vote ignored as outdated",
				zap.Any("playerID", playerVote.PlayerID),
				zap.Any("currentVote", currentVote),
				zap.Any("receivedVote", playerVote.VoteResult),
			)
			return
		}

		g.logger.Info("player vote accepted",
			zap.String("name", string(playerVote.PlayerID)),
			zap.String("voteFor", string(playerVote.Issue)),
			zap.String("voteResult", string(playerVote.VoteResult.Value)),
			zap.Any("timestamp", playerVote.Timestamp),
		)

		if playerVote.VoteResult.Value == "" {
			delete(item.Votes, playerVote.PlayerID)
		} else {
			item.Votes[playerVote.PlayerID] = playerVote.VoteResult
		}

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
		zap.Bool("publish", publish),
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
	g.publishUserOnline()
	for {
		select {
		case <-time.After(config.OnlineMessagePeriod):
			g.publishUserOnline()
		case <-g.exitRoom:
			return
		case <-g.ctx.Done():
			return
		}
	}
}

func (g *Game) publishStateLoop() {
	logger := g.logger.With(zap.String("source", "state publish loop"))
	logger.Debug("started")
	for {
		select {
		case <-time.After(config.StateMessagePeriod):
			logger.Debug("tick")
			g.notifyChangedState(true)
		case <-g.exitRoom:
			logger.Debug("finished: room left")
			return
		case <-g.ctx.Done():
			logger.Debug("finished: ctx done")
			return
		}
	}
}

func (g *Game) checkPlayersStateLoop() {
	g.logger.Debug("check users state loop")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-g.exitRoom:
			return
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			if g.state == nil {
				continue
			}
			now := time.Now()
			for i, player := range g.state.Players {
				diff := now.Sub(player.OnlineTimestamp)
				if diff > 20*time.Second {
					g.state.Players[i].Online = false
				}
			}
		}
	}
}

func (g *Game) processIncomingMessages(sub *MessagesSubscription) {
	if sub.Unsubscribe != nil {
		defer sub.Unsubscribe()
	}
	for {
		select {
		case payload, more := <-sub.Ch:
			if !more {
				return
			}
			g.processMessage(payload)
		case <-g.exitRoom:
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

	if config.EnableSymmetricEncryption {
		err = g.transport.PublishPublicMessage(g.room, payload)
	} else {
		err = g.transport.PublishUnencryptedMessage(g.room, payload)
	}

	if err != nil {
		g.logger.Error("failed to publish message", zap.Error(err))
		return
	}
}

func (g *Game) publishUserOnline() {
	g.logger.Debug("publishing online state")
	g.publishMessage(protocol.PlayerOnlineMessage{
		Message: protocol.Message{
			Type:      protocol.MessageTypePlayerOnline,
			Timestamp: g.timestamp(),
		},
		Player: *g.player,
	})
}

func (g *Game) publishUserOffline() {
	g.logger.Debug("publishing offline state")
	g.publishMessage(protocol.PlayerOfflineMessage{
		Message: protocol.Message{
			Type:      protocol.MessageTypePlayerOffline,
			Timestamp: g.timestamp(),
		},
		Player: *g.player,
	})
}

func (g *Game) PublishVote(vote protocol.VoteValue) error {
	if g.state.VoteState() != protocol.VotingState {
		return errors.New("no voting in progress")
	}
	if vote != "" && !slices.Contains(g.state.Deck, vote) {
		return fmt.Errorf("invalid vote")
	}
	g.logger.Debug("publishing vote", zap.Any("vote", vote))
	g.myVote = *protocol.NewVoteResult(vote)
	g.publishMessage(protocol.PlayerVoteMessage{
		Message: protocol.Message{
			Type:      protocol.MessageTypePlayerVote,
			Timestamp: g.timestamp(),
		},
		PlayerID:   g.player.ID,
		Issue:      g.state.ActiveIssue,
		VoteResult: g.myVote,
	})
	return nil
}

func (g *Game) RetrieveVote() error {
	return g.PublishVote("")
}

func (g *Game) publishState(state *protocol.State) {
	if !g.isDealer {
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

func (g *Game) Deal(input string) (protocol.IssueID, error) {
	if !g.isDealer {
		return "", errors.New("only dealer can deal")
	}

	if g.state.VoteState() == protocol.RevealedState {
		return "", errors.New("finish current vote to deal another issue")
	}

	issueID, err := g.addIssue(input)
	if err != nil {
		return "", errors.Wrap(err, "failed to add issue")
	}

	err = g.SelectIssue(len(g.state.Issues) - 1)

	return issueID, err
}

func (g *Game) CreateNewRoom() (*protocol.Room, *protocol.State, error) {
	room, err := protocol.NewRoom()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create a new room")
	}

	_, err = room.ToRoomID()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to convert to room id")
	}

	deckName := Fibonacci // FIXME: Remove hardcoded deck
	deck, deckFound := GetDeck(deckName)
	if !deckFound {
		return nil, nil, errors.Wrap(err, fmt.Sprintf("unknown deck '%s'", deckName))
	}

	state := &protocol.State{
		Players:     []protocol.Player{*g.player},
		Deck:        deck,
		ActiveIssue: "",
		Issues:      make([]*protocol.Issue, 0),
		Timestamp:   g.timestamp(),
	}

	return room, state, nil
}

func (g *Game) JoinRoom(roomID protocol.RoomID, state *protocol.State) error {
	if g.RoomID() == roomID {
		return errors.New("already in this room")
	}

	if g.room != nil {
		return errors.New("exit current room to join another one")
	}

	if roomID.Empty() {
		return errors.New("empty room ID")
	}

	room, err := protocol.ParseRoomID(roomID.String())
	if err != nil {
		return errors.Wrap(err, "failed to parse room ID")
	}
	if !room.VersionSupported() {
		return errors.Wrap(err, fmt.Sprintf("this room has unsupported version %d", room.Version))
	}

	g.exitRoom = make(chan struct{})

	sub, err := g.transport.SubscribeToMessages(room)
	if err != nil {
		return errors.Wrap(err, "failed to subscribe to messages")
	}

	g.isDealer = state != nil
	g.room = room
	g.roomID = roomID
	g.state = state
	g.stateTimestamp = 0
	if g.isDealer {
		g.state.Deck, _ = GetDeck(Fibonacci) // FIXME: remove hardcoded deck
	}
	g.resetMyVote()

	go g.processIncomingMessages(sub)
	go g.publishOnlineState()
	if g.isDealer {
		go g.publishStateLoop()
		go g.checkPlayersStateLoop()
	}
	g.notifyChangedState(g.isDealer)

	if state == nil {
		g.logger.Info("joined room", zap.Any("roomID", roomID))

	} else {
		g.stateTimestamp = g.timestamp()
		g.logger.Info("loaded room",
			zap.Any("roomID", roomID),
			zap.Bool("isDealer", g.isDealer))
	}

	return nil
}

func (g *Game) IsDealer() bool {
	return g.isDealer
}

func (g *Game) Room() protocol.Room {
	return *g.room
}

func (g *Game) RoomID() protocol.RoomID {
	return g.roomID
}

func (g *Game) Player() protocol.Player {
	return *g.player
}

func (g *Game) MyVote() protocol.VoteResult {
	return g.myVote
}

func (g *Game) RenamePlayer(name string) {
	g.player.Name = name
	g.publishUserOnline()
}

func (g *Game) Reveal() error {
	if !g.isDealer {
		return errors.New("only dealer can reveal cards")
	}

	if g.state.VoteState() != protocol.VotingState {
		return errors.New("cannot reveal when voting is not in progress")
	}

	g.state.VotesRevealed = true
	g.notifyChangedState(true)
	return nil
}

func (g *Game) hiddenCurrentState() *protocol.State {
	if g.state == nil {
		return nil
	}

	// Create a deep copy of the state
	hiddenState := *g.state

	if hiddenState.VoteState() != protocol.VotingState {
		return &hiddenState
	}

	hiddenState.Issues = make([]*protocol.Issue, 0, len(g.state.Issues))
	for _, item := range g.state.Issues {
		copiedItem := *item
		copiedItem.Votes = make(map[protocol.PlayerID]protocol.VoteResult, len(item.Votes))
		for playerID, vote := range item.Votes {
			if item.ID == g.state.ActiveIssue {
				copiedItem.Votes[playerID] = vote.Hidden()
			} else {
				copiedItem.Votes[playerID] = vote
			}
		}
		hiddenState.Issues = append(hiddenState.Issues, &copiedItem)
	}

	return &hiddenState
}

func (g *Game) SetDeck(deck protocol.Deck) error {
	if !g.features.EnableDeckSelection {
		return errors.New("deck selection is disabled")
	}
	if !g.isDealer {
		return errors.New("only dealer can set deck")
	}
	if g.state.VoteState() != protocol.IdleState && g.state.VoteState() != protocol.FinishedState {
		return errors.New("cannot set deck when voting is in progress")
	}
	g.state.Deck = deck
	g.notifyChangedState(true)
	return nil
}

func (g *Game) Finish(result protocol.VoteValue) error {
	if !g.isDealer {
		return errors.New("only dealer can finish")
	}
	if g.state.VoteState() != protocol.RevealedState {
		return errors.New("cannot finish when voting is not revealed")
	}
	if !slices.Contains(g.state.Deck, result) {
		return errors.New("result is not in the deck")
	}

	item := g.state.Issues.Get(g.state.ActiveIssue)
	if item == nil {
		return errors.New("vote item not found in the vote list")
	}

	item.Result = &result
	g.state.ActiveIssue = g.state.Issues.GetNextIssueToDeal(g.state.ActiveIssue)
	g.state.VotesRevealed = false
	g.resetMyVote()
	g.notifyChangedState(true)

	return nil
}

func (g *Game) resetMyVote() {
	g.myVote = protocol.VoteResult{
		Value:     "",
		Timestamp: 0,
	}
}

func (g *Game) AddIssue(titleOrURL string) (protocol.IssueID, error) {
	if !g.isDealer {
		return "", errors.New("only dealer can add issues")
	}
	issueID, err := g.addIssue(titleOrURL)
	if err != nil {
		return "", err
	}
	g.notifyChangedState(true)
	return issueID, nil
}

func (g *Game) addIssue(titleOrURL string) (protocol.IssueID, error) {
	issueID, err := GenerateIssueID()
	if err != nil {
		return "", errors.New("failed to generate UUID")
	}

	issueExist := slices.ContainsFunc(g.state.Issues, func(item *protocol.Issue) bool {
		return item.TitleOrURL == titleOrURL
	})
	if issueExist {
		return "", errors.New("issue already exists")
	}

	g.logger.Debug("adding issue", zap.String("titleOrUrl", titleOrURL))
	issue := protocol.Issue{
		ID:         issueID,
		TitleOrURL: titleOrURL,
		Votes:      make(map[protocol.PlayerID]protocol.VoteResult),
		Result:     nil,
	}

	g.state.Issues = append(g.state.Issues, &issue)
	return issue.ID, nil
}

func (g *Game) SelectIssue(index int) error {
	if !g.isDealer {
		return errors.New("only dealer can deal")
	}

	if g.state.VoteState() == protocol.RevealedState {
		return errors.New("cannot deal when voting is in progress")
	}

	if index < 0 || index >= len(g.state.Issues) {
		return errors.New("invalid issue index")
	}

	g.state.Issues[index].Result = nil
	g.state.Issues[index].Votes = make(map[protocol.PlayerID]protocol.VoteResult)
	g.state.ActiveIssue = g.state.Issues[index].ID
	g.notifyChangedState(true)

	return nil
}

func (g *Game) playerIndex(playerID protocol.PlayerID) int {
	return slices.IndexFunc(g.state.Players, func(player protocol.Player) bool {
		return player.ID == playerID
	})
}
