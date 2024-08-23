package repository

import (
	"context"
	"encoding/json"
	log "mem-db/cmd/logger"
	"os"
	"sync"
	"testing"
	"time"
)

func TestSaveDataToFile(t *testing.T) {

	dir, err := os.MkdirTemp("", "snapshot_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	snapshotter := &Snapshotter{
		dirPath: dir,
	}

	data := []byte(`{"key1": 100, "key2": 200}`)

	err = snapshotter.SaveDataToFile(data)
	if err != nil {
		t.Fatalf("SaveDataToFile failed: %v", err)
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file in directory, but found %d", len(files))
	}
}

func TestCreateSnapshot(t *testing.T) {

	dir, err := os.MkdirTemp("", "snapshot_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	db := &Database{
		datastore: &sync.Map{},
	}
	db.datastore.Store("word1", 10)
	db.datastore.Store("word2", 20)

	snapshotter := &Snapshotter{
		dirPath: dir,
	}

	err = snapshotter.CreateSnapshot(db)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}

	// Ensure snapshot file was created
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Expected 1 file in directory, but found %d", len(files))
	}

	// Verify the contents of the snapshot
	snapshotFile := files[0]
	data, err := os.ReadFile(dir + "/" + snapshotFile.Name())
	if err != nil {
		t.Fatalf("Failed to read snapshot file: %v", err)
	}

	var snapshotData map[string]int
	err = json.Unmarshal(data, &snapshotData)
	if err != nil {
		t.Fatalf("Failed to unmarshal snapshot data: %v", err)
	}

	if snapshotData["word1"] != 10 || snapshotData["word2"] != 20 {
		t.Fatalf("Snapshot data does not match expected values")
	}
}

func TestLoadSnapshot(t *testing.T) {
	data := []byte(`{"word1": 10, "word2": 20}`)

	snapshotter := &Snapshotter{}

	snapshotMap, err := snapshotter.LoadSnapshot(data)
	if err != nil {
		t.Fatalf("LoadSnapshot failed: %v", err)
	}

	// Check if data was loaded correctly into the sync.Map
	word1, _ := snapshotMap.Load("word1")
	word2, _ := snapshotMap.Load("word2")

	if word1.(int) != 10 || word2.(int) != 20 {
		t.Fatalf("Snapshot data does not match expected values")
	}
}

func TestStartSnapshotRoutine(t *testing.T) {

	dir, err := os.MkdirTemp("", "snapshot_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(dir)

	db := &Database{
		datastore: &sync.Map{},
	}
	db.datastore.Store("word1", 10)
	db.datastore.Store("word2", 20)

	// logger, err = log.NewConsoleLogger(&config.LoggerOptions)
	logger, _ := log.NewConsoleLogger(&log.LoggerOptions{
		LogLevel:    "debug",
		LogFilepath: "",
		Console:     true,
	})

	snapshotter := &Snapshotter{
		dirPath:   dir,
		syncTimer: time.NewTicker(1 * time.Second),
		logger:    logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go snapshotter.StartSnapshotRoutine(ctx, db)

	// wait to trigger the routine
	time.Sleep(3 * time.Second)

	// Cancel the context to stop the routine
	cancel()

	// Verify that at least one snapshot was created
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}

	if len(files) == 0 {
		t.Fatalf("Expected at least one snapshot file, but found none")
	}
}
