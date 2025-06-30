package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// StreamingConfig defines configuration for streaming operations
type StreamingConfig struct {
	BatchSize         int     // Number of records to process in each batch
	MaxMemoryMB       int     // Maximum memory to use (in MB)
	EnableProgressLog bool    // Whether to log progress
	WriteBufferSize   int     // Buffer size for writing results
	HammingThreshold  uint32  // Hamming distance threshold for matches
	JaccardThreshold  float64 // Jaccard similarity threshold
}

// RecordBatch represents a batch of records for processing
type RecordBatch struct {
	Records []PatientRecord
	Offset  int
	Size    int
}

// MatchResultWriter handles streaming output of match results
type MatchResultWriter struct {
	file       *os.File
	writer     *csv.Writer
	jsonFile   *os.File
	jsonBuffer []byte
	isFirst    bool
	count      int
}

// StreamingRecordReader provides memory-efficient record reading
type StreamingRecordReader struct {
	csvDB        *db.Database
	batchSize    int
	offset       int
	fields       []string
	randomBits   float64
	totalRecords int
}

// NewStreamingRecordReader creates a new streaming record reader
func NewStreamingRecordReader(csvDB *db.Database, fields []string, batchSize int, randomBits float64) (*StreamingRecordReader, error) {
	return &StreamingRecordReader{
		csvDB:      csvDB,
		batchSize:  batchSize,
		offset:     0,
		fields:     fields,
		randomBits: randomBits,
	}, nil
}

// ReadBatch reads the next batch of records
func (r *StreamingRecordReader) ReadBatch() (*RecordBatch, error) {
	// Get records from database in batches
	rawRecords, err := (*r.csvDB).List(r.offset, r.batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read batch at offset %d: %w", r.offset, err)
	}

	if len(rawRecords) == 0 {
		return nil, io.EOF // No more records
	}

	// Convert raw records to PatientRecord format
	var records []PatientRecord
	sharedMinHash, err := GetGlobalMinHash()
	if err != nil {
		return nil, fmt.Errorf("failed to get global MinHash: %w", err)
	}

	for _, record := range rawRecords {
		// Create Bloom filter for this record
		bf := pprl.NewBloomFilterWithRandomBits(1000, 5, r.randomBits)

		// Create MinHash instance
		recordMinHash, err := recreateMinHashFromShared(sharedMinHash)
		if err != nil {
			return nil, fmt.Errorf("failed to create MinHash: %w", err)
		}

		// Add configured fields to Bloom filter using q-grams
		for _, field := range r.fields {
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
			return nil, fmt.Errorf("failed to compute MinHash signature: %w", err)
		}

		records = append(records, PatientRecord{
			ID:               record["id"],
			BloomFilter:      bf,
			MinHash:          recordMinHash,
			MinHashSignature: signature,
		})
	}

	batch := &RecordBatch{
		Records: records,
		Offset:  r.offset,
		Size:    len(records),
	}

	r.offset += len(rawRecords)
	return batch, nil
}

// HasMore returns true if there are more records to read
func (r *StreamingRecordReader) HasMore() bool {
	return true // We'll rely on ReadBatch returning io.EOF
}

// NewMatchResultWriter creates a new streaming result writer
func NewMatchResultWriter(timestamp, connID string) (*MatchResultWriter, error) {
	// Ensure output directory exists
	if err := EnsureOutputDirectory(); err != nil {
		return nil, fmt.Errorf("failed to ensure output directory: %w", err)
	}

	// Create CSV file
	csvFilename := fmt.Sprintf("out/matches_%s_%s.csv", timestamp, connID)
	csvFile, err := os.Create(csvFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file: %w", err)
	}

	writer := csv.NewWriter(csvFile)

	// Write CSV header
	header := []string{"Receiver_ID", "Sender_ID", "Match_Score", "Hamming_Distance", "Timestamp"}
	if err := writer.Write(header); err != nil {
		csvFile.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Create JSON file
	jsonFilename := fmt.Sprintf("out/matches_%s_%s.json", timestamp, connID)
	jsonFile, err := os.Create(jsonFilename)
	if err != nil {
		csvFile.Close()
		return nil, fmt.Errorf("failed to create JSON file: %w", err)
	}

	// Start JSON array
	if _, err := jsonFile.WriteString("[\n"); err != nil {
		csvFile.Close()
		jsonFile.Close()
		return nil, fmt.Errorf("failed to start JSON array: %w", err)
	}

	Info("Created streaming result writers: %s, %s", csvFilename, jsonFilename)

	return &MatchResultWriter{
		file:     csvFile,
		writer:   writer,
		jsonFile: jsonFile,
		isFirst:  true,
		count:    0,
	}, nil
}

// WriteMatch writes a single match result immediately
func (w *MatchResultWriter) WriteMatch(result *match.MatchResult) error {
	// Write to CSV
	timestamp := time.Now().UTC().Format(time.RFC3339)
	row := []string{
		result.ID1,
		result.ID2,
		fmt.Sprintf("%.6f", result.MatchScore),
		fmt.Sprintf("%d", result.HammingDistance),
		timestamp,
	}

	if err := w.writer.Write(row); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}

	// Flush immediately for streaming
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV: %w", err)
	}

	// Write to JSON
	if !w.isFirst {
		if _, err := w.jsonFile.WriteString(",\n"); err != nil {
			return fmt.Errorf("failed to write JSON separator: %w", err)
		}
	} else {
		w.isFirst = false
	}

	// Create JSON record
	jsonRecord := map[string]interface{}{
		"receiver_id":      result.ID1,
		"sender_id":        result.ID2,
		"match_score":      result.MatchScore,
		"hamming_distance": result.HammingDistance,
		"timestamp":        timestamp,
	}

	jsonData, err := json.MarshalIndent(jsonRecord, "  ", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if _, err := w.jsonFile.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write JSON: %w", err)
	}

	w.count++
	return nil
}

// Close finalizes and closes the result writers
func (w *MatchResultWriter) Close() error {
	var firstErr error

	// Close CSV
	if w.writer != nil {
		w.writer.Flush()
		if err := w.writer.Error(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("CSV writer error: %w", err)
		}
	}

	if w.file != nil {
		if err := w.file.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("CSV file close error: %w", err)
		}
	}

	// Close JSON
	if w.jsonFile != nil {
		// End JSON array
		if _, err := w.jsonFile.WriteString("\n]\n"); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("JSON end array error: %w", err)
		}

		if err := w.jsonFile.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("JSON file close error: %w", err)
		}
	}

	Info("Closed result writers with %d matches written", w.count)
	return firstErr
}

// GetCount returns the number of matches written
func (w *MatchResultWriter) GetCount() int {
	return w.count
}

// StreamingMatcher provides memory-efficient matching operations
type StreamingMatcher struct {
	config           *StreamingConfig
	resultWriter     *MatchResultWriter
	totalComparisons int
	totalMatches     int
}

// NewStreamingMatcher creates a new streaming matcher
func NewStreamingMatcher(config *StreamingConfig, resultWriter *MatchResultWriter) *StreamingMatcher {
	return &StreamingMatcher{
		config:       config,
		resultWriter: resultWriter,
	}
}

// MatchBatchAgainstSender performs streaming matching of a receiver batch against sender data
func (m *StreamingMatcher) MatchBatchAgainstSender(
	receiverBatch *RecordBatch,
	senderData MatchingData,
	connID string,
) error {

	batchStart := time.Now()
	batchComparisons := 0
	batchMatches := 0

	for _, receiverRecord := range receiverBatch.Records {
		for senderID, senderBloomData := range senderData.Records {
			batchComparisons++
			m.totalComparisons++

			// Decode sender Bloom filter
			senderBF, err := pprl.BloomFromBase64(senderBloomData)
			if err != nil {
				Debug("Failed to decode sender Bloom filter: %v", err)
				continue
			}

			// Calculate Hamming distance
			hammingDist, err := receiverRecord.BloomFilter.HammingDistance(senderBF)
			if err != nil {
				Debug("Failed to calculate Hamming distance: %v", err)
				continue
			}

			// Calculate match score using PROVEN working method from validate command
			bfSize := receiverRecord.BloomFilter.GetSize()
			matchScore := 1.0
			if hammingDist > 0 {
				matchScore = 1.0 - (float64(hammingDist) / float64(bfSize))
			}

			// Calculate Jaccard similarity for additional scoring
			var jaccardSim float64
			// For Bloom filters, use Hamming-based approximation for consistency
			jaccardSim = 1.0 - (float64(hammingDist) / float64(bfSize))

			// Determine if this is a match based on Hamming threshold ONLY (same as validate command)
			// This is the KEY difference - validate command uses ONLY Hamming distance
			isMatch := hammingDist <= m.config.HammingThreshold

			// Create match result
			result := &match.MatchResult{
				ID1:               receiverRecord.ID,
				ID2:               senderID,
				MatchScore:        matchScore, // Use normalized 0-1 score
				JaccardSimilarity: jaccardSim, // For compatibility
				HammingDistance:   hammingDist,
				IsMatch:           isMatch,
			}

			// Write ALL matches that meet the threshold (same as validate)
			if isMatch {
				// Write result immediately (streaming)
				if err := m.resultWriter.WriteMatch(result); err != nil {
					return fmt.Errorf("failed to write match result: %w", err)
				}

				batchMatches++
				m.totalMatches++

				// Debug first few matches found
				if m.totalMatches <= 5 {
					Debug("Match #%d: %s <-> %s (Hamming: %d ≤ %d, Score: %.6f)",
						m.totalMatches, receiverRecord.ID, senderID, hammingDist, m.config.HammingThreshold, matchScore)
				}
			}
		}
	}

	duration := time.Since(batchStart)
	if m.config.EnableProgressLog {
		Info("Batch %d: %d comparisons, %d matches in %v (avg: %v/comparison) - using Hamming <= %d",
			receiverBatch.Offset/m.config.BatchSize+1,
			batchComparisons, batchMatches, duration,
			duration/time.Duration(batchComparisons), m.config.HammingThreshold)
	}

	return nil
}

// GetStats returns current matching statistics
func (m *StreamingMatcher) GetStats() (int, int) {
	return m.totalComparisons, m.totalMatches
}
