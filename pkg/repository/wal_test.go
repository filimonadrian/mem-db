package repository

import (
	"bytes"
	"context"
	config "mem-db/cmd/config"
	log "mem-db/cmd/logger"
	"os"
	"testing"
	"time"
)

// func generateWord(wordLength int) string {
// 	var sb strings.Builder
// 	for i := 0; i < wordLength; i++ {
// 		sb.WriteByte(letterBytes[rand.Intn(len(letterBytes))])
// 	}
// 	return sb.String()
// }

// // Generates a slice of n random words
// func generateWords(n, wordLength int) []string {
// 	words := make([]string, n)
// 	for i := 0; i < n; i++ {
// 		words[i] = generateWord(wordLength)
// 	}
// 	return words
// }

func getLoggerContext() context.Context {
	options := &log.LoggerOptions{
		LogLevel:    "info",
		LogFilepath: "mem-db/data/mem-db.log",
		Console:     true,
	}

	logger, _ := log.NewConsoleLogger(options)

	ctx := context.WithValue(context.Background(), log.LoggerKey, logger)

	return ctx
}

func TestInit(t *testing.T) {
	options := &config.WALOptions{
		WalFilePath:  "/home/adrian/Documents/mem-db/data/test_wal.log",
		SyncTimer:    2,
		SyncMaxBytes: 1024,
	}
	ctx := getLoggerContext()
	wal := NewWAL(ctx, options)

	err := wal.Init(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if wal.bufWriter == nil {
		t.Errorf("expected bufWriter to be initialized")
	}

	// Clean up the file after the test
	os.Remove(options.WalFilePath)
}

func TestWrite(t *testing.T) {
	options := &config.WALOptions{
		WalFilePath:  "/home/adrian/Documents/mem-db/data/test_wal.log",
		SyncTimer:    1,
		SyncMaxBytes: 1024,
	}

	ctx := getLoggerContext()
	wal := NewWAL(ctx, options)
	wal.Init(ctx)

	data := []byte("Hello, World!")
	err := wal.Write(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = wal.Sync()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Read the file and check its contents
	wal.Close()
	content, err := os.ReadFile(options.WalFilePath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bytes.Equal(content, data) {
		t.Errorf("expected file content to be %v, got %v", data, content)
	}

	// Clean up the file after the test
	os.Remove(options.WalFilePath)
}

func TestSync(t *testing.T) {
	options := &config.WALOptions{
		WalFilePath:  "/home/adrian/Documents/mem-db/data/test_wal.log",
		SyncTimer:    1,
		SyncMaxBytes: 1024,
	}
	ctx := getLoggerContext()

	wal := NewWAL(ctx, options)
	wal.Init(ctx)

	emptyData := []byte("")
	data := []byte("Testing Sync func!")
	wal.Write(data)

	emptyContent, err := os.ReadFile(options.WalFilePath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bytes.Equal(emptyContent, emptyData) {
		t.Errorf("expected file content to be %v, got %v", emptyData, emptyContent)
	}

	err = wal.Sync()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// wal.Close()

	content, err := os.ReadFile(options.WalFilePath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bytes.Equal(content, data) {
		t.Errorf("expected file content to be %v, got %v", data, content)
	}

	// Clean up the file after the test
	os.Remove(options.WalFilePath)
}

func TestKeepSyncing(t *testing.T) {
	options := &config.WALOptions{
		WalFilePath:  "/home/adrian/Documents/mem-db/data/test_wal.log",
		SyncTimer:    5,
		SyncMaxBytes: 1024,
	}

	ctx := getLoggerContext()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wal := NewWAL(ctx, options)
	wal.Init(ctx)

	emptyData := []byte("")
	data := []byte("Test Keep Syncing")

	go wal.KeepSyncing(ctx)
	wal.Write(data)

	emptyContent, err := os.ReadFile(options.WalFilePath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bytes.Equal(emptyContent, emptyData) {
		t.Errorf("expected file content to be %v, got %v", emptyData, emptyContent)
	}

	// Wait for a few seconds to allow the sync to happen
	time.Sleep(8 * time.Second)

	content, err := os.ReadFile(options.WalFilePath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !bytes.Equal(content, data) {
		t.Errorf("expected file content to be %v, got %v", data, content)
	}

	// Clean up the file after the test
	os.Remove(options.WalFilePath)
}
