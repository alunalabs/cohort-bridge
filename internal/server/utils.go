package server

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"

	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// Global shared MinHash instance to ensure consistent parameters across all datasets
var (
	globalMinHashInstance *pprl.MinHash
	globalMinHashOnce     sync.Once
)

// GetGlobalMinHash returns a shared MinHash instance with consistent parameters
func GetGlobalMinHash() (*pprl.MinHash, error) {
	var err error
	globalMinHashOnce.Do(func() {
		// Create deterministic MinHash by manually creating one with fixed parameters
		globalMinHashInstance, err = createDeterministicMinHash(1000, 128)
	})
	return globalMinHashInstance, err
}

// createDeterministicMinHash creates a MinHash with deterministic parameters using a fixed seed
func createDeterministicMinHash(m, s uint32) (*pprl.MinHash, error) {
	if m == 0 || s == 0 {
		return nil, fmt.Errorf("invalid parameters: m=%d, s=%d", m, s)
	}

	// Use the same prime as the original implementation
	const prime uint32 = 2147483647 // Mersenne prime (2^31 - 1)
	if m >= prime {
		return nil, fmt.Errorf("m too large for chosen prime")
	}

	// Use a fixed seed for deterministic results
	rng := rand.New(rand.NewSource(42)) // Fixed seed = 42

	a := make([]uint32, s)
	b := make([]uint32, s)

	// Generate deterministic coefficients using the seeded RNG
	for i := uint32(0); i < s; i++ {
		a[i] = uint32(rng.Int31n(int32(prime-1))) + 1 // [1..prime-1]
		b[i] = uint32(rng.Int31n(int32(prime)))       // [0..prime-1]
	}

	// Create binary data with our deterministic parameters using proper encoding
	bufSize := 4 + int(s)*4 + int(s)*4 + 4 + int(s)*4
	buf := make([]byte, bufSize)

	offset := 0

	// Write s
	binary.LittleEndian.PutUint32(buf[offset:offset+4], s)
	offset += 4

	// Write a array
	for i := uint32(0); i < s; i++ {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], a[i])
		offset += 4
	}

	// Write b array
	for i := uint32(0); i < s; i++ {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], b[i])
		offset += 4
	}

	// Write prime
	binary.LittleEndian.PutUint32(buf[offset:offset+4], prime)
	offset += 4

	// Write signature array (all prime values initially)
	for i := uint32(0); i < s; i++ {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], prime)
		offset += 4
	}

	// Create new MinHash and unmarshal our deterministic data
	deterministicMH := &pprl.MinHash{}
	err := deterministicMH.UnmarshalBinary(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal deterministic MinHash: %v", err)
	}

	return deterministicMH, nil
}

// EnsureOutputDirectory creates the output directory if it doesn't exist
func EnsureOutputDirectory() error {
	outputDir := "out"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory '%s': %w", outputDir, err)
	}
	return nil
}

// EnsureLogsDirectory creates the logs directory if it doesn't exist
func EnsureLogsDirectory() error {
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory '%s': %w", logsDir, err)
	}
	return nil
}

// LoadTokenizedRecords loads patient records from tokenized data
func LoadTokenizedRecords(filename string) ([]PatientRecord, error) {
	// Load tokenized database
	tokenDB, err := db.NewTokenizedDatabase(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load tokenized database: %w", err)
	}

	// Convert to Bloom filter records
	bfRecords, err := tokenDB.ToBloomFilterRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to Bloom filter records: %w", err)
	}

	// Convert to PatientRecord format
	var records []PatientRecord
	for _, bfRecord := range bfRecords {
		records = append(records, PatientRecord{
			ID:          bfRecord.ID,
			BloomFilter: bfRecord.BloomFilter,
			MinHash:     bfRecord.MinHash,
		})
	}

	return records, nil
}

// LoadPatientRecordsUtil converts CSV data to Bloom filter representations
func LoadPatientRecordsUtil(csvDB *db.CSVDatabase, fields []string) ([]PatientRecord, error) {
	return LoadPatientRecordsUtilWithRandomBits(csvDB, fields, 0.0)
}

// LoadPatientRecordsUtilWithRandomBits converts CSV data to Bloom filter representations with configurable random bits
func LoadPatientRecordsUtilWithRandomBits(csvDB *db.CSVDatabase, fields []string, randomBitsPercent float64) ([]PatientRecord, error) {
	// Get all records
	allRecords, err := csvDB.List(0, 1000000) // Large number to get all records
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %v", err)
	}

	// Get the GLOBAL shared MinHash instance to ensure consistent parameters across ALL datasets
	sharedMinHash, err := GetGlobalMinHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get global shared MinHash: %v", err)
	}

	var records []PatientRecord
	for _, record := range allRecords {
		// Create Bloom filter for this record with optional random bits
		bf := pprl.NewBloomFilterWithRandomBits(1000, 5, randomBitsPercent) // 1000 bits, 5 hash functions

		// Create MinHash instance with SAME parameters as shared instance
		recordMinHash, err := recreateMinHashFromShared(sharedMinHash)
		if err != nil {
			return nil, fmt.Errorf("failed to create MinHash for record: %v", err)
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

		// Compute MinHash signature from Bloom filter ONCE and store it
		signature, err := recordMinHash.ComputeSignature(bf)
		if err != nil {
			return nil, fmt.Errorf("failed to compute MinHash signature: %v", err)
		}

		records = append(records, PatientRecord{
			ID:               record["id"], // Assuming 'id' is the primary key
			BloomFilter:      bf,
			MinHash:          recordMinHash,
			MinHashSignature: signature, // Store the computed signature
		})
	}

	return records, nil
}

// normalizeFieldUtil normalizes field values for consistent hashing
func normalizeFieldUtil(value string) string {
	// Convert to lowercase and remove spaces for consistent matching
	return strings.ToLower(strings.ReplaceAll(value, " ", ""))
}

// recreateMinHashFromShared creates a new MinHash instance with the same parameters as the shared one
func recreateMinHashFromShared(sharedMinHash *pprl.MinHash) (*pprl.MinHash, error) {
	// Serialize the shared MinHash to get its parameters
	data, err := sharedMinHash.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal shared MinHash: %v", err)
	}

	// Create a new MinHash instance and deserialize the parameters
	newMinHash := &pprl.MinHash{}
	err = newMinHash.UnmarshalBinary(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal MinHash: %v", err)
	}

	return newMinHash, nil
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
