package repository

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

func RecoverDB(walFilePath string) (*sync.Map, *os.File, error) {

	file, err := os.OpenFile(walFilePath, os.O_RDWR, 0666)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to open WAL file: %v", err)
	}

	// Initialize the in-memory database.
	db := &sync.Map{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			// Increment the count for this word in the database.
			val, loaded := db.LoadOrStore(word, 1)
			if loaded {
				db.Store(word, val.(int)+1)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("Error reading WAL file: %v", err)
	}

	// Move the file pointer to the end for future appends
	_, err = file.Seek(0, os.SEEK_END)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to seek to end of WAL file: %v", err)
	}

	// Return the reconstructed database and the file pointer.
	return db, file, nil
}
