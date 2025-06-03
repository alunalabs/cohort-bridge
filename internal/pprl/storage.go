// storage.go
// Package pprl provides a simple JSON‐line file storage for de‐identified records.
// Each record holds an internal ID, a Bloom filter (as raw []byte), and a MinHash signature.
package pprl

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
)

// Record wraps everything we need to persist per patient (no PHI anywhere).
type Record struct {
	ID        string   `json:"id"`
	BloomData string   `json:"bloom"`   // base64-encoded BloomFilter bytes
	MinHash   []uint32 `json:"minhash"` // signature
}

// Storage writes and reads Record entries to/from a JSON‐line file.
type Storage struct {
	filePath string
}

// NewStorage creates a Storage bound to filePath. If the file does not exist, it will be created.
func NewStorage(filePath string) (*Storage, error) {
	// Ensure file exists
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	f.Close()
	return &Storage{filePath: filePath}, nil
}

// Append writes a single Record as one JSON line (appended).
func (s *Storage) Append(rec *Record) error {
	if rec == nil {
		return errors.New("storage: nil record")
	}
	f, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriter(f)
	line, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = writer.Write(line)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte("\n"))
	if err != nil {
		return err
	}
	return writer.Flush()
}

// LoadAll reads every JSON‐line in the file into a slice of Records.
func (s *Storage) LoadAll() ([]*Record, error) {
	f, err := os.Open(s.filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var results []*Record
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var rec Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return nil, err
		}
		results = append(results, &rec)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// Clear truncates the storage file, removing all existing records
func (s *Storage) Clear() error {
	f, err := os.OpenFile(s.filePath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	return f.Close()
}

// WriteAll writes all records to the storage file, replacing any existing content
func (s *Storage) WriteAll(records []*Record) error {
	// Clear the file first
	if err := s.Clear(); err != nil {
		return err
	}

	// Write all records
	for _, record := range records {
		if err := s.Append(record); err != nil {
			return err
		}
	}
	return nil
}

// Helper: Serialize a BloomFilter into base64 (for Record.BloomData).
func BloomToBase64(bf *BloomFilter) (string, error) {
	bytes, err := bf.MarshalBinary()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Helper: Deserialize a BloomFilter from base64 string.
func BloomFromBase64(encoded string) (*BloomFilter, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	bf := &BloomFilter{}
	if err := bf.UnmarshalBinary(raw); err != nil {
		return nil, err
	}
	return bf, nil
}
