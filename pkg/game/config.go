package game

import "time"

type configuration struct {
	PlayerName                string
	EnableSymmetricEncryption bool
	OnlineMessagePeriod       time.Duration
	StateMessagePeriod        time.Duration
	PublishStateLoopEnabled   bool
}

var defaultConfig = configuration{
	PlayerName:                "",
	EnableSymmetricEncryption: true,
	OnlineMessagePeriod:       5 * time.Second,
	StateMessagePeriod:        30 * time.Second,
	PublishStateLoopEnabled:   true,
}
