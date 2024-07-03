package main

import (
	"context"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jonboulle/clockwork"
	"go.uber.org/zap"

	"github.com/six78/2-story-points-cli/cmd/2sp/demo"
	"github.com/six78/2-story-points-cli/internal/config"
	"github.com/six78/2-story-points-cli/internal/transport"
	"github.com/six78/2-story-points-cli/internal/view"
	"github.com/six78/2-story-points-cli/pkg/game"
	"github.com/six78/2-story-points-cli/pkg/storage"
)

func main() {
	config.ParseArguments()
	config.SetupLogger()

	ctx, quit := context.WithCancel(context.Background())
	defer quit()

	waku := transport.NewNode(ctx, config.Logger)
	defer waku.Stop()

	options := []game.Option{
		game.WithContext(ctx),
		game.WithTransport(waku),
		game.WithStorage(createStorage()),
		game.WithLogger(config.Logger.Named("game")),
		game.WithPlayerName(config.PlayerName()),
		game.WithOnlineMessagePeriod(config.OnlineMessagePeriod),
		game.WithStateMessagePeriod(config.StateMessagePeriod),
		game.WithEnableSymmetricEncryption(config.EnableSymmetricEncryption),
		game.WithClock(clockwork.NewRealClock()),
	}

	game := game.NewGame(options)
	if game == nil {
		config.Logger.Fatal("could not create game")
	}
	defer game.Stop()

	// Create UI model and program
	model := view.InitialModel(game, waku)
	program := tea.NewProgram(model)

	// Run demo if enabled
	if config.Demo() {
		demonstration := demo.New(ctx, game, program)
		go func() {
			demonstration.Routine()
			program.Quit()
		}()
	}

	if _, err := program.Run(); err != nil {
		config.Logger.Error("error running program", zap.Error(err))
		os.Exit(1)
		return
	}

	os.Exit(0)
}

func createStorage() storage.Service {
	if config.Anonymous() {
		return nil
	}
	return storage.NewLocalStorage("")
}
