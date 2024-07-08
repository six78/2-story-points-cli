package testcommon

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/six78/2-story-points-cli/internal/config"
)

func SetupConfigLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	config.Logger = logger
	return logger
}
