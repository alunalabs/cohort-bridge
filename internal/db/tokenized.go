package db

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// TokenizedRecord represents a record with tokenized data
type TokenizedRecord struct {
	ID          string `json:"id"`
	BloomFilter string `json:"bloom_filter"` // Base64 encoded
	MinHash     string `json:"minhash"`      // Base64 encoded
	Timestamp   string `json:"timestamp"`
}

// TokenizedDatabase handles operations on tokenized patient data
type TokenizedDatabase struct {
	filename string
	records  []TokenizedRecord
}

// NewTokenizedDatabase creates a new database for tokenized records
func NewTokenizedDatabase(filename string) (*TokenizedDatabase, error) {
	db := &TokenizedDatabase{
		filename: filename,
	}

	if err := db.load(); err != nil {
		return nil, err
	}

	return db, nil
}

// load reads tokenized data from file (supports both JSON and CSV formats)
func (db *TokenizedDatabase) load() error {
	file, err := os.Open(db.filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", db.filename, err)
	}
	defer file.Close()

	// Detect file format by extension
	if db.filename[len(db.filename)-5:] == ".json" {
		return db.loadJSON(file)
	} else if db.filename[len(db.filename)-4:] == ".csv" {
		return db.loadCSV(file)
	} else {
		return fmt.Errorf("unsupported file format: %s (use .json or .csv)", db.filename)
	}
}

// loadJSON loads tokenized data from JSON format
func (db *TokenizedDatabase) loadJSON(file *os.File) error {
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&db.records); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}
	return nil
}

// loadCSV loads tokenized data from CSV format
func (db *TokenizedDatabase) loadCSV(file *os.File) error {
	reader := csv.NewReader(file)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Validate header
	if len(header) < 3 {
		return fmt.Errorf("invalid CSV header: expected at least id, bloom_filter, minhash")
	}

	// Read records
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read CSV row: %w", err)
		}

		if len(row) < 3 {
			continue // Skip invalid rows
		}

		record := TokenizedRecord{
			ID:          row[0],
			BloomFilter: row[1],
			MinHash:     row[2],
		}

		if len(row) > 3 {
			record.Timestamp = row[3]
		}

		db.records = append(db.records, record)
	}

	return nil
}

// List returns a slice of tokenized records with pagination
func (db *TokenizedDatabase) List(offset, limit int) ([]TokenizedRecord, error) {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(db.records) {
		return []TokenizedRecord{}, nil
	}

	end := offset + limit
	if end > len(db.records) {
		end = len(db.records)
	}

	return db.records[offset:end], nil
}

// GetByID returns a specific tokenized record by ID
func (db *TokenizedDatabase) GetByID(id string) (*TokenizedRecord, error) {
	for _, record := range db.records {
		if record.ID == id {
			return &record, nil
		}
	}
	return nil, fmt.Errorf("record with ID %s not found", id)
}

// Count returns the total number of records
func (db *TokenizedDatabase) Count() int {
	return len(db.records)
}

// ToBloomFilterRecords converts tokenized records to Bloom filter objects
func (db *TokenizedDatabase) ToBloomFilterRecords() ([]BloomFilterRecord, error) {
	var bfRecords []BloomFilterRecord

	for _, record := range db.records {
		// Decode Bloom filter
		bf, err := pprl.BloomFromBase64(record.BloomFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to decode Bloom filter for ID %s: %w", record.ID, err)
		}

		// Decode MinHash
		mh, err := pprl.MinHashFromBase64(record.MinHash)
		if err != nil {
			return nil, fmt.Errorf("failed to decode MinHash for ID %s: %w", record.ID, err)
		}

		bfRecord := BloomFilterRecord{
			ID:          record.ID,
			BloomFilter: bf,
			MinHash:     mh,
		}

		bfRecords = append(bfRecords, bfRecord)
	}

	return bfRecords, nil
}

// BloomFilterRecord represents a record with decoded Bloom filter objects
type BloomFilterRecord struct {
	ID          string
	BloomFilter *pprl.BloomFilter
	MinHash     *pprl.MinHash
}

// Save saves tokenized records to file
func (db *TokenizedDatabase) Save() error {
	file, err := os.Create(db.filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", db.filename, err)
	}
	defer file.Close()

	// Save based on file extension
	if db.filename[len(db.filename)-5:] == ".json" {
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		return encoder.Encode(db.records)
	} else if db.filename[len(db.filename)-4:] == ".csv" {
		writer := csv.NewWriter(file)
		defer writer.Flush()

		// Write header
		header := []string{"id", "bloom_filter", "minhash", "timestamp"}
		if err := writer.Write(header); err != nil {
			return err
		}

		// Write records
		for _, record := range db.records {
			row := []string{record.ID, record.BloomFilter, record.MinHash, record.Timestamp}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("unsupported file format for saving: %s", db.filename)
}

// Add adds a new tokenized record
func (db *TokenizedDatabase) Add(record TokenizedRecord) {
	db.records = append(db.records, record)
}

// GetStats returns basic statistics about the tokenized database
func (db *TokenizedDatabase) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_records":  len(db.records),
		"filename":       db.filename,
		"has_timestamps": false,
	}

	// Check if any records have timestamps
	for _, record := range db.records {
		if record.Timestamp != "" {
			stats["has_timestamps"] = true
			break
		}
	}

	// Sample record for inspection
	if len(db.records) > 0 {
		stats["sample_id"] = db.records[0].ID
		stats["bloom_filter_length"] = len(db.records[0].BloomFilter)
		stats["minhash_length"] = len(db.records[0].MinHash)
	}

	return stats
}
