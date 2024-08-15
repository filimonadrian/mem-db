package repository

import (
	"encoding/json"
	"os"
	"sync"
)

func (db *Database) CreateSnapshot(snapshotPath string) error {
	file, err := os.Create(snapshotPath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := make(map[string]int)
	db.datastore.Range(func(key, value interface{}) bool {
		data[key.(string)] = value.(int)
		return true
	})

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

func (db *Database) LoadSnapshot(snapshotPath string) error {
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
