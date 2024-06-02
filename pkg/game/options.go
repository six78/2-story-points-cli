package game

import (
	"2sp/internal/transport"
	"2sp/pkg/storage"
	"context"
	"go.uber.org/zap"
	"time"
)

type Option func(*Game)

func WithContext(ctx context.Context) Option {
	return func(g *Game) {
		g.ctx = ctx
	}
}

func WithTransport(t transport.Service) Option {
	return func(g *Game) {
		g.transport = t
	}
}

func WithLogger(l *zap.Logger) Option {
	return func(g *Game) {
		g.logger = l
	}
}

func WithStorage(s storage.Service) Option {
	return func(g *Game) {
		g.storage = s
	}
}

func WithEnableSymmetricEncryption(b bool) Option {
	return func(g *Game) {
		g.config.EnableSymmetricEncryption = b
	}
}

func WithPlayerName(name string) Option {
	return func(g *Game) {
		g.config.PlayerName = name
	}
}

func WithOnlineMessagePeriod(d time.Duration) Option {
	return func(g *Game) {
		g.config.OnlineMessagePeriod = d
	}
}

func WithStateMessagePeriod(d time.Duration) Option {
	return func(g *Game) {
		g.config.StateMessagePeriod = d
	}
}
