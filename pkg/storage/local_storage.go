package storage

import (
	"2sp/internal/config"
	protocol2 "2sp/pkg/protocol"
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

type LocalStorage struct {
	player playerStorage

	configDirs configdir.ConfigDir
	mutex      *sync.RWMutex
}

type playerStorage struct {
	ID   protocol2.PlayerID `json:"id"`
	Name string             `json:"name"`
}

type roomStorage struct {
	// TODO: PrivateKey string
	State *protocol2.State `json:"state"`
}

func NewStorage() (*LocalStorage, error) {
	s := &LocalStorage{
		configDirs: configdir.New(config.VendorName, config.ApplicationName),
		mutex:      &sync.RWMutex{},
	}
	return s, s.initialize()
}

func (s *LocalStorage) initialize() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	err := s.readPlayer()
	config.Logger.Info("storage initialized",
		zap.Any("player", s.player),
		zap.String("configDir", s.configDirs.QueryFolderContainsFile(playerStorageFileName).Path),
		zap.Error(err),
	)
	return err
}

func (s *LocalStorage) readPlayer() error {
	folder := s.configDirs.QueryFolderContainsFile(playerStorageFileName)
	if folder == nil {
		config.Logger.Info("no player storage found")
		return nil
	}

	data, err := folder.ReadFile(playerStorageFileName)
	if err != nil {
		return errors.Wrap(err, "failed to read player data")
	}

	err = json.Unmarshal(data, &s.player)
	if err == nil {
		return nil
	}

	config.Logger.Error("failed to parse player storage, clearing storage", zap.Error(err))

	err = s.ResetPlayer()
	if err != nil {
		config.Logger.Error("failed to reset player storage", zap.Error(err))
	}

	return nil
}

func (s *LocalStorage) savePlayerStorage() error {
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

func (s *LocalStorage) ResetPlayer() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.player.ID = ""
	s.player.Name = ""
	return s.savePlayerStorage()
}

func (s *LocalStorage) PlayerID() protocol2.PlayerID {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.player.ID
}

func (s *LocalStorage) SetPlayerID(id protocol2.PlayerID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.player.ID = id
	return s.savePlayerStorage()
}

func (s *LocalStorage) PlayerName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.player.Name
}

func (s *LocalStorage) SetPlayerName(name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.player.Name = name
	return s.savePlayerStorage()
}

func (s *LocalStorage) LoadRoomState(roomID protocol2.RoomID) (*protocol2.State, error) {
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

func (s *LocalStorage) SaveRoomState(roomID protocol2.RoomID, state *protocol2.State) error {
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

func roomFilePath(roomID protocol2.RoomID) string {
	return path.Join(roomsDirectory, roomID.String()+".json")
}
