package main

import (
	"fmt"
	"runtime/debug"

	"github.com/six78/2-story-points-cli/internal/config"
	"github.com/six78/2-story-points-cli/pkg/storage"
)

func main() {

	info, ok := debug.ReadBuildInfo()
	fmt.Println(info, ok)

	//short := versioninfo.Short()
	//fmt.Println(short)

	//fmt.Println("Version:", versioninfo.Version)
	//fmt.Println("Revision:", versioninfo.Revision)
	//fmt.Println("DirtyBuild:", versioninfo.DirtyBuild)
	//fmt.Println("LastCommit:", versioninfo.LastCommit)

	//program.Println("ok: %w", ok)

	//
	//config.ParseArguments()
	//config.SetupLogger()
	//
	//ctx, quit := context.WithCancel(context.Background())
	//defer quit()
	//
	//waku := transport.NewNode(ctx, config.Logger)
	//defer waku.Stop()
	//
	//options := []game.Option{
	//	game.WithContext(ctx),
	//	game.WithTransport(waku),
	//	game.WithStorage(createStorage()),
	//	game.WithLogger(config.Logger.Named("game")),
	//	game.WithPlayerName(config.PlayerName()),
	//	game.WithOnlineMessagePeriod(config.OnlineMessagePeriod),
	//	game.WithStateMessagePeriod(config.StateMessagePeriod),
	//	game.WithEnableSymmetricEncryption(config.EnableSymmetricEncryption),
	//	game.WithClock(clockwork.NewRealClock()),
	//}
	//
	//game := game.NewGame(options)
	//if game == nil {
	//	config.Logger.Fatal("could not create game")
	//}
	//defer game.Stop()
	//
	//// Create UI model and program
	//model := view.InitialModel(game, waku)
	//program := tea.NewProgram(model)
	//
	//
	//// Run demo if enabled
	//if config.Demo() {
	//	demonstration := demo.New(ctx, game, program)
	//	go func() {
	//		demonstration.Routine()
	//		program.Quit()
	//	}()
	//}
	//
	//if _, err := program.Run(); err != nil {
	//	config.Logger.Error("error running program", zap.Error(err))
	//	os.Exit(1)
	//	return
	//}
	//
	//os.Exit(0)
}

func createStorage() storage.Service {
	if config.Anonymous() {
		return nil
	}
	return storage.NewLocalStorage("")
}
