package repository

import (
	"context"
	"encoding/json"
	"fmt"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	"os"
	"sync"
	"time"
)

type Snapshotter struct {
	dirPath   string
	syncTimer *time.Ticker
	logger    log.Logger
}

func NewSnapshotter(ctx context.Context, options *config.SnapshotOptions, db *Database) *Snapshotter {
	return &Snapshotter{
		dirPath:   options.DirPath,
		syncTimer: time.NewTicker(time.Duration(options.SyncTimer) * time.Second),
		logger:    ctx.Value(log.LoggerKey).(log.Logger),
	}
}

func generateSnapshotFilename() string {
	return fmt.Sprintf("snapshot_%s.json", time.Now().Format("20060102_150405"))
}

func (s *Snapshotter) EncodeDatastore(db *Database) ([]byte, error) {

	data := make(map[string]int)
	db.datastore.Range(func(key, value interface{}) bool {
		data[key.(string)] = value.(int)
		return true
	})

	// Marshal the map into JSON
	encodedData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Cannot encode map to JSON: %v", err)
	}

	return encodedData, nil
}

func (s *Snapshotter) SaveDataToFile(encodedData []byte) error {

	snapshotPath := fmt.Sprintf("%s/%s", s.dirPath, generateSnapshotFilename())

	// Create the snapshot file
	file, err := os.Create(snapshotPath)
	if err != nil {
		return fmt.Errorf("Cannot create snapshot file: %v", err)
	}
	defer file.Close()

	// Write the encoded data to the file
	_, err = file.Write(encodedData)
	if err != nil {
		return fmt.Errorf("Cannot write data to snapshot file: %v", err)
	}

	return nil
}

func (s *Snapshotter) CreateSnapshot(db *Database) error {
	snapshotPath := fmt.Sprintf("%s/%s", s.dirPath, generateSnapshotFilename())
	file, err := os.Create(snapshotPath)
	if err != nil {
		return fmt.Errorf("Cannot create snapshot file: %v", err)
	}
	defer file.Close()

	data := make(map[string]int)
	db.datastore.Range(func(key, value interface{}) bool {
		data[key.(string)] = value.(int)
		return true
	})

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("Cannot encode map to json: %v", err)
	}

	return nil
}

func (s *Snapshotter) StartSnapshotRoutine(ctx context.Context, db *Database) {

	createAndLogSnapshot := func() {
		if err := s.CreateSnapshot(db); err != nil {
			s.logger.Warn("Cannot create snapshot: %v", err)
		}
	}

	for {
		select {
		case <-s.syncTimer.C:
			s.logger.Debug("Timer for creating db snapshot")
			createAndLogSnapshot()
		case <-ctx.Done():
			s.logger.Info("Creating last snapshot of database before stopping..")
			createAndLogSnapshot()
			return
		}
	}
}

func (s *Snapshotter) LoadSnapshot(encodedData []byte) (*sync.Map, error) {
	data := make(map[string]int)

	// Decode the JSON data from the byte slice
	if err := json.Unmarshal(encodedData, &data); err != nil {
		return nil, err
	}

	// Create and populate the sync.Map
	snapshotMap := &sync.Map{}
	for k, v := range data {
		snapshotMap.Store(k, v)
	}

	return snapshotMap, nil
}

func (s *Snapshotter) LoadSnapshotFromFile(snapshotPath string, db *Database) error {

	file, err := os.Open(snapshotPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := make(map[string]int)
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	db.datastore = &sync.Map{}
	for k, v := range data {
		db.datastore.Store(k, v)
	}

	return nil
}
