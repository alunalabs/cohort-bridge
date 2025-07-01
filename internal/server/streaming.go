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

// StreamingConfig defines zero-knowledge streaming configuration
// All similarity thresholds are hardcoded for security
type StreamingConfig struct {
	BatchSize         int  // Number of records to process in each batch
	MaxMemoryMB       int  // Maximum memory to use (in MB)
	EnableProgressLog bool // Whether to log progress
	WriteBufferSize   int  // Buffer size for writing results
	Party             int  // Party number for zero-knowledge protocol (0 or 1)
	// NO configurable thresholds - all hardcoded for security
}

// RecordBatch represents a batch of PPRL records for zero-knowledge processing
type RecordBatch struct {
	Records []*pprl.Record
	Offset  int
	Size    int
}

// ZKMatchResultWriter handles streaming output of ONLY matches (zero information leakage)
type ZKMatchResultWriter struct {
	file       *os.File
	writer     *csv.Writer
	jsonFile   *os.File
	jsonBuffer []byte
	isFirst    bool
	count      int
}

// StreamingRecordReader provides memory-efficient record reading for zero-knowledge processing
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

// ReadBatch reads the next batch of records and converts to PPRL format
func (r *StreamingRecordReader) ReadBatch() (*RecordBatch, error) {
	// Get records from database in batches
	rawRecords, err := (*r.csvDB).List(r.offset, r.batchSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read batch at offset %d: %w", r.offset, err)
	}

	if len(rawRecords) == 0 {
		return nil, io.EOF // No more records
	}

	// Convert raw records to PPRL Record format
	var records []*pprl.Record
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

		// Encode Bloom filter to base64
		bloomData, err := bf.ToBase64()
		if err != nil {
			return nil, fmt.Errorf("failed to encode Bloom filter: %w", err)
		}

		records = append(records, &pprl.Record{
			ID:        record["id"],
			BloomData: bloomData,
			MinHash:   signature,
			QGramData: "", // Not used in streaming
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

// NewZKMatchResultWriter creates a new zero-knowledge streaming result writer
// ONLY writes matches - no other information
func NewZKMatchResultWriter(timestamp, connID string) (*ZKMatchResultWriter, error) {
	// Ensure output directory exists
	if err := EnsureOutputDirectory(); err != nil {
		return nil, fmt.Errorf("failed to ensure output directory: %w", err)
	}

	// Create CSV file for ONLY matches
	csvFilename := fmt.Sprintf("out/zk_matches_%s_%s.csv", timestamp, connID)
	csvFile, err := os.Create(csvFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create CSV file: %w", err)
	}

	writer := csv.NewWriter(csvFile)

	// Write CSV header - ONLY essential match information
	header := []string{"Local_ID", "Peer_ID", "Timestamp"}
	if err := writer.Write(header); err != nil {
		csvFile.Close()
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Create JSON file
	jsonFilename := fmt.Sprintf("out/zk_matches_%s_%s.json", timestamp, connID)
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

	Info("Created zero-knowledge streaming result writers: %s, %s", csvFilename, jsonFilename)

	return &ZKMatchResultWriter{
		file:     csvFile,
		writer:   writer,
		jsonFile: jsonFile,
		isFirst:  true,
		count:    0,
	}, nil
}

// WriteMatch writes a single zero-knowledge match result (ONLY if it matches)
func (w *ZKMatchResultWriter) WriteMatch(result *match.PrivateMatchResult) error {
	// Write to CSV - ONLY the matching pair
	timestamp := time.Now().UTC().Format(time.RFC3339)
	row := []string{
		result.LocalID,
		result.PeerID,
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

	// Write to JSON - ONLY essential match information
	if !w.isFirst {
		if _, err := w.jsonFile.WriteString(",\n"); err != nil {
			return fmt.Errorf("failed to write JSON separator: %w", err)
		}
	} else {
		w.isFirst = false
	}

	// Create JSON record with ZERO information leakage
	jsonRecord := map[string]interface{}{
		"local_id":  result.LocalID,
		"peer_id":   result.PeerID,
		"timestamp": timestamp,
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

// Close finalizes and closes all output files
func (w *ZKMatchResultWriter) Close() error {
	// Close CSV
	w.writer.Flush()
	if err := w.writer.Error(); err != nil {
		return fmt.Errorf("failed to flush CSV on close: %w", err)
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("failed to close CSV file: %w", err)
	}

	// Close JSON array and file
	if _, err := w.jsonFile.WriteString("\n]"); err != nil {
		return fmt.Errorf("failed to close JSON array: %w", err)
	}
	if err := w.jsonFile.Close(); err != nil {
		return fmt.Errorf("failed to close JSON file: %w", err)
	}

	Info("Closed result files. Total matches written: %d", w.count)
	return nil
}

// GetCount returns the number of matches written (ONLY information revealed)
func (w *ZKMatchResultWriter) GetCount() int {
	return w.count
}

// ZKStreamingMatcher performs zero-knowledge streaming matching
type ZKStreamingMatcher struct {
	config       *StreamingConfig
	resultWriter *ZKMatchResultWriter
	fuzzyMatcher *match.FuzzyMatcher
	matchCount   int // ONLY count matches - no other statistics
}

// NewZKStreamingMatcher creates a new zero-knowledge streaming matcher
func NewZKStreamingMatcher(config *StreamingConfig, resultWriter *ZKMatchResultWriter) *ZKStreamingMatcher {
	// Configure zero-knowledge fuzzy matcher
	fuzzyConfig := &match.FuzzyMatchConfig{
		Party: config.Party,
	}

	return &ZKStreamingMatcher{
		config:       config,
		resultWriter: resultWriter,
		fuzzyMatcher: match.NewFuzzyMatcher(fuzzyConfig),
		matchCount:   0,
	}
}

// MatchBatchAgainstPeer performs zero-knowledge matching between batches
func (m *ZKStreamingMatcher) MatchBatchAgainstPeer(
	localBatch *RecordBatch,
	peerRecords []*pprl.Record,
	connID string,
) error {

	for _, localRecord := range localBatch.Records {
		for _, peerRecord := range peerRecords {
			// Perform zero-knowledge comparison
			result, err := m.fuzzyMatcher.CompareRecords(localRecord, peerRecord)
			if err != nil {
				continue // Continue processing - no error information leaked
			}

			// Only write if it's a match (result will be nil for non-matches)
			if result != nil {
				if err := m.resultWriter.WriteMatch(result); err != nil {
					return fmt.Errorf("failed to write match result: %w", err)
				}
				m.matchCount++
			}
		}
	}

	// Log minimal progress information (no dataset details)
	if m.config.EnableProgressLog {
		Info("Processed batch with %d local records against %d peer records",
			len(localBatch.Records), len(peerRecords))
	}

	return nil
}

// GetMatchCount returns ONLY the number of matches found (no other statistics)
func (m *ZKStreamingMatcher) GetMatchCount() int {
	return m.matchCount
}
