package main

import (
	"2sp/internal/config"
	"2sp/internal/transport"
	"2sp/internal/view"
	"2sp/pkg/game"
	"2sp/pkg/storage"
	"context"
	"os"
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
		game.WithEnableSymmetricEncryption(config.EnableSymmetricEncryption),
	}

	game := game.NewGame(options)
	if game == nil {
		config.Logger.Fatal("could not create game")
	}
	defer game.Stop()

	code := view.Run(game, waku)
	os.Exit(code)
}

func createStorage() storage.Service {
	if config.Anonymous() {
		return nil
	}
	return storage.NewLocalStorage("")
}
