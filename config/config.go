package config

import (
	"flag"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

const OnlineMessagePeriod = 10 * time.Second
const StateMessagePeriod = 10 * time.Second
const logsDirectory = "logs"
const SymmetricKeyLength = 16

var Fleet string
var PlayerName string
var PlayerID string

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

func ParseArguments() {
	flag.StringVar(&Fleet, "fleet", "wakuv2.prod", "Waku fleet name")
	flag.StringVar(&PlayerName, "name", generatePlayerName(), "Player name")
	flag.Parse()

	// TODO: Find a better place for ths
	playerUuid, _ := uuid.NewUUID()
	PlayerID = playerUuid.String()

	// TODO: Consider positional arguments as first command
	// return strings.Join(flag.Args(), " ")
}

func generatePlayerName() string {
	return fmt.Sprintf("player-%d", time.Now().Unix())
}
