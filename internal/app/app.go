package app

import (
	"2sp/internal/config"
	"2sp/internal/transport"
	"2sp/pkg/game"
	"2sp/pkg/storage"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type App struct {
	Game *game.Game
	waku *transport.Node

	ctx  context.Context
	quit context.CancelFunc
}

func NewApp() *App {
	ctx, quit := context.WithCancel(context.Background())

	return &App{
		Game: nil,
		waku: nil,
		ctx:  ctx,
		quit: quit,
	}
}

func (a *App) Initialize() error {
	var err error
	var localStorage storage.Service

	if !config.Anonymous() {
		localStorage, err = storage.NewLocalStorage("")
		if err != nil {
			return errors.Wrap(err, "failed to create storage")
		}
	}

	a.waku = transport.NewNode(a.ctx, config.Logger)

	err = a.waku.Initialize()
	if err != nil {
		printedErr := errors.New("failed to initialize waku node")
		config.Logger.Error(printedErr.Error(), zap.Error(err))
		return printedErr
	}

	// NOTE: Before transportEventHandler we were subscribing here before starting waku.
	// 		 Hopefully this is covered by "Force notify current status".
	//a.wakuStatusSubscription = a.waku.SubscribeToConnectionStatus()

	err = a.waku.Start()
	if err != nil {
		printedErr := errors.New("failed to start waku node")
		config.Logger.Error(printedErr.Error(), zap.Error(err))
		return printedErr
	}

	a.Game = game.NewGame(a.ctx, a.waku, localStorage)
	err = a.Game.Initialize()
	if err != nil {
		return errors.Wrap(err, "failed to initialize game")
	}

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

// NOTE: temp method
func (a *App) Transport() transport.Service {
	return a.waku
}
