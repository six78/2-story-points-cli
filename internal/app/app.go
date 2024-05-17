package app

import (
	"2sp/internal/config"
	"2sp/internal/waku"
	game2 "2sp/pkg/game"
	protocol2 "2sp/pkg/protocol"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type App struct {
	Game    *game2.Game
	waku    *waku.Node
	storage *Storage

	ctx  context.Context
	quit context.CancelFunc

	gameStateSubscription  game2.StateSubscription
	wakuStatusSubscription waku.ConnectionStatusSubscription
}

func NewApp() *App {
	ctx, quit := context.WithCancel(context.Background())

	return &App{
		Game:                  nil,
		waku:                  nil,
		storage:               nil,
		ctx:                   ctx,
		quit:                  quit,
		gameStateSubscription: nil,
	}
}

func (a *App) GameState() *protocol2.State {
	if a.Game == nil {
		return &protocol2.State{}
	}
	return a.Game.CurrentState()
}

func (a *App) Initialize() error {
	var err error
	a.storage, err = NewStorage()

	if err != nil {
		return errors.Wrap(err, "failed to create storage")
	}

	a.waku, err = waku.NewNode(a.ctx, config.Logger)
	if err != nil {
		printedErr := errors.New("failed to create waku node")
		config.Logger.Error(printedErr.Error(), zap.Error(err))
		return printedErr
	}

	a.wakuStatusSubscription = a.waku.SubscribeToConnectionStatus()

	err = a.waku.Start()
	if err != nil {
		printedErr := errors.New("failed to start waku node")
		config.Logger.Error(printedErr.Error(), zap.Error(err))
		return printedErr
	}

	player, err := a.loadPlayer()
	if err != nil {
		return err
	}

	a.Game = game2.NewGame(a.ctx, a.waku, player)
	//a.Game.RenamePlayer(a.storage.GetPlayerName())
	a.gameStateSubscription = a.Game.SubscribeToStateChanges()

	return nil
}

func (a *App) Stop() {
	if a.Game != nil {
		a.Game.Stop()
	}
	if a.waku != nil {
		a.waku.Stop()
	}
	a.quit()
}

func (a *App) WaitForGameState() (*protocol2.State, bool, error) {
	if a.gameStateSubscription == nil {
		config.Logger.Error("Game state subscription not created")
		return &protocol2.State{}, false, errors.New("Game state subscription not created")
	}

	state, more := <-a.gameStateSubscription
	if !more {
		a.gameStateSubscription = nil
	}

	if !config.Anonymous() && a.Game.IsDealer() { // Only store room state for non-anonymously joined rooms
		err := a.storage.SaveRoomState(a.Game.RoomID(), state)
		if err != nil {
			config.Logger.Error("failed to save room state", zap.Error(err))
		}
	}

	return state, more, nil
}

func (a *App) WaitForConnectionStatus() (waku.ConnectionStatus, bool, error) {
	if a.wakuStatusSubscription == nil {
		config.Logger.Error("Waku connection status subscription not created")
		return waku.ConnectionStatus{}, false, errors.New("Waku connection status subscription not created")
	}

	status, more := <-a.wakuStatusSubscription
	if !more {
		a.wakuStatusSubscription = nil
	}
	return status, more, nil
}

func (a *App) RenamePlayer(name string) error {
	a.Game.RenamePlayer(name)
	if config.Anonymous() {
		return nil
	}
	return a.storage.SetPlayerName(name)
}

func (a *App) LoadRoomState(roomID protocol2.RoomID) (*protocol2.State, error) {
	if config.Anonymous() {
		return nil, nil
	}
	return a.storage.LoadRoomState(roomID)
}

func (a *App) loadPlayer() (*protocol2.Player, error) {
	var err error
	var player protocol2.Player

	// Load ID
	if !config.Anonymous() {
		player.ID = a.storage.PlayerID()
	} else {
		player.ID, err = game2.GeneratePlayerID()
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate player ID")
		}
	}

	// Load Name
	if config.PlayerName() != "" {
		player.Name = config.PlayerName()
	} else if !config.Anonymous() {
		player.Name = a.storage.PlayerName()
	}

	return &player, nil
}
