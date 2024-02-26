package cmd

import (
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to an existing poker planning session",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("connect called")

	},
}

func init() {
	rootCmd.AddCommand(connectCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	//connectCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	connectCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
