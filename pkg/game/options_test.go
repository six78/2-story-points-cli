package game

import (
	mocktransport "2sp/internal/transport/mock"
	mockstorage "2sp/pkg/storage/mock"
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	ctx := context.WithValue(context.Background(), gofakeit.LetterN(3), gofakeit.LetterN(3))
	transport := &mocktransport.MockService{}
	storage := &mockstorage.MockService{}
	logger := zap.NewNop()
	clock := clockwork.NewFakeClock()
	enableSymmetricEncryption := gofakeit.Bool()
	playerName := gofakeit.Username()
	onlineMessagePeriod := time.Duration(gofakeit.Int64())
	stateMessagePeriod := time.Duration(gofakeit.Int64())
	publishStateLoop := gofakeit.Bool()

	options := []Option{
		WithContext(ctx),
		WithTransport(transport),
		WithStorage(storage),
		WithLogger(logger),
		WithClock(clock),
		WithEnableSymmetricEncryption(enableSymmetricEncryption),
		WithPlayerName(playerName),
		WithOnlineMessagePeriod(onlineMessagePeriod),
		WithStateMessagePeriod(stateMessagePeriod),
		WithPublishStateLoop(publishStateLoop),
	}
	game := NewGame(options)

	require.NotNil(t, game)
	require.Equal(t, ctx, game.ctx)
	require.Equal(t, transport, game.transport)
	require.Equal(t, storage, game.storage)
	require.Equal(t, logger, game.logger)
	require.Equal(t, clock, game.clock)
	require.Equal(t, enableSymmetricEncryption, game.config.EnableSymmetricEncryption)
	require.Equal(t, playerName, game.config.PlayerName)
	require.Equal(t, onlineMessagePeriod, game.config.OnlineMessagePeriod)
	require.Equal(t, stateMessagePeriod, game.config.StateMessagePeriod)
	require.Equal(t, publishStateLoop, game.config.PublishStateLoopEnabled)
}

func TestNoTransport(t *testing.T) {
	options := []Option{
		WithTransport(nil),
		WithClock(clockwork.NewFakeClock()),
	}
	game := NewGame(options)
	require.Nil(t, game)
}

func TestNoClock(t *testing.T) {
	options := []Option{
		WithClock(nil),
		WithTransport(&mocktransport.MockService{}),
	}
	game := NewGame(options)
	require.Nil(t, game)
}

func TestNoContext(t *testing.T) {
	options := []Option{
		WithContext(nil),
		WithTransport(&mocktransport.MockService{}),
		WithClock(clockwork.NewFakeClock()),
	}
	game := NewGame(options)
	require.NotNil(t, game)
	require.Equal(t, context.Background(), game.ctx)
}

func TestNotLogger(t *testing.T) {
	options := []Option{
		WithLogger(nil),
		WithTransport(&mocktransport.MockService{}),
		WithClock(clockwork.NewFakeClock()),
	}
	game := NewGame(options)
	require.NotNil(t, game)
	require.NotNil(t, game.logger)
	require.Equal(t, zapcore.InvalidLevel, game.logger.Level())
}
