package config

import (
	"go.uber.org/zap"
	"time"
)

const OnlineMessagePeriod = 10 * time.Second

var Fleet string
var SessionName string
var PlayerName string
var Logger *zap.Logger

func SetupLogger() {
	config := zap.NewDevelopmentConfig()
	config.Development = false
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	Logger = logger
}
