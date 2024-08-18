package repository

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecoverDB(t *testing.T) {
	// Setup: Create a temporary WAL file for testing
	tmpFile, err := os.CreateTemp("", "test-wal-*.wal")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write test data to the WAL file
	_, err = tmpFile.WriteString("apple\nbanana\napple\n")
	assert.NoError(t, err)

	// Close the file to flush the contents to disk
	err = tmpFile.Close()
	assert.NoError(t, err)

	// Test: Recover the database from the WAL file
	db, file, err := RecoverDB(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, file)
	defer file.Close()

	// Validate the recovered database
	val, ok := db.Load("apple")
	assert.True(t, ok)
	assert.Equal(t, 2, val)

	val, ok = db.Load("banana")
	assert.True(t, ok)
	assert.Equal(t, 1, val)

	// Ensure file pointer is at the end
	pos, err := file.Seek(0, os.SEEK_CUR)
	assert.NoError(t, err)
	stat, err := file.Stat()
	assert.NoError(t, err)
	assert.Equal(t, stat.Size(), pos)
}

func TestRecoverDB_EmptyFile(t *testing.T) {
	// Setup: Create an empty temporary WAL file for testing
	tmpFile, err := os.CreateTemp("", "test-wal-empty-*.wal")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Close the file to flush the contents to disk
	err = tmpFile.Close()
	assert.NoError(t, err)

	// Test: Recover the database from the empty WAL file
	db, file, err := RecoverDB(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.NotNil(t, file)
	defer file.Close()

	// Validate the recovered database is empty
	count := 0
	db.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count)

	// Ensure file pointer is at the end
	pos, err := file.Seek(0, os.SEEK_CUR)
	assert.NoError(t, err)
	stat, err := file.Stat()
	assert.NoError(t, err)
	assert.Equal(t, stat.Size(), pos)
}

func TestRecoverDB_NonExistentFile(t *testing.T) {
	// Test: Attempt to recover the database from a non-existent WAL file
	db, file, err := RecoverDB("non-existent-file.wal")
	assert.Nil(t, db)
	assert.Nil(t, file)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Failed to open WAL file")
}
