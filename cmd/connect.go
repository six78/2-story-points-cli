package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"waku-poker-planning/config"
	"waku-poker-planning/game"
	"waku-poker-planning/view"
	"waku-poker-planning/waku"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an existing poker planning session",
	Run: func(cmd *cobra.Command, args []string) {
		waku, err := waku.NewNode(config.Logger)
		if err != nil {
			config.Logger.Error("failed to create waku node", zap.Error(err))
			return
		}

		err = waku.Start()
		if err != nil {
			config.Logger.Error("failed to start waku node", zap.Error(err))
			return
		}

		defer waku.Stop()

		config.Logger.Info("waiting for peers to connect")

		if !waku.WaitForPeersConnected() {
			config.Logger.Error("failed to connect to peers")
			return
		}

		config.Logger.Info("peers connected")

		game := game.NewGame(config.Logger, waku)
		//gameState := game.SubscribeToStateChanges()

		game.Start()

		//for {
		//	select {
		//	case state := <-gameState:
		//		config.Logger.Info("game state changed", zap.Any("state", state))
		//	}
		//}

		v := view.NewView(game)
		v.Run()
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
