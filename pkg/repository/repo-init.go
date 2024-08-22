package repository

import (
	"context"
	"fmt"
	// log "mem-db/cmd/logger"
	config "mem-db/cmd/config"
	"os"
	"path/filepath"
)

// receives the entire database config
func InitDBSystem(ctx context.Context, options *config.WALOptions) (*Database, *WriteAheadLog, error) {
	walFilePath := options.WalFilePath

	// check if the path exists
	dir := filepath.Dir(walFilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("The directory %s does not exist: %v", dir, err)
	}

	wal := NewWAL(ctx, options)

	// check if the file exists
	_, err := os.Stat(walFilePath)

	if err != nil && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("Cannot access file %s: %v", walFilePath, err)
	}

	if options.Restore && err == nil {

		// retrieve data and move pointer to the end of the file
		datastore, file, err := RecoverDB(walFilePath)
		if err != nil {
			return nil, nil, fmt.Errorf("Cannot recover database: %v", err.Error())
		}

		wal.SetFile(file)
		if err := wal.Init(ctx); err != nil {
			return nil, nil, fmt.Errorf("Failed to initialize WAL: %v", err)
		}
		return &Database{datastore: datastore}, wal, nil
	}

	// If the WAL file doesn't exist, initialize a new WAL
	if os.IsNotExist(err) {
		if err := wal.Init(ctx); err != nil {
			return nil, nil, fmt.Errorf("Failed to initialize WAL: %v", err)
		}

		return NewDatabase(ctx), wal, nil
	}

	return nil, nil, fmt.Errorf("Unexpected error while initializing DB system")
}
