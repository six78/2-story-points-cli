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

	game := game.NewGame(ctx, waku, createStorage())
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
