package repository

import (
	"context"
	"encoding/json"
	"fmt"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	"sync"
)

type DBService interface {
	Insert(string)
	Get(string) int
	EncodeDatastore() ([]byte, error)
	LoadDatastore(encodedData []byte) error
}

type Database struct {
	datastore   *sync.Map
	snapshotter *Snapshotter
	wal         *WriteAheadLog
	logger      log.Logger
}

func NewDatabase(ctx context.Context, config *config.Config, isMaster bool) *Database {
	var db *Database
	var err error

	snapshotter := NewSnapshotter(ctx, &config.SnapshotOptions)
	logger := ctx.Value(log.LoggerKey).(log.Logger)

	if !isMaster {
		wal := NewWAL(ctx, &config.WALOptions)
		if err := wal.Init(ctx); err != nil {
			panic(fmt.Sprintf("Failed to initialize WAL: ", err))
		}
		db := &Database{
			datastore:   &sync.Map{},
			wal:         wal,
			snapshotter: snapshotter,
			logger:      logger,
		}
		go db.snapshotter.StartSnapshotRoutine(ctx, db)

		return db
	}

	db, err = InitDBFromWal(ctx, &config.WALOptions)
	if err != nil {
		panic(fmt.Sprintf("Cannot restore MasterDB from wal", err))
	}

	db.snapshotter = snapshotter
	db.logger = logger

	go db.snapshotter.StartSnapshotRoutine(ctx, db)

	return db
}

func (db *Database) Insert(word string) {

	// Update in-memory store
	val, loaded := db.datastore.LoadOrStore(word, 1)
	if loaded {
		db.datastore.Store(word, val.(int)+1)
	}
	err := db.wal.Write([]byte(word + "\n"))
	if err != nil {
		db.logger.Error("Cannot write into wal buffer: ", err.Error())
	}
}

func (db *Database) Get(word string) int {
	if val, found := db.datastore.Load(word); found {
		return val.(int)
	}
	return 0
}

func (db *Database) EncodeDatastore() ([]byte, error) {

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

func (db *Database) LoadDatastore(encodedData []byte) error {
	data := make(map[string]int)

	// Decode the JSON data from the byte slice
	if err := json.Unmarshal(encodedData, &data); err != nil {
		return err
	}

	// Create and populate the sync.Map
	snapshotMap := &sync.Map{}
	for k, v := range data {
		snapshotMap.Store(k, v)
	}

	db.datastore = snapshotMap
	return nil
}
