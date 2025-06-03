// pipeline.go
// Package match provides the main pipeline orchestrator for the secure fuzzy matching system.
// It coordinates blocking, candidate generation, and fuzzy matching phases.
package match

import (
	"fmt"
	"log"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// PipelineConfig defines the configuration for the entire matching pipeline
type PipelineConfig struct {
	BlockingConfig   *BlockingConfig   `json:"blocking_config"`
	FuzzyMatchConfig *FuzzyMatchConfig `json:"fuzzy_match_config"`
	OutputPath       string            `json:"output_path"`
	EnableStats      bool              `json:"enable_stats"`
	MaxCandidates    int               `json:"max_candidates"` // Limit on candidate pairs
}

// Pipeline orchestrates the entire secure fuzzy matching process
type Pipeline struct {
	config  *PipelineConfig
	blocker *SecureBlocker
	matcher *FuzzyMatcher
	stats   *PipelineStats
	records map[string]*pprl.Record
}

// PipelineStats tracks statistics throughout the pipeline execution
type PipelineStats struct {
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	TotalRecords     int           `json:"total_records"`
	BlockingStats    BlockingStats `json:"blocking_stats"`
	MatchingStats    MatchingStats `json:"matching_stats"`
	CandidatePairs   int           `json:"candidate_pairs"`
	ProcessingTimeMs int64         `json:"processing_time_ms"`
}

// NewPipeline creates a new matching pipeline instance
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

// ExecuteMatching runs the complete matching pipeline
func (p *Pipeline) ExecuteMatching() ([]*MatchResult, error) {
	p.stats.StartTime = time.Now()
	defer func() {
		p.stats.EndTime = time.Now()
		p.stats.ProcessingTimeMs = p.stats.EndTime.Sub(p.stats.StartTime).Milliseconds()
	}()

	log.Println("Starting secure fuzzy matching pipeline...")

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

	// Phase 3: Perform fuzzy matching
	log.Println("Phase 3: Performing fuzzy matching...")
	results, err := p.performMatching(candidates)
	if err != nil {
		return nil, fmt.Errorf("fuzzy matching failed: %w", err)
	}

	// Generate statistics
	if p.config.EnableStats {
		p.generateStats(buckets, candidates, results)
	}

	log.Printf("Pipeline completed. Found %d matches from %d candidates",
		len(p.matcher.GetMatchingPairs(results)), len(candidates))

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

// performMatching executes the fuzzy matching phase
func (p *Pipeline) performMatching(candidates []CandidatePair) ([]*MatchResult, error) {
	results, err := p.matcher.BatchCompare(candidates, p.records)
	if err != nil {
		return nil, err
	}

	matches := p.matcher.GetMatchingPairs(results)
	log.Printf("Found %d matches from %d comparisons", len(matches), len(results))

	return results, nil
}

// generateStats compiles comprehensive statistics
func (p *Pipeline) generateStats(buckets []*BlockingBucket, candidates []CandidatePair, results []*MatchResult) {
	p.stats.BlockingStats = p.blocker.GetBlockingStats(buckets)
	p.stats.MatchingStats = p.matcher.GetMatchingStats(results)
	p.stats.CandidatePairs = len(candidates)
}

// GetStats returns the current pipeline statistics
func (p *Pipeline) GetStats() *PipelineStats {
	return p.stats
}

// SimulateTwoPartyMatching simulates the two-party matching protocol
func (p *Pipeline) SimulateTwoPartyMatching(otherPipeline *Pipeline) (*TwoPartyMatchResult, error) {
	log.Println("Simulating two-party secure matching protocol...")

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
	// Each party encrypts their buckets with the other's key
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

	// Phase 4: Perform secure fuzzy matching on candidate pairs
	var allResults []*MatchResult
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

				// Perform secure comparison
				result, err := p.matcher.CompareRecords(record1, record2)
				if err != nil {
					log.Printf("Failed to compare records %s and %s: %v", id1, id2, err)
					continue
				}

				result.BucketID = bucketMatch.MatchingKey
				allResults = append(allResults, result)
				totalCandidates++
			}
		}
	}

	matches := p.matcher.GetMatchingPairs(allResults)

	return &TwoPartyMatchResult{
		MatchingBuckets: len(matchingBuckets),
		CandidatePairs:  totalCandidates,
		TotalMatches:    len(matches),
		MatchResults:    allResults,
		Matches:         matches,
		Party1Records:   len(p.records),
		Party2Records:   len(otherPipeline.records),
	}, nil
}

// TwoPartyMatchResult contains the results of a two-party matching protocol
type TwoPartyMatchResult struct {
	MatchingBuckets int            `json:"matching_buckets"`
	CandidatePairs  int            `json:"candidate_pairs"`
	TotalMatches    int            `json:"total_matches"`
	MatchResults    []*MatchResult `json:"match_results"`
	Matches         []*MatchResult `json:"matches"`
	Party1Records   int            `json:"party1_records"`
	Party2Records   int            `json:"party2_records"`
}

// ExportResults exports match results to various formats
func (p *Pipeline) ExportResults(results []*MatchResult, format string) error {
	if p.config.OutputPath == "" {
		return fmt.Errorf("no output path configured")
	}

	switch format {
	case "json":
		return p.exportResultsAsJSON(results)
	case "csv":
		return p.exportResultsAsCSV(results)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportResultsAsJSON exports results in JSON format
func (p *Pipeline) exportResultsAsJSON(results []*MatchResult) error {
	// Implementation would write JSON to configured output path
	log.Printf("Exporting %d results as JSON to %s", len(results), p.config.OutputPath)
	return nil
}

// exportResultsAsCSV exports results in CSV format
func (p *Pipeline) exportResultsAsCSV(results []*MatchResult) error {
	// Implementation would write CSV to configured output path
	log.Printf("Exporting %d results as CSV to %s", len(results), p.config.OutputPath)
	return nil
}
