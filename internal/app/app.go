package app

import (
	"2sp/internal/config"
	"2sp/internal/waku"
	"2sp/pkg/game"
	"2sp/pkg/storage"
	"context"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type App struct {
	Game    *game.Game
	waku    *waku.Node
	storage *storage.LocalStorage

	ctx  context.Context
	quit context.CancelFunc

	wakuStatusSubscription waku.ConnectionStatusSubscription
}

func NewApp() *App {
	ctx, quit := context.WithCancel(context.Background())

	return &App{
		Game:    nil,
		waku:    nil,
		storage: nil,
		ctx:     ctx,
		quit:    quit,
	}
}

func (a *App) Initialize() error {
	var err error

	if !config.Anonymous() {
		a.storage, err = storage.NewStorage("")
		if err != nil {
			return errors.Wrap(err, "failed to create storage")
		}
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

	a.Game, err = game.NewGame(a.ctx, a.waku, a.storage)
	if err != nil {
		return err
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
