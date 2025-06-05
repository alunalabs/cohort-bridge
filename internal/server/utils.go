package server

import (
	"fmt"
	"strings"

	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// LoadPatientRecordsUtil converts CSV data to Bloom filter representations
func LoadPatientRecordsUtil(csvDB *db.CSVDatabase, fields []string) ([]PatientRecord, error) {
	// Get all records
	allRecords, err := csvDB.List(0, 1000000) // Large number to get all records
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %v", err)
	}

	var records []PatientRecord
	for _, record := range allRecords {
		// Create Bloom filter for this record
		bf := pprl.NewBloomFilter(1000, 5) // 1000 bits, 5 hash functions

		// Create MinHash for this record
		mh, err := pprl.NewMinHash(1000, 128) // m=1000 (same as BF), s=128 hash functions
		if err != nil {
			return nil, fmt.Errorf("failed to create MinHash: %v", err)
		}

		// Add configured fields to Bloom filter using q-grams
		for _, field := range fields {
			if value, exists := record[field]; exists && value != "" {
				// Normalize and convert to q-grams
				normalized := normalizeFieldUtil(value)
				qgrams := generateQGrams(normalized, 2) // Use 2-grams

				// Add each q-gram to the Bloom filter
				for _, qgram := range qgrams {
					bf.Add([]byte(qgram))
				}
			}
		}

		// Compute MinHash signature from Bloom filter
		_, err = mh.ComputeSignature(bf)
		if err != nil {
			return nil, fmt.Errorf("failed to compute MinHash signature: %v", err)
		}

		records = append(records, PatientRecord{
			ID:          record["id"], // Assuming 'id' is the primary key
			BloomFilter: bf,
			MinHash:     mh,
		})
	}

	return records, nil
}

// normalizeFieldUtil normalizes field values for consistent hashing
func normalizeFieldUtil(value string) string {
	// Convert to lowercase and remove spaces for consistent matching
	return strings.ToLower(strings.ReplaceAll(value, " ", ""))
}

// generateQGrams creates character q-grams from a string
func generateQGrams(text string, q int) []string {
	if len(text) < q {
		return []string{text} // Return the whole string if shorter than q
	}

	// Use a map to store unique q-grams (prevents duplicates)
	qgramSet := make(map[string]bool)

	// Add padding for beginning and end
	padded := strings.Repeat("_", q-1) + text + strings.Repeat("_", q-1)

	// Generate q-grams and add to set
	for i := 0; i <= len(padded)-q; i++ {
		qgram := padded[i : i+q]
		qgramSet[qgram] = true
	}

	// Convert map keys to slice
	var qgrams []string
	for qgram := range qgramSet {
		qgrams = append(qgrams, qgram)
	}

	return qgrams
}
