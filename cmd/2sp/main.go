package main

import (
	"2sp/internal/app"
	"2sp/internal/config"
	"2sp/internal/view"
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
