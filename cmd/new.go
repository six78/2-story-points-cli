package cmd

import (
	"fmt"
	"waku-poker-planning/app"

	"github.com/spf13/cobra"
)

var sessionName string

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Start a new poker planning session",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("new called")
		app, err := app.NewApp()
		if err != nil {
			return err
		}
		return app.CreateNewSession(sessionName)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVar(&sessionName, "name", "", "Name of the poker planning session")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// newCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// newCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
