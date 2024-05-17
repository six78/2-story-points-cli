package app

import (
	"2sp/config"
	"2sp/game"
	"2sp/protocol"
	"encoding/json"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"path"
	"sync"

	"github.com/shibukawa/configdir"
)

const (
	playerStorageFileName = "player.json"
	roomsDirectory        = "rooms"
)

type Storage struct {
	player playerStorage

	configDirs configdir.ConfigDir
	mutex      *sync.RWMutex
}

type playerStorage struct {
	ID   protocol.PlayerID `json:"id"`
	Name string            `json:"name"`
}

type roomStorage struct {
	// TODO: PrivateKey string
	State *protocol.State `json:"state"`
}

func NewStorage() (*Storage, error) {
	s := &Storage{
		configDirs: configdir.New(config.VendorName, config.ApplicationName),
		mutex:      &sync.RWMutex{},
	}
	return s, s.initialize()
}

func (s *Storage) initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	err := s.readPlayer()
	config.Logger.Info("storage initialized",
		zap.Any("player", s.player),
		zap.String("configDir", s.configDirs.QueryFolderContainsFile(playerStorageFileName).Path),
	)
	return err
}

func (s *Storage) readPlayer() error {
	folder := s.configDirs.QueryFolderContainsFile(playerStorageFileName)
	if folder == nil {
		config.Logger.Info("no player UUID found, creating a new one")
		return s.createPlayerID()
	}

	data, err := folder.ReadFile(playerStorageFileName)
	if err != nil {
		return errors.Wrap(err, "failed to read player data")
	}

	err = json.Unmarshal(data, &s.player)
	if err == nil {
		return nil
	}

	config.Logger.Error("failed to parse player storage, creating a new one", zap.Error(err))
	return s.createPlayerID()
}

func (s *Storage) createPlayerID() error {
	playerUUID, err := game.GeneratePlayerID()
	if err != nil {
		return errors.Wrap(err, "failed to generate player id")
	}

	s.player.ID = playerUUID
	s.player.Name = ""

	return s.savePlayerStorage()
}

func (s *Storage) savePlayerStorage() error {
	playerJson, err := json.Marshal(s.player)
	if err != nil {
		return errors.Wrap(err, "failed to marshal player storage")
	}

	folders := s.configDirs.QueryFolders(configdir.Global)
	err = folders[0].WriteFile(playerStorageFileName, playerJson)
	if err != nil {
		return errors.Wrap(err, "failed to save player storage")
	}

	return nil
}

func (s *Storage) PlayerID() protocol.PlayerID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.player.ID
}

func (s *Storage) PlayerName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.player.Name
}

func (s *Storage) SetPlayerName(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.player.Name = name
	return s.savePlayerStorage()
}

func (s *Storage) LoadRoomState(roomID protocol.RoomID) (*protocol.State, error) {
	filePath := roomFilePath(roomID)
	folder := s.configDirs.QueryFolderContainsFile(filePath)
	if folder == nil {
		return nil, nil
	}

	data, err := folder.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read room storage file")
	}

	var room roomStorage
	err = json.Unmarshal(data, &room)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal storage file")
	}

	return room.State, nil
}

func (s *Storage) SaveRoomState(roomID protocol.RoomID, state *protocol.State) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	room := roomStorage{
		State: state,
	}

	roomJson, err := json.Marshal(room)
	if err != nil {
		return errors.Wrap(err, "failed to marshal room data")
	}

	filePath := roomFilePath(roomID)
	folders := s.configDirs.QueryFolders(configdir.Global)

	err = folders[0].WriteFile(filePath, roomJson)
	if err != nil {
		return errors.Wrap(err, "failed to write room storage")
	}

	return nil
}

func roomFilePath(roomID protocol.RoomID) string {
	return path.Join(roomsDirectory, roomID.String()+".json")
}
