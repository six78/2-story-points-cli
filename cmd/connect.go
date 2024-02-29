package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"waku-poker-planning/config"
	"waku-poker-planning/game"
	"waku-poker-planning/waku"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an existing poker planning session",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("connect called")

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
		gameState := game.SubscribeToStateChanges()

		game.Start()

		for {
			select {
			case state := <-gameState:
				config.Logger.Info("game state changed", zap.Any("state", state))
			}
		}

		//app.ReceiveMessages(messagesChannel)

		//go func() {
		//i := 0
		//for {
		//	time.Sleep(5 * time.Second)
		//	i++
		//	message := []byte(fmt.Sprintf("%s: Hello from Go mazafaka (%d)", config.PlayerName, i))
		//	_ = waku.PublishMessage(contentTopic, message)
		//}
		//}()
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
