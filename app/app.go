package app

import (
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"waku-poker-planning/config"
	"waku-poker-planning/game"
	"waku-poker-planning/protocol"
	"waku-poker-planning/waku"
)

type App struct {
	Game    *game.Game
	waku    *waku.Node
	storage *Storage

	ctx  context.Context
	quit context.CancelFunc

	gameStateSubscription  game.StateSubscription
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

func (a *App) GameState() *protocol.State {
	if a.Game == nil {
		return &protocol.State{}
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

	playerID := a.storage.PlayerID()
	if config.Anonymous() {
		playerID, err = game.GeneratePlayerID()
		if err != nil {
			return errors.Wrap(err, "failed to generate player ID")
		}
	}

	playerName := a.storage.PlayerName()
	if config.PlayerName() != "" {
		playerName = config.PlayerName()
	}

	a.Game = game.NewGame(a.ctx, a.waku, playerID, playerName)
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

func (a *App) WaitForGameState() (*protocol.State, bool, error) {
	if a.gameStateSubscription == nil {
		config.Logger.Error("Game state subscription not created")
		return &protocol.State{}, false, errors.New("Game state subscription not created")
	}

	state, more := <-a.gameStateSubscription
	if !more {
		a.gameStateSubscription = nil
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

func (a *App) IsDealer() bool {
	return a.Game != nil && a.Game.IsDealer()
}

func (a *App) RenamePlayer(name string) error {
	a.Game.RenamePlayer(name)
	if config.Anonymous() {
		return nil
	}
	return a.storage.SetPlayerName(name)
}
