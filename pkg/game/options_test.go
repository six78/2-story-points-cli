package game

import (
	mocktransport "2sp/internal/transport/mock"
	mockstorage "2sp/pkg/storage/mock"
	"context"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"testing"
)

func TestOptions(t *testing.T) {
	ctx := context.Background()
	transport := &mocktransport.MockService{}
	storage := &mockstorage.MockService{}
	logger := zap.NewNop()
	const enableSymmetricEncryption = false

	options := []Option{
		WithContext(ctx),
		WithTransport(transport),
		WithStorage(storage),
		WithLogger(logger),
		WithEnableSymmetricEncryption(false),
	}
	game := NewGame(options)

	require.NotNil(t, game)
	require.Equal(t, ctx, game.ctx)
	require.Equal(t, transport, game.transport)
	require.Equal(t, storage, game.storage)
	require.Equal(t, logger, game.logger)
	require.Equal(t, enableSymmetricEncryption, game.config.EnableSymmetricEncryption)
}

func TestNoTransport(t *testing.T) {
	options := []Option{
		WithTransport(nil),
	}
	game := NewGame(options)
	require.Nil(t, game)
}
