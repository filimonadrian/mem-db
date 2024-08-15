package repository

import (
	"sync"
)

type Database struct {
	datastore *sync.Map
	mutex     sync.RWMutex
}

type Repository interface {
}

func NewDatabase() *Database {
	return &Database{
		datastore: &sync.Map{},
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
