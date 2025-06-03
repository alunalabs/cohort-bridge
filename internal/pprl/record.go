// record.go
// Package pprl provides record creation and processing functionality.
package pprl

import (
	"fmt"
)

// RecordConfig holds configuration for record creation
type RecordConfig struct {
	BloomSize    uint32  // Size of Bloom filter in bits
	BloomHashes  uint32  // Number of hash functions for Bloom filter
	MinHashSize  uint32  // Size of MinHash signature
	QGramLength  int     // Length of q-grams
	QGramPadding string  // Padding character for q-grams
	NoiseLevel   float64 // Probability of noise in Bloom filter (0-1)
}

// CreateRecord creates a new record from a set of fields
func CreateRecord(id string, fields []string, config *RecordConfig) (*Record, error) {
	if config == nil {
		return nil, fmt.Errorf("record: nil config")
	}

	// Create and populate Bloom filter
	bf := NewBloomFilter(config.BloomSize, config.BloomHashes)
	if bf == nil {
		return nil, fmt.Errorf("record: failed to create bloom filter")
	}

	// Create q-gram set
	qgs := NewQGramSet(config.QGramLength, config.QGramPadding)

	// Process each field
	for _, field := range fields {
		// Normalize the field
		normalized := NormalizeString(field)

		// Extract q-grams
		qgs.ExtractQGrams(normalized)

		// Add to Bloom filter
		if config.NoiseLevel > 0 {
			bf.AddWithNoise([]byte(normalized), config.NoiseLevel)
		} else {
			bf.Add([]byte(normalized))
		}
	}

	// Create MinHash
	mh, err := NewMinHash(config.BloomSize, config.MinHashSize)
	if err != nil {
		return nil, fmt.Errorf("record: failed to create minhash: %w", err)
	}

	// Compute MinHash signature
	signature, err := mh.ComputeSignature(bf)
	if err != nil {
		return nil, fmt.Errorf("record: failed to compute signature: %w", err)
	}

	// Serialize Bloom filter
	bloomData, err := BloomToBase64(bf)
	if err != nil {
		return nil, fmt.Errorf("record: failed to serialize bloom filter: %w", err)
	}

	// Serialize q-gram set
	qgramData, err := QGramToBase64(qgs)
	if err != nil {
		return nil, fmt.Errorf("record: failed to serialize q-gram set: %w", err)
	}

	return &Record{
		ID:        id,
		BloomData: bloomData,
		MinHash:   signature,
		QGramData: qgramData,
	}, nil
}

// ProcessRecord processes an existing record with new fields
func ProcessRecord(record *Record, fields []string, config *RecordConfig) (*Record, error) {
	if record == nil {
		return nil, fmt.Errorf("record: nil record")
	}
	if config == nil {
		return nil, fmt.Errorf("record: nil config")
	}

	// Deserialize existing Bloom filter
	bf, err := BloomFromBase64(record.BloomData)
	if err != nil {
		return nil, fmt.Errorf("record: failed to deserialize bloom filter: %w", err)
	}

	// Deserialize existing q-gram set
	qgs, err := QGramFromBase64(record.QGramData)
	if err != nil {
		return nil, fmt.Errorf("record: failed to deserialize q-gram set: %w", err)
	}

	// Process new fields
	for _, field := range fields {
		normalized := NormalizeString(field)

		// Update q-grams
		qgs.ExtractQGrams(normalized)

		// Update Bloom filter
		if config.NoiseLevel > 0 {
			bf.AddWithNoise([]byte(normalized), config.NoiseLevel)
		} else {
			bf.Add([]byte(normalized))
		}
	}

	// Create new MinHash
	mh, err := NewMinHash(config.BloomSize, config.MinHashSize)
	if err != nil {
		return nil, fmt.Errorf("record: failed to create minhash: %w", err)
	}

	// Compute new signature
	signature, err := mh.ComputeSignature(bf)
	if err != nil {
		return nil, fmt.Errorf("record: failed to compute signature: %w", err)
	}

	// Serialize updated data
	bloomData, err := BloomToBase64(bf)
	if err != nil {
		return nil, fmt.Errorf("record: failed to serialize bloom filter: %w", err)
	}

	qgramData, err := QGramToBase64(qgs)
	if err != nil {
		return nil, fmt.Errorf("record: failed to serialize q-gram set: %w", err)
	}

	return &Record{
		ID:        record.ID,
		BloomData: bloomData,
		MinHash:   signature,
		QGramData: qgramData,
	}, nil
}
