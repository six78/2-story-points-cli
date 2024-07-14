package game

import "time"

type configuration struct {
	PlayerName                string
	EnableSymmetricEncryption bool
	OnlineMessagePeriod       time.Duration
	StateMessagePeriod        time.Duration
	PublishStateLoopEnabled   bool
	AutoRevealEnabled         bool
	AutoRevealDelay           time.Duration
}

var defaultConfig = configuration{
	PlayerName:                "",
	EnableSymmetricEncryption: true,
	OnlineMessagePeriod:       5 * time.Second,
	StateMessagePeriod:        30 * time.Second,
	PublishStateLoopEnabled:   true,
	AutoRevealEnabled:         true,
	AutoRevealDelay:           1 * time.Second,
}
