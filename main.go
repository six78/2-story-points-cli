package main

import (
	"os"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
	"waku-poker-planning/view"
)

func main() {
	config.SetupLogger()
	config.ParseArguments()

	a := app.NewApp()
	defer a.Stop()

	code := view.Run(a)
	os.Exit(code)
}
