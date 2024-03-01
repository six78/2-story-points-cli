package config

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

const OnlineMessagePeriod = 10 * time.Second
const logsDirectory = "logs"

var Fleet string
var SessionName string
var PlayerName string
var Logger *zap.Logger

var LogFilePath string

func SetupLogger() {
	LogFilePath = createLogFile()
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{LogFilePath}
	config.Development = false
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	Logger = logger
}

func createLogFile() string {
	executableFilePath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	executableDir := filepath.Dir(executableFilePath)
	name := fmt.Sprintf("waku-pp-%s.log", time.Now().Format(time.RFC3339))
	path := filepath.Join(executableDir, logsDirectory, name)

	if err := os.MkdirAll(filepath.Dir(path), 0770); err != nil {
		panic(err)
	}

	if _, err := os.Create(path); err != nil {
		panic(err)
	}

	return path
}
