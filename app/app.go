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

type State int

const (
	Idle State = iota
	Initializing
	WaitingForPeers
	UserInput
	CreatingSession
	JoiningSession
)

type App struct {
	Game *game.Game

	waku *waku.Node
	ctx  context.Context
	quit context.CancelFunc

	gameStateSubscription game.StateSubscription
}

func NewApp() *App {
	ctx, quit := context.WithCancel(context.Background())

	return &App{
		waku:                  nil,
		Game:                  nil,
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
	w, err := waku.NewNode(a.ctx, config.Logger)
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
	a.Game = game.NewGame(a.ctx, a.waku)
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

func (a *App) StartGame() {
	a.Game.Start()
}

func (a *App) WaitForPeersConnected() bool {
	if a.waku == nil {
		config.Logger.Error("waku node not created")
		return false
	}

	return a.waku.WaitForPeersConnected()
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

func (a *App) Deal(input string) error {
	if a.Game == nil {
		return errors.New("Game not created")
	}

	return a.Game.Deal(input)
}

func (a *App) CreateNewSession() error {
	if a.Game == nil {
		return errors.New("Game not created")
	}

	a.Game.LeaveSession()

	err := a.Game.CreateNewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create new session")
	}

	a.Game.Start()
	return nil
}

func (a *App) JoinSession(sessionID string) error {
	if a.Game == nil {
		return errors.New("Game not created")
	}

	if a.Game.SessionID() == sessionID {
		return errors.New("already in this session")
	}

	a.Game.LeaveSession()

	err := a.Game.JoinSession(sessionID)
	if err != nil {
		return errors.Wrap(err, "failed to join session")
	}

	a.Game.Start()
	return nil
}

func (a *App) IsDealer() bool {
	return a.Game != nil && a.Game.IsDealer()
}

func (a *App) GameSessionID() string {
	return a.Game.SessionID()
}

func (a *App) RenamePlayer(name string) {
	if a.Game == nil {
		config.Logger.Error("Game not created")
		return
	}

	a.Game.RenamePlayer(name)
}
