package app

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"sync"
	"waku-poker-planning/config"
	"waku-poker-planning/protocol"

	"github.com/shibukawa/configdir"
)

const playerFileName = "player.json"
const dbFileName = "db.json"

type Storage struct {
	playerID protocol.PlayerID

	configDirs configdir.ConfigDir
	mutex      *sync.RWMutex
}

func NewStorage() (*Storage, error) {
	s := &Storage{
		configDirs: configdir.New(config.VendorName, config.ApplicationName),
		mutex:      &sync.RWMutex{},
	}
	return s, s.initialize()
}

func (s *Storage) initialize() error {
	var err error
	s.playerID, err = s.readPlayerID()
	config.Logger.Info("storage initialized",
		zap.String("playerID", string(s.playerID)),
		zap.String("configDir", s.configDirs.QueryFolderContainsFile(playerFileName).Path),
	)
	return err
}

func (s *Storage) readPlayerID() (protocol.PlayerID, error) {
	folder := s.configDirs.QueryFolderContainsFile(playerFileName)
	if folder == nil {
		config.Logger.Info("no player UUID found, creating a new one")
		return s.createPlayerID()
	}

	data, err := folder.ReadFile(playerFileName)
	if err != nil {
		return "", errors.Wrap(err, "failed to read config data")
	}

	playerUUID, err := uuid.ParseBytes(data)
	if err == nil {
		return protocol.PlayerID(playerUUID.String()), nil
	}

	config.Logger.Error("failed to parse player UUID, creating a new one", zap.Error(err))
	return s.createPlayerID()
}

func (s *Storage) createPlayerID() (protocol.PlayerID, error) {
	playerUUID, err := config.GeneratePlayerID()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate player UUID")
	}

	folders := s.configDirs.QueryFolders(configdir.Global)
	err = folders[0].WriteFile(playerFileName, []byte(playerUUID))
	if err != nil {
		return "", errors.Wrap(err, "failed to save player UUID")
	}

	return playerUUID, nil
}

func (s *Storage) PlayerID() protocol.PlayerID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.playerID
}
