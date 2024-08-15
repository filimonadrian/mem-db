package repository

import (
	"context"
	log "mem-db/cmd/logger"
	"sync"
)

type Database struct {
	datastore *sync.Map
	logger    log.Logger
}

type Repository interface {
}

func NewDatabase(ctx context.Context) *Database {
	return &Database{
		datastore: &sync.Map{},
		logger:    ctx.Value(log.LoggerKey).(log.Logger),
	}
}

func (db *Database) Insert(word string) {

	// Update in-memory store
	val, loaded := db.datastore.LoadOrStore(word, 1)
	if loaded {
		db.datastore.Store(word, val.(int)+1)
	}
}

func (db *Database) Get(word string) int {
	if val, found := db.datastore.Load(word); found {
		return val.(int)
	}
	return 0
}
