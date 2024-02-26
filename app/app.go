package app

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/waku-org/go-waku/waku/v2/node"
	"github.com/waku-org/go-waku/waku/v2/protocol"
	"github.com/waku-org/go-waku/waku/v2/protocol/relay"
	"go.uber.org/zap"
)

type App struct {
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc

	waku                 *node.WakuNode
	wakuConnectionStatus chan node.ConnStatus

	session *Session
}

func NewApp() (*App, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, errors.Wrap(err, "failed to configure logging")
	}

	ctx, cancel := context.WithCancel(context.Background())
	waku, status, err := createWakuNode()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create waku node")
	}

	return &App{
		logger:               logger,
		ctx:                  ctx,
		cancel:               cancel,
		waku:                 waku,
		wakuConnectionStatus: status,
	}, nil
}

func (a *App) Start() error {
	err := a.waku.Start(a.ctx)
	if err != nil {
		return errors.Wrap(err, "failed to start waku node")
	}

	a.logger.Info("waku started", zap.String("peerID", a.waku.ID()))

	err = a.discoverNodes()
	if err != nil {
		return errors.Wrap(err, "failed to discover nodes")
	}

	go a.watchConnectionStatus()
	go a.receiveMessages()
	go a.readUserInput()
}

func (a *App) CreateNewSession(name string) error {
	session, err := NewSession(true, name)
	if err != nil {
		return errors.Wrap(err, "failed to create new session")
	}
	a.session = session
	// WARNING: start app here?
}

func (a *App) ConnectToSession(name string) error {
	session, err := NewSession(false, name)
	if err != nil {
		return errors.Wrap(err, "failed to connect to session")
	}
	a.session = session
	// WARNING: start app here?
}

func (a *App) Stop() {
	a.cancel()
	a.waku.Stop()
}

func (a *App) receiveMessages() {
	contentFilter := protocol.NewContentFilter(relay.DefaultWakuTopic, a.session.contentTopic)
	subs, err := a.waku.Relay().Subscribe(a.ctx, contentFilter)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(subs) != 1 {
		a.logger.Error("unexpected number of subscriptions: ", zap.Int("subs", len(subs)))
	}

	for value := range subs[0].Ch {
		a.logger.Info("<<< MESSAGE RECEIVED",
			zap.String("text", string(value.Message().Payload)),
			zap.String("string", value.Message().String()),
		)
	}
}
