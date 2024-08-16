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

	var wal *WriteAheadLog

	// logger := ctx.Value(log.LoggerKey).(log.Logger),
	walFilePath := options.WalFilePath
	// check if the path exists
	dir := filepath.Dir(walFilePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("The directory does not exist: %s", dir)
	}

	wal = NewWAL(ctx, options)

	// check if the file exists
	_, err := os.Stat(walFilePath)

	if os.IsNotExist(err) {
		if err := wal.Init(ctx); err != nil {
			return nil, nil, err
		}

		return NewDatabase(ctx), wal, nil

		// other issues with file
	} else if err != nil {
		return nil, nil, fmt.Errorf("Cannot access file %s: %v", walFilePath, err.Error())
	}

	// WALFile exist, retrieve data and move pointer to the end of the file
	// start the recovery process
	datastore, file, err := RecoverDB(walFilePath)

	if err != nil {
		return nil, nil, fmt.Errorf("Cannot recover database: %v", err.Error())
	}

	wal.SetFile(file)
	wal.Init(ctx)
	return &Database{
			datastore: datastore,
		},
		wal,
		nil
}
