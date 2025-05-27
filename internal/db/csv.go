package db

import (
	"encoding/csv"
	"errors"
	"os"
	"sync"
)

type CSVDatabase struct {
	data map[string]string
	keys []string
	mu   sync.RWMutex
}

// NewCSVDatabase reads the CSV file and initializes the CSVDatabase.
// The CSV file must have two columns: key and value.
func NewCSVDatabase(filePath string) (*CSVDatabase, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	db := &CSVDatabase{
		data: make(map[string]string),
		keys: make([]string, 0, len(records)),
	}

	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		key := record[0]
		value := record[1]
		db.data[key] = value
		db.keys = append(db.keys, key)
	}

	return db, nil
}

// Get retrieves the value for a given key.
func (db *CSVDatabase) Get(key string) (string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	value, ok := db.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return value, nil
}

// List returns a slice of keys, starting from `start` index and up to `size` entries.
func (db *CSVDatabase) List(start, size int) ([]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if start < 0 || start >= len(db.keys) {
		return nil, errors.New("start index out of bounds")
	}
	end := start + size
	if end > len(db.keys) {
		end = len(db.keys)
	}
	return db.keys[start:end], nil
}
