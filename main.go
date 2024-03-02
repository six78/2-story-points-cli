package main

import (
	"os"
	"waku-poker-planning/app"
	"waku-poker-planning/config"
	"waku-poker-planning/view"
)

/*
    waku-pp connect --session="helloworld" --name="igor"
	waku-pp new --session="six78 sprint 42" --fleet="wakuv2.prod"
*/

func main() {
	config.SetupLogger()
	config.ParseArguments()

	a := app.NewApp()
	defer a.Stop()

	code := view.Run(a)
	os.Exit(code)
}
