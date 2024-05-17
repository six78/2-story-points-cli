package main

import (
	"2sp/app"
	"2sp/config"
	"2sp/view"
	"os"
)

func main() {
	config.ParseArguments()
	config.SetupLogger()

	a := app.NewApp()
	defer a.Stop()

	code := view.Run(a)
	os.Exit(code)
}
