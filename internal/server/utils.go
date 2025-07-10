package server

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

var (
	globalMinHash *pprl.MinHash
	minHashMutex  sync.Mutex
)

// GetGlobalMinHash returns a shared MinHash instance to ensure consistent signatures across all datasets
// This is CRITICAL for proper matching - all parties must use the same MinHash parameters
func GetGlobalMinHash() (*pprl.MinHash, error) {
	minHashMutex.Lock()
	defer minHashMutex.Unlock()

	if globalMinHash == nil {
		// Create with deterministic, agreed-upon parameters for consistency
		var err error
		globalMinHash, err = createDeterministicMinHash(100, 1000) // 100 signatures, 1000 bloom size - matches PPRL workflow
		if err != nil {
			return nil, fmt.Errorf("failed to create global MinHash: %v", err)
		}
	}

	return globalMinHash, nil
}

// createDeterministicMinHash creates a MinHash with deterministic parameters
func createDeterministicMinHash(m, s uint32) (*pprl.MinHash, error) {
	return pprl.NewMinHashSeeded(m, s, "cohort-bridge-pprl-seed")
}

// EnsureOutputDirectory ensures the output directory exists
func EnsureOutputDirectory() error {
	outputDir := "out"

	// Check if directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		// Create directory with appropriate permissions
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
		}
		Info("Created output directory: %s", outputDir)
	} else if err != nil {
		return fmt.Errorf("failed to check output directory %s: %w", outputDir, err)
	}

	return nil
}

// EnsureLogsDirectory ensures the logs directory exists
func EnsureLogsDirectory() error {
	logsDir := "logs"

	// Check if directory exists
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		// Create directory with appropriate permissions
		if err := os.MkdirAll(logsDir, 0755); err != nil {
			return fmt.Errorf("failed to create logs directory %s: %w", logsDir, err)
		}
		Info("Created logs directory: %s", logsDir)
	} else if err != nil {
		return fmt.Errorf("failed to check logs directory %s: %w", logsDir, err)
	}

	return nil
}

// LoadTokenizedRecords loads PPRL records from tokenized data for zero-knowledge processing
func LoadTokenizedRecords(filename string, isEncrypted bool, encryptionKey string, encryptionKeyFile string) ([]*pprl.Record, error) {
	var actualFilename string
	var needsCleanup bool

	// Auto-detect encryption if filename ends with .enc
	if !isEncrypted && strings.HasSuffix(filename, ".enc") {
		isEncrypted = true
	}

	// Handle encryption if specified or auto-detected
	if isEncrypted {
		var keyHex string
		var err error

		// Determine encryption key source
		if encryptionKey != "" {
			// Use provided hex key
			keyHex = encryptionKey
		} else if encryptionKeyFile != "" {
			// Load key from file
			keyHex, err = loadKeyFromFile(encryptionKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load encryption key from %s: %w", encryptionKeyFile, err)
			}
		} else {
			// Try to find key file based on data filename
			if strings.HasSuffix(filename, ".enc") {
				keyFile := strings.TrimSuffix(filename, ".enc") + ".key"
				if _, err := os.Stat(keyFile); err == nil {
					keyHex, err = loadKeyFromFile(keyFile)
					if err != nil {
						return nil, fmt.Errorf("failed to load encryption key from %s: %w", keyFile, err)
					}
				} else {
					return nil, fmt.Errorf("encrypted tokenized file %s found but no encryption key specified and no key file %s available", filename, keyFile)
				}
			} else {
				return nil, fmt.Errorf("data marked as encrypted but no encryption key specified")
			}
		}

		// Decrypt the file to a temporary location
		tempFile := filename + ".tmp_decrypted"

		if err := decryptFile(filename, tempFile, keyHex); err != nil {
			return nil, fmt.Errorf("failed to decrypt tokenized file %s: %w", filename, err)
		}

		actualFilename = tempFile
		needsCleanup = true

		// Schedule cleanup of temporary file
		defer func() {
			if needsCleanup {
				os.Remove(tempFile)
			}
		}()
	} else {
		// Regular plaintext file
		actualFilename = filename
	}

	// Load tokenized database from the actual file (decrypted temp file or original plaintext)
	tokenDB, err := db.NewTokenizedDatabase(actualFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to load tokenized database: %w", err)
	}

	// Convert to Bloom filter records
	bfRecords, err := tokenDB.ToBloomFilterRecords()
	if err != nil {
		return nil, fmt.Errorf("failed to convert to Bloom filter records: %w", err)
	}

	// Convert to PPRL Record format for zero-knowledge processing
	var records []*pprl.Record
	for _, bfRecord := range bfRecords {
		// Encode Bloom filter to base64
		bloomData, err := bfRecord.BloomFilter.ToBase64()
		if err != nil {
			return nil, fmt.Errorf("failed to encode Bloom filter: %w", err)
		}

		// Compute MinHash signature from the Bloom filter
		signature, err := bfRecord.MinHash.ComputeSignature(bfRecord.BloomFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to compute MinHash signature: %w", err)
		}

		records = append(records, &pprl.Record{
			ID:        bfRecord.ID,
			BloomData: bloomData,
			MinHash:   signature,
			QGramData: "", // Not used in tokenized records
		})
	}

	return records, nil
}

// LoadPatientRecordsUtil converts CSV data to zero-knowledge PPRL records
func LoadPatientRecordsUtil(csvDB *db.CSVDatabase, fields []string) ([]*pprl.Record, error) {
	return LoadPatientRecordsUtilWithRandomBits(csvDB, fields, 0.0)
}

// LoadPatientRecordsUtilWithRandomBits converts CSV data to zero-knowledge PPRL records with configurable random bits
func LoadPatientRecordsUtilWithRandomBits(csvDB *db.CSVDatabase, fields []string, randomBitsPercent float64) ([]*pprl.Record, error) {
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

	var records []*pprl.Record
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

		// Encode Bloom filter to base64
		bloomData, err := bf.ToBase64()
		if err != nil {
			return nil, fmt.Errorf("failed to encode Bloom filter: %w", err)
		}

		records = append(records, &pprl.Record{
			ID:        record["id"], // Assuming 'id' is the primary key
			BloomData: bloomData,
			MinHash:   signature, // Store the computed signature
			QGramData: "",        // Not used in this format
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

// ZKStreamingRecordIterator provides streaming access to zero-knowledge PPRL records
// This function is designed to work with the new zero-knowledge matching infrastructure
type ZKStreamingRecordIterator struct {
	csvDB         *db.CSVDatabase
	fields        []string
	batchSize     int
	randomBits    float64
	currentBatch  []*pprl.Record
	offset        int
	hasMore       bool
	sharedMinHash *pprl.MinHash
}

// NewZKStreamingRecordIterator creates a new streaming record iterator for zero-knowledge processing
func NewZKStreamingRecordIterator(csvDB *db.CSVDatabase, fields []string, batchSize int, randomBitsPercent float64) (*ZKStreamingRecordIterator, error) {
	sharedMinHash, err := GetGlobalMinHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get global MinHash: %v", err)
	}

	return &ZKStreamingRecordIterator{
		csvDB:         csvDB,
		fields:        fields,
		batchSize:     batchSize,
		randomBits:    randomBitsPercent,
		offset:        0,
		hasMore:       true,
		sharedMinHash: sharedMinHash,
	}, nil
}

// NextBatch returns the next batch of zero-knowledge PPRL records
func (iter *ZKStreamingRecordIterator) NextBatch() ([]*pprl.Record, error) {
	// Get records from database
	rawRecords, err := iter.csvDB.List(iter.offset, iter.batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list records at offset %d: %v", iter.offset, err)
	}

	if len(rawRecords) == 0 {
		iter.hasMore = false
		return nil, nil
	}

	var records []*pprl.Record
	for _, record := range rawRecords {
		// Create Bloom filter for this record
		bf := pprl.NewBloomFilterWithRandomBits(1000, 5, iter.randomBits)

		// Create MinHash instance
		recordMinHash, err := recreateMinHashFromShared(iter.sharedMinHash)
		if err != nil {
			return nil, fmt.Errorf("failed to create MinHash: %v", err)
		}

		// Add configured fields to Bloom filter using q-grams
		for _, field := range iter.fields {
			if value, exists := record[field]; exists && value != "" {
				normalized := normalizeFieldUtil(value)
				qgrams := generateQGrams(normalized, 2)

				for _, qgram := range qgrams {
					bf.Add([]byte(qgram))
				}
			}
		}

		// Compute MinHash signature
		signature, err := recordMinHash.ComputeSignature(bf)
		if err != nil {
			return nil, fmt.Errorf("failed to compute MinHash signature: %v", err)
		}

		// Encode Bloom filter to base64
		bloomData, err := bf.ToBase64()
		if err != nil {
			return nil, fmt.Errorf("failed to encode Bloom filter: %w", err)
		}

		records = append(records, &pprl.Record{
			ID:        record["id"],
			BloomData: bloomData,
			MinHash:   signature,
			QGramData: "",
		})
	}

	iter.currentBatch = records
	iter.offset += len(rawRecords)

	// Check if we have fewer records than requested - indicates end of data
	if len(rawRecords) < iter.batchSize {
		iter.hasMore = false
	}

	return records, nil
}

// HasMore returns true if there are more records to process
func (iter *ZKStreamingRecordIterator) HasMore() bool {
	return iter.hasMore
}

// GetCurrentOffset returns the current offset in the dataset
func (iter *ZKStreamingRecordIterator) GetCurrentOffset() int {
	return iter.offset
}

// GetBatchSize returns the configured batch size
func (iter *ZKStreamingRecordIterator) GetBatchSize() int {
	return iter.batchSize
}

// GetEstimatedTotalRecords returns an estimate of total records (if available)
func (iter *ZKStreamingRecordIterator) GetEstimatedTotalRecords() (int, error) {
	// This is a rough estimate - get a large batch to count
	tempRecords, err := iter.csvDB.List(0, 1000000)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate total records: %v", err)
	}
	return len(tempRecords), nil
}

// loadKeyFromFile loads an encryption key from a file
func loadKeyFromFile(keyFile string) (string, error) {
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	// Parse the file line by line, ignoring comments
	lines := strings.Split(string(keyData), "\n")
	var keyHex string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// First non-comment line should be the key
		keyHex = line
		break
	}

	if keyHex == "" {
		return "", fmt.Errorf("no encryption key found in file")
	}

	// Validate hex format
	if _, err := hex.DecodeString(keyHex); err != nil {
		return "", fmt.Errorf("invalid hex key format: %w", err)
	}

	return keyHex, nil
}

// decryptFile decrypts a file using AES-GCM with the provided hex key
func decryptFile(inputFile, outputFile, keyHex string) error {
	// Decode hex key
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("invalid hex key: %w", err)
	}

	// Read encrypted file
	encryptedData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:nonceSize]
	ciphertext := encryptedData[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	// Write decrypted data
	if err := os.WriteFile(outputFile, plaintext, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}
