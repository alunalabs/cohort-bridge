// blocking.go
// Package match provides secure blocking functionality using commutative encryption
// to privately determine which records should be compared without revealing the blocking keys.
package match

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
)

// BlockingConfig defines the configuration for secure blocking
type BlockingConfig struct {
	MaxBucketsPerRecord int     // Maximum number of buckets a single record can be in
	SimilarityThreshold float64 // Minimum Jaccard similarity to consider for blocking
}

// BlockingBucket represents a secure blocking bucket with encrypted MinHash values
type BlockingBucket struct {
	BucketID        string                     `json:"bucket_id"`
	EncryptedValues []*crypto.CommutativePoint `json:"encrypted_values"`
	RecordIDs       []string                   `json:"record_ids"`
}

// SecureBlocker handles the secure blocking process
type SecureBlocker struct {
	config  *BlockingConfig
	key     *crypto.CommutativeKey
	buckets map[string]*BlockingBucket
}

// NewSecureBlocker creates a new secure blocking instance
func NewSecureBlocker(config *BlockingConfig) (*SecureBlocker, error) {
	key, err := crypto.GenerateCommutativeKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate commutative key: %w", err)
	}

	return &SecureBlocker{
		config:  config,
		key:     key,
		buckets: make(map[string]*BlockingBucket),
	}, nil
}

// CreateBlocks generates secure blocking buckets from MinHash signatures
// Each record is placed into multiple buckets based on different hash bands
func (sb *SecureBlocker) CreateBlocks(records []RecordWithMinHash) ([]*BlockingBucket, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records provided")
	}

	// LSH parameters: divide MinHash signature into bands
	sigLength := len(records[0].MinHash)
	bandSize := 4 // Number of hash values per band
	numBands := sigLength / bandSize

	if numBands == 0 {
		numBands = 1
		bandSize = sigLength
	}

	// Create blocking buckets for each band
	bucketMap := make(map[string]*BlockingBucket)

	for _, record := range records {
		if len(record.MinHash) != sigLength {
			return nil, fmt.Errorf("inconsistent MinHash signature length")
		}

		// Process each band
		for band := 0; band < numBands; band++ {
			start := band * bandSize
			end := start + bandSize
			if end > len(record.MinHash) {
				end = len(record.MinHash)
			}

			// Create a blocking key from this band
			blockingKey := createBlockingKey(record.MinHash[start:end], band)

			// Encrypt the blocking key
			encryptedKey, err := sb.key.EncryptString(blockingKey)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt blocking key: %w", err)
			}

			// Create bucket ID from encrypted key
			bucketID := base64.StdEncoding.EncodeToString(encryptedKey.Bytes())

			// Add to bucket
			if bucket, exists := bucketMap[bucketID]; exists {
				bucket.EncryptedValues = append(bucket.EncryptedValues, encryptedKey)
				bucket.RecordIDs = append(bucket.RecordIDs, record.ID)
			} else {
				bucketMap[bucketID] = &BlockingBucket{
					BucketID:        bucketID,
					EncryptedValues: []*crypto.CommutativePoint{encryptedKey},
					RecordIDs:       []string{record.ID},
				}
			}
		}
	}

	// Convert map to slice - include all buckets for two-party matching
	var result []*BlockingBucket
	for _, bucket := range bucketMap {
		// Include all buckets - filtering will happen during two-party intersection
		result = append(result, bucket)
	}

	return result, nil
}

// RecordWithMinHash represents a record with its MinHash signature
type RecordWithMinHash struct {
	ID      string
	MinHash []uint32
}

// createBlockingKey creates a deterministic string from a MinHash band
func createBlockingKey(band []uint32, bandIndex int) string {
	h := sha256.New()

	// Add band index to ensure different bands create different keys
	binary.Write(h, binary.LittleEndian, uint32(bandIndex))

	// Add all values in the band
	for _, val := range band {
		binary.Write(h, binary.LittleEndian, val)
	}

	hash := h.Sum(nil)
	return fmt.Sprintf("block_%x", hash[:16]) // Use first 16 bytes as blocking key
}

// GetCandidatePairs returns candidate pairs from blocking buckets
func (sb *SecureBlocker) GetCandidatePairs(buckets []*BlockingBucket) []CandidatePair {
	var pairs []CandidatePair
	seen := make(map[string]bool)

	for _, bucket := range buckets {
		// Generate all pairs within this bucket
		for i := 0; i < len(bucket.RecordIDs); i++ {
			for j := i + 1; j < len(bucket.RecordIDs); j++ {
				id1, id2 := bucket.RecordIDs[i], bucket.RecordIDs[j]

				// Ensure consistent ordering for deduplication
				if id1 > id2 {
					id1, id2 = id2, id1
				}

				pairKey := id1 + "|" + id2
				if !seen[pairKey] {
					seen[pairKey] = true
					pairs = append(pairs, CandidatePair{
						ID1:      id1,
						ID2:      id2,
						BucketID: bucket.BucketID,
					})
				}
			}
		}
	}

	return pairs
}

// CandidatePair represents a pair of records that should be compared
type CandidatePair struct {
	ID1      string `json:"id1"`
	ID2      string `json:"id2"`
	BucketID string `json:"bucket_id"`
}

// ExchangeEncryptedBuckets simulates the exchange of encrypted bucket information
// between two parties without revealing the actual bucket contents
func (sb *SecureBlocker) ExchangeEncryptedBuckets(myBuckets []*BlockingBucket,
	theirKey *crypto.CommutativeKey) ([]*BlockingBucket, error) {

	var doubleEncryptedBuckets []*BlockingBucket

	for _, bucket := range myBuckets {
		doubleEncrypted := &BlockingBucket{
			BucketID:        bucket.BucketID,
			EncryptedValues: make([]*crypto.CommutativePoint, len(bucket.EncryptedValues)),
			RecordIDs:       bucket.RecordIDs, // Keep record IDs for internal tracking
		}

		// Double encrypt each value with the other party's key
		for i, encVal := range bucket.EncryptedValues {
			doubleEncrypted.EncryptedValues[i] = theirKey.Encrypt(encVal)
		}

		doubleEncryptedBuckets = append(doubleEncryptedBuckets, doubleEncrypted)
	}

	return doubleEncryptedBuckets, nil
}

// FindMatchingBuckets finds buckets that match between two sets of double-encrypted buckets
func FindMatchingBuckets(buckets1, buckets2 []*BlockingBucket) []BucketMatch {
	var matches []BucketMatch

	// Create lookup map for buckets2
	bucket2Map := make(map[string]*BlockingBucket)
	for _, bucket := range buckets2 {
		for _, encVal := range bucket.EncryptedValues {
			key := base64.StdEncoding.EncodeToString(encVal.Bytes())
			bucket2Map[key] = bucket
		}
	}

	// Find matches
	for _, bucket1 := range buckets1 {
		for _, encVal1 := range bucket1.EncryptedValues {
			key := base64.StdEncoding.EncodeToString(encVal1.Bytes())
			if bucket2, exists := bucket2Map[key]; exists {
				matches = append(matches, BucketMatch{
					Bucket1:     bucket1,
					Bucket2:     bucket2,
					MatchingKey: key,
				})
			}
		}
	}

	return matches
}

// BucketMatch represents a matching bucket between two parties
type BucketMatch struct {
	Bucket1     *BlockingBucket `json:"bucket1"`
	Bucket2     *BlockingBucket `json:"bucket2"`
	MatchingKey string          `json:"matching_key"`
}

// GetBlockingStats returns statistics about the blocking process
func (sb *SecureBlocker) GetBlockingStats(buckets []*BlockingBucket) BlockingStats {
	totalBuckets := len(buckets)
	totalRecords := 0
	bucketSizes := make([]int, len(buckets))

	for i, bucket := range buckets {
		size := len(bucket.RecordIDs)
		bucketSizes[i] = size
		totalRecords += size
	}

	sort.Ints(bucketSizes)

	var avgBucketSize float64
	if totalBuckets > 0 {
		avgBucketSize = float64(totalRecords) / float64(totalBuckets)
	}

	var medianBucketSize int
	if totalBuckets > 0 {
		medianBucketSize = bucketSizes[totalBuckets/2]
	}

	return BlockingStats{
		TotalBuckets:      totalBuckets,
		TotalRecords:      totalRecords,
		AverageBucketSize: avgBucketSize,
		MedianBucketSize:  medianBucketSize,
		MaxBucketSize:     bucketSizes[len(bucketSizes)-1],
		MinBucketSize:     bucketSizes[0],
	}
}

// BlockingStats contains statistics about the blocking process
type BlockingStats struct {
	TotalBuckets      int     `json:"total_buckets"`
	TotalRecords      int     `json:"total_records"`
	AverageBucketSize float64 `json:"average_bucket_size"`
	MedianBucketSize  int     `json:"median_bucket_size"`
	MaxBucketSize     int     `json:"max_bucket_size"`
	MinBucketSize     int     `json:"min_bucket_size"`
}
