package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Start a new poker planning session",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("new called")
		return nil
		//app, err := app.NewApp()
		//if err != nil {
		//	return err
		//}
		//return app.CreateNewSession(sessionName)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
