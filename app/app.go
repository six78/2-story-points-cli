package app

import (
	"errors"
	"go.uber.org/zap"
	"waku-poker-planning/config"
	"waku-poker-planning/game"
	"waku-poker-planning/protocol"
	"waku-poker-planning/waku"
)

type State int

const (
	Idle State = iota
	Initializing
	WaitingForPeers
	Playing
)

type App struct {
	waku *waku.Node
	game *game.Game

	gameStateSubscription chan protocol.State
}

func NewApp() *App {
	return &App{
		waku:                  nil,
		game:                  nil,
		gameStateSubscription: nil,
	}
}

func (a *App) GameState() protocol.State {
	if a.game == nil {
		return protocol.State{}
	}
	return a.game.CurrentState()
}

func (a *App) Initialize() error {
	w, err := waku.NewNode(config.Logger)
	if err != nil {
		printedErr := errors.New("failed to create waku node")
		config.Logger.Error(printedErr.Error(), zap.Error(err))
		return printedErr
	}

	err = w.Start()
	if err != nil {
		printedErr := errors.New("failed to start waku node")
		config.Logger.Error(printedErr.Error(), zap.Error(err))
		return printedErr
	}

	a.waku = w
	a.game = game.NewGame(a.waku)
	a.gameStateSubscription = a.game.SubscribeToStateChanges()

	return nil
}

func (a *App) Stop() {
	if a.game != nil {
		a.game.Stop()
	}
	if a.waku != nil {
		a.waku.Stop()
	}
}

func (a *App) StartGame() {
	a.game.Start()
}

func (a *App) WaitForPeersConnected() bool {
	if a.waku == nil {
		config.Logger.Error("waku node not created")
		return false
	}

	return a.waku.WaitForPeersConnected()
}

func (a *App) WaitForGameState() (protocol.State, bool) {
	if a.gameStateSubscription == nil {
		config.Logger.Error("game state subscription not created")
		return protocol.State{}, false
	}

	state, more := <-a.gameStateSubscription
	if !more {
		a.gameStateSubscription = nil
	}
	return state, more
}

// PublishUserOnline should be used for testing purposes only
func (a *App) PublishUserOnline(username string) {
	if a.game == nil {
		config.Logger.Error("game not created")
		return
	}

	a.game.PublishUserOnline(username)
}
