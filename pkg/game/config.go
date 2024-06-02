package game

import "time"

type gameConfig struct {
	PlayerName                string
	EnableSymmetricEncryption bool
	OnlineMessagePeriod       time.Duration
	StateMessagePeriod        time.Duration
}

var defaultConfig = gameConfig{
	PlayerName:                "",
	EnableSymmetricEncryption: true,
	OnlineMessagePeriod:       5 * time.Second,
	StateMessagePeriod:        30 * time.Second,
}
