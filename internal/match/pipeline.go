// pipeline.go
// Package match provides the main pipeline orchestrator for the secure fuzzy matching system.
// It coordinates blocking, candidate generation, and fuzzy matching phases.
package match

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// PipelineConfig defines the configuration for the matching pipeline
// Updated to use zero-knowledge protocols only
type PipelineConfig struct {
	BlockingConfig   *BlockingConfig   `json:"blocking_config"`
	FuzzyMatchConfig *FuzzyMatchConfig `json:"fuzzy_match_config"`
	OutputPath       string            `json:"output_path"`
	EnableStats      bool              `json:"enable_stats"`   // Limited stats only
	MaxCandidates    int               `json:"max_candidates"` // Limit on candidate pairs
}

// Pipeline orchestrates the complete zero-knowledge matching process
type Pipeline struct {
	config  *PipelineConfig
	blocker *SecureBlocker
	matcher *FuzzyMatcher
	stats   *PipelineStats
	records map[string]*pprl.Record
}

// PipelineStats tracks LIMITED statistics with no information leakage
type PipelineStats struct {
	StartTime        time.Time            `json:"start_time"`
	EndTime          time.Time            `json:"end_time"`
	TotalRecords     int                  `json:"total_records"`
	BlockingStats    BlockingStats        `json:"blocking_stats"`
	MatchingStats    PrivateMatchingStats `json:"matching_stats"`
	CandidatePairs   int                  `json:"candidate_pairs"`
	ProcessingTimeMs int64                `json:"processing_time_ms"`
}

// NewPipeline creates a new zero-knowledge matching pipeline instance
func NewPipeline(config *PipelineConfig) (*Pipeline, error) {
	blocker, err := NewSecureBlocker(config.BlockingConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create secure blocker: %w", err)
	}

	matcher := NewFuzzyMatcher(config.FuzzyMatchConfig)

	return &Pipeline{
		config:  config,
		blocker: blocker,
		matcher: matcher,
		stats:   &PipelineStats{},
		records: make(map[string]*pprl.Record),
	}, nil
}

// LoadRecords loads records from storage into the pipeline
func (p *Pipeline) LoadRecords(storage *pprl.Storage) error {
	records, err := storage.LoadAll()
	if err != nil {
		return fmt.Errorf("failed to load records: %w", err)
	}

	p.records = make(map[string]*pprl.Record, len(records))
	for _, record := range records {
		p.records[record.ID] = record
	}

	p.stats.TotalRecords = len(records)
	log.Printf("Loaded %d records into pipeline", len(records))
	return nil
}

// ExecuteMatching runs the complete zero-knowledge matching pipeline
func (p *Pipeline) ExecuteMatching() ([]*PrivateMatchResult, error) {
	p.stats.StartTime = time.Now()
	defer func() {
		p.stats.EndTime = time.Now()
		p.stats.ProcessingTimeMs = p.stats.EndTime.Sub(p.stats.StartTime).Milliseconds()
	}()

	log.Println("Starting zero-knowledge fuzzy matching pipeline...")

	// Phase 1: Create blocking buckets
	log.Println("Phase 1: Creating secure blocking buckets...")
	buckets, err := p.createBlocks()
	if err != nil {
		return nil, fmt.Errorf("blocking phase failed: %w", err)
	}

	// Phase 2: Generate candidate pairs
	log.Println("Phase 2: Generating candidate pairs...")
	candidates, err := p.generateCandidates(buckets)
	if err != nil {
		return nil, fmt.Errorf("candidate generation failed: %w", err)
	}

	// Phase 3: Perform zero-knowledge fuzzy matching
	log.Println("Phase 3: Performing zero-knowledge fuzzy matching...")
	results, err := p.performZKMatching(candidates)
	if err != nil {
		return nil, fmt.Errorf("zero-knowledge matching failed: %w", err)
	}

	// Generate LIMITED statistics (no information leakage)
	if p.config.EnableStats {
		p.generateLimitedStats(buckets, candidates, results)
	}

	log.Printf("Pipeline completed. Found %d matches from %d candidates", len(results), len(candidates))
	return results, nil
}

// createBlocks implements the secure blocking phase
func (p *Pipeline) createBlocks() ([]*BlockingBucket, error) {
	// Convert records to MinHash format
	var recordsWithMinHash []RecordWithMinHash
	for id, record := range p.records {
		recordsWithMinHash = append(recordsWithMinHash, RecordWithMinHash{
			ID:      id,
			MinHash: record.MinHash,
		})
	}

	// Create secure blocking buckets
	buckets, err := p.blocker.CreateBlocks(recordsWithMinHash)
	if err != nil {
		return nil, err
	}

	log.Printf("Created %d blocking buckets", len(buckets))
	return buckets, nil
}

// generateCandidates creates candidate pairs from blocking buckets
func (p *Pipeline) generateCandidates(buckets []*BlockingBucket) ([]CandidatePair, error) {
	candidates := p.blocker.GetCandidatePairs(buckets)

	// Apply candidate limit if configured
	if p.config.MaxCandidates > 0 && len(candidates) > p.config.MaxCandidates {
		log.Printf("Limiting candidates from %d to %d", len(candidates), p.config.MaxCandidates)
		candidates = candidates[:p.config.MaxCandidates]
	}

	p.stats.CandidatePairs = len(candidates)
	log.Printf("Generated %d candidate pairs", len(candidates))
	return candidates, nil
}

// performZKMatching executes the zero-knowledge fuzzy matching phase
func (p *Pipeline) performZKMatching(candidates []CandidatePair) ([]*PrivateMatchResult, error) {
	results, err := p.matcher.BatchPrivateCompare(candidates, p.records)
	if err != nil {
		return nil, err
	}

	log.Printf("Found %d matches from %d comparisons", len(results), len(candidates))
	return results, nil
}

// generateLimitedStats compiles LIMITED statistics with no information leakage
func (p *Pipeline) generateLimitedStats(buckets []*BlockingBucket, candidates []CandidatePair, results []*PrivateMatchResult) {
	p.stats.BlockingStats = p.blocker.GetBlockingStats(buckets)
	p.stats.MatchingStats = p.matcher.GetPrivateMatchingStats(results)
	p.stats.CandidatePairs = len(candidates)
}

// GetStats returns the current pipeline statistics
func (p *Pipeline) GetStats() *PipelineStats {
	return p.stats
}

// SimulateTwoPartyMatching simulates the two-party zero-knowledge matching protocol
func (p *Pipeline) SimulateTwoPartyMatching(otherPipeline *Pipeline) (*TwoPartyMatchResult, error) {
	log.Println("Simulating two-party zero-knowledge matching protocol...")

	// Phase 1: Create local blocking buckets
	myBuckets, err := p.createBlocks()
	if err != nil {
		return nil, fmt.Errorf("failed to create local blocks: %w", err)
	}

	theirBuckets, err := otherPipeline.createBlocks()
	if err != nil {
		return nil, fmt.Errorf("failed to create remote blocks: %w", err)
	}

	// Phase 2: Exchange encrypted bucket information
	myDoubleEncrypted, err := p.blocker.ExchangeEncryptedBuckets(myBuckets, otherPipeline.blocker.key)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange encrypted buckets: %w", err)
	}

	theirDoubleEncrypted, err := otherPipeline.blocker.ExchangeEncryptedBuckets(theirBuckets, p.blocker.key)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange encrypted buckets: %w", err)
	}

	// Phase 3: Find matching buckets (intersection)
	matchingBuckets := FindMatchingBuckets(myDoubleEncrypted, theirDoubleEncrypted)
	log.Printf("Found %d matching buckets between parties", len(matchingBuckets))

	// Phase 4: Perform zero-knowledge fuzzy matching on candidate pairs
	var allResults []*PrivateMatchResult
	totalCandidates := 0

	for _, bucketMatch := range matchingBuckets {
		// Generate candidate pairs from matching buckets
		for _, id1 := range bucketMatch.Bucket1.RecordIDs {
			for _, id2 := range bucketMatch.Bucket2.RecordIDs {
				record1, exists1 := p.records[id1]
				record2, exists2 := otherPipeline.records[id2]

				if !exists1 || !exists2 {
					continue
				}

				// Perform zero-knowledge comparison
				result, err := p.matcher.CompareRecords(record1, record2)
				if err != nil {
					log.Printf("Failed to compare records %s and %s: %v", id1, id2, err)
					continue
				}

				// Only add if it's a match (result will be nil for non-matches)
				if result != nil {
					allResults = append(allResults, result)
				}
				totalCandidates++
			}
		}
	}

	// Filter to get only matches (redundant since we already filter above)
	matches := p.matcher.GetPrivateMatches(allResults)

	log.Printf("Two-party matching completed: %d matches from %d candidates across %d matching buckets",
		len(matches), totalCandidates, len(matchingBuckets))

	return &TwoPartyMatchResult{
		MatchingBuckets: len(matchingBuckets),
		CandidatePairs:  totalCandidates,
		TotalMatches:    len(matches),
		PrivateMatches:  matches, // Use new field name
		Party1Records:   len(p.records),
		Party2Records:   len(otherPipeline.records),
	}, nil
}

// TwoPartyMatchResult represents the result of zero-knowledge two-party matching
type TwoPartyMatchResult struct {
	MatchingBuckets int                   `json:"matching_buckets"`
	CandidatePairs  int                   `json:"candidate_pairs"`
	TotalMatches    int                   `json:"total_matches"`
	PrivateMatches  []*PrivateMatchResult `json:"private_matches"` // ONLY matches, no other info
	Party1Records   int                   `json:"party1_records"`
	Party2Records   int                   `json:"party2_records"`
}

// ExportResults exports zero-knowledge results to file (ONLY matches)
func (p *Pipeline) ExportResults(results []*PrivateMatchResult, format string) error {
	if format == "json" {
		return p.exportResultsAsJSON(results)
	} else if format == "csv" {
		return p.exportResultsAsCSV(results)
	}
	return fmt.Errorf("unsupported export format: %s", format)
}

// exportResultsAsJSON exports results as JSON with ZERO information leakage
func (p *Pipeline) exportResultsAsJSON(results []*PrivateMatchResult) error {
	file, err := os.Create(p.config.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	// Export ONLY matches - no additional metadata
	return encoder.Encode(map[string]interface{}{
		"matches": results,
		"count":   len(results), // ONLY count, no other statistics
	})
}

// exportResultsAsCSV exports results as CSV with ZERO information leakage
func (p *Pipeline) exportResultsAsCSV(results []*PrivateMatchResult) error {
	file, err := os.Create(p.config.OutputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header - ONLY the essential match information
	if err := writer.Write([]string{"local_id", "peer_id"}); err != nil {
		return err
	}

	// Write ONLY the matching pairs - no scores, distances, or metadata
	for _, result := range results {
		if err := writer.Write([]string{result.LocalID, result.PeerID}); err != nil {
			return err
		}
	}

	return nil
}
