package main

import (
	"waku-poker-planning/cmd"
	"waku-poker-planning/config"
)

/*
    waku-pp connect --session="helloworld" --name="igor"
	waku-pp new --session="six78 sprint 42" --fleet="wakuv2.prod"
*/

func main() {
	config.SetupLogger()
	cmd.Execute()
}
