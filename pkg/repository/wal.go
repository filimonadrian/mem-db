package repository

import (
	"bufio"
	"fmt"
	"os"
	// "path/filepath"
	"context"
	log "mem-db/cmd/logger"
	"sync"
	"time"
)

type WALOptions struct {
	WalFilePath  string `json:"walFilePath"`
	SyncTimer    int    `json:"syncTimer"`
	SyncMaxBytes int    `json:"syncMaxBytes"`
}

type WriteAheadLog struct {
	walFilePath  string
	mutex        sync.RWMutex
	bufWriter    *bufio.Writer
	syncTimer    *time.Ticker
	file         *os.File
	syncMaxBytes int64
	logger       log.Logger
}

func NewWAL(ctx context.Context, options *WALOptions) *WriteAheadLog {

	wal := &WriteAheadLog{
		walFilePath:  options.WalFilePath,
		syncTimer:    time.NewTicker(time.Duration(options.SyncTimer) * time.Second),
		syncMaxBytes: int64(options.SyncMaxBytes),
		logger:       ctx.Value(log.LoggerKey).(log.Logger),
	}

	return wal
}

func (wal *WriteAheadLog) SetFile(file *os.File) {
	wal.file = file
}

func (wal *WriteAheadLog) Init() error {
	if wal.file == nil {
		if err := wal.createWALFile(); err != nil {
			return err
		}
	}

	wal.bufWriter = bufio.NewWriter(wal.file)
	// go wal.keepSyncing()

	return nil
}

// appends the provided data to the log
func (wal *WriteAheadLog) Write(data []byte) error {
	wal.mutex.Lock()
	defer wal.mutex.Unlock()

	// Write data payload to the buffer
	if _, err := wal.bufWriter.Write(data); err != nil {
		return err
	}
	return nil
}

func (wal *WriteAheadLog) createWALFile() error {

	walFilePath := wal.walFilePath
	file, err := os.OpenFile(walFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("Could not create WALfile %s: %v", walFilePath, err.Error())
	}
	// defer file.Close()

	fmt.Printf("WALFile %s created.\n", walFilePath)
	wal.file = file
	return nil

}

// keepSyncing periodically triggers a synchronous write to the disk to ensure data durability.
func (wal *WriteAheadLog) KeepSyncing(ctx context.Context) {
	for {
		select {
		case <-wal.syncTimer.C:
			wal.mutex.Lock()

			wal.logger.Debug("Ticker for flushing data")
			err := wal.Sync()
			if err != nil {
				fmt.Printf("Error while performing sync %v", err.Error())
			}
			wal.mutex.Unlock()
		case <-ctx.Done():
			wal.mutex.Lock()

			wal.logger.Debug("Flushing buffer before stopping the app")
			err := wal.Sync()
			if err != nil {
				fmt.Printf("Error while performing sync %v", err.Error())
			}
			wal.mutex.Unlock()
			wal.Close()
			return
		}
	}
}

// writes data to the disk
func (wal *WriteAheadLog) Sync() error {
	err := wal.bufWriter.Flush()
	wal.logger.Info("Flushing Data..")
	if err != nil {
		return fmt.Errorf("Cannot flush data: %v", err.Error())
	}
	return wal.file.Sync()
}

func (wal *WriteAheadLog) Close() error {
	wal.mutex.Lock()
	defer wal.mutex.Unlock()

	// Flush data to disk
	if err := wal.Sync(); err != nil {
		return err
	}

	return wal.file.Close()
}
