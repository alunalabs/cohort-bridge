package db

import (
	"encoding/csv"
	"errors"
	"os"
	"sync"
)

type CSVDatabase struct {
	headers []string
	data    map[string][]string
	keys    []string
	mu      sync.RWMutex
}

// NewCSVDatabase reads the CSV file and initializes the CSVDatabase.
// The CSV file must have a header and at least one column (the key).
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
	if len(records) < 2 {
		return nil, errors.New("CSV file must have a header and at least one data row")
	}

	headers := records[0]
	data := make(map[string][]string)
	keys := make([]string, 0, len(records)-1)

	for _, record := range records[1:] {
		if len(record) < 1 {
			continue
		}
		key := record[0]
		data[key] = record
		keys = append(keys, key)
	}

	return &CSVDatabase{
		headers: headers,
		data:    data,
		keys:    keys,
	}, nil
}

// Get returns the row as a map[columnName]value for the given key.
func (db *CSVDatabase) Get(key string) (map[string]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	record, ok := db.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}

	result := make(map[string]string)
	for i, header := range db.headers {
		if i < len(record) {
			result[header] = record[i]
		} else {
			result[header] = ""
		}
	}
	return result, nil
}

// List returns a slice of keys starting from `start`, up to `size` entries.
// ListRecords returns a slice of row maps starting from `start` index, up to `size` entries.
func (db *CSVDatabase) List(start, size int) ([]map[string]string, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if start < 0 || start >= len(db.keys) {
		return nil, errors.New("start index out of bounds")
	}

	end := start + size
	if end > len(db.keys) {
		end = len(db.keys)
	}

	result := make([]map[string]string, 0, end-start)
	for _, key := range db.keys[start:end] {
		row := make(map[string]string)
		record := db.data[key]
		for i, header := range db.headers {
			if i < len(record) {
				row[header] = record[i]
			} else {
				row[header] = ""
			}
		}
		result = append(result, row)
	}
	return result, nil
}
