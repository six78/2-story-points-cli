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
	waku *waku.Node
	game *game.Game

	ctx  context.Context
	quit context.CancelFunc

	gameStateSubscription game.StateSubscription
}

func NewApp() *App {
	ctx, quit := context.WithCancel(context.Background())

	return &App{
		waku:                  nil,
		game:                  nil,
		ctx:                   ctx,
		quit:                  quit,
		gameStateSubscription: nil,
	}
}

func (a *App) GameState() *protocol.State {
	if a.game == nil {
		return &protocol.State{}
	}
	return a.game.CurrentState()
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
	a.game = game.NewGame(a.ctx, a.waku)
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
	a.quit()
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

func (a *App) WaitForGameState() (*protocol.State, bool, error) {
	if a.gameStateSubscription == nil {
		config.Logger.Error("game state subscription not created")
		return &protocol.State{}, false, errors.New("game state subscription not created")
	}

	state, more := <-a.gameStateSubscription
	if !more {
		a.gameStateSubscription = nil
	}
	return state, more, nil
}

// PublishUserOnline should be used for testing purposes only
func (a *App) PublishUserOnline(username string) {
	if a.game == nil {
		config.Logger.Error("game not created")
		return
	}

	a.game.PublishUserOnline(username)
}

func (a *App) PublishVote(vote int) {
	if a.game == nil {
		config.Logger.Error("game not created")
		return
	}

	a.game.PublishVote(vote)
}

func (a *App) Deal(input string) error {
	if a.game == nil {
		return errors.New("game not created")
	}

	return a.game.Deal(input)
}

func (a *App) CreateNewSession() error {
	if a.game == nil {
		return errors.New("game not created")
	}

	a.game.LeaveSession()

	err := a.game.CreateNewSession()
	if err != nil {
		return errors.Wrap(err, "failed to create new session")
	}

	a.game.Start()
	return nil
}

func (a *App) JoinSession(sessionID string) error {
	if a.game == nil {
		return errors.New("game not created")
	}

	if a.game.SessionID() == sessionID {
		return errors.New("already in this session")
	}

	a.game.LeaveSession()

	err := a.game.JoinSession(sessionID)
	if err != nil {
		return errors.Wrap(err, "failed to join session")
	}

	a.game.Start()
	return nil
}

func (a *App) IsDealer() bool {
	return a.game != nil && a.game.IsDealer()
}

func (a *App) GameSessionID() string {
	return a.game.SessionID()
}
