package cmd

import (
	"fmt"
	"os"
	"time"

	"waku-poker-planning/config"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "waku-poker-planning",
	Short: "Decentralized poker planning tool using Waku protocol",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.PersistentFlags().StringVar(&config.Fleet, "fleet", "wakuv2.prod", "Waku fleet to use")
	rootCmd.PersistentFlags().StringVar(&config.PlayerName, "name", generatePlayerName(), "Player name")
	rootCmd.PersistentFlags().StringVar(&config.SessionName, "session", "helloworld", "Session name")
}

func generatePlayerName() string {
	return fmt.Sprintf("player-%d", time.Now().Unix())
}
