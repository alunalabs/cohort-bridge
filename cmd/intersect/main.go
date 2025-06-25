package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

type IntersectConfig struct {
	Dataset1         string  `json:"dataset1"`          // Path to first tokenized dataset
	Dataset2         string  `json:"dataset2"`          // Path to second tokenized dataset
	OutputFile       string  `json:"output_file"`       // Where to save intersection results
	HammingThreshold uint32  `json:"hamming_threshold"` // Maximum Hamming distance for match
	JaccardThreshold float64 `json:"jaccard_threshold"` // Minimum Jaccard similarity
	BatchSize        int     `json:"batch_size"`        // Processing batch size
	Streaming        bool    `json:"streaming"`         // Enable streaming mode
	ConfigFile       string  `json:"config_file"`       // Optional main config file
}

type IntersectionResult struct {
	ID1               string  `json:"id1"`
	ID2               string  `json:"id2"`
	IsMatch           bool    `json:"is_match"`
	HammingDistance   uint32  `json:"hamming_distance"`
	JaccardSimilarity float64 `json:"jaccard_similarity"`
	MatchScore        float64 `json:"match_score"`
	Timestamp         string  `json:"timestamp"`
}

func main() {
	fmt.Println("🔍 CohortBridge Intersection Finder")
	fmt.Println("====================================")
	fmt.Println("Find matches between tokenized datasets using PPRL techniques")
	fmt.Println()

	var (
		dataset1         = flag.String("dataset1", "", "Path to first tokenized dataset file")
		dataset2         = flag.String("dataset2", "", "Path to second tokenized dataset file")
		outputFile       = flag.String("output", "intersection_results.csv", "Output file for intersection results")
		configFile       = flag.String("config", "", "Optional configuration file")
		hammingThreshold = flag.Uint("hamming-threshold", 300, "Maximum Hamming distance for match")
		jaccardThreshold = flag.Float64("jaccard-threshold", 0.8, "Minimum Jaccard similarity")
		batchSize        = flag.Int("batch-size", 1000, "Processing batch size for streaming mode")
		streaming        = flag.Bool("streaming", false, "Enable streaming mode for large datasets")
		help             = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *dataset1 == "" || *dataset2 == "" {
		fmt.Println("❌ Error: Both dataset1 and dataset2 must be specified")
		fmt.Println()
		showHelp()
		os.Exit(1)
	}

	config := &IntersectConfig{
		Dataset1:         *dataset1,
		Dataset2:         *dataset2,
		OutputFile:       *outputFile,
		HammingThreshold: uint32(*hammingThreshold),
		JaccardThreshold: *jaccardThreshold,
		BatchSize:        *batchSize,
		Streaming:        *streaming,
		ConfigFile:       *configFile,
	}

	// Display configuration
	fmt.Printf("📁 Dataset 1: %s\n", config.Dataset1)
	fmt.Printf("📁 Dataset 2: %s\n", config.Dataset2)
	fmt.Printf("📊 Output: %s\n", config.OutputFile)
	fmt.Printf("🎯 Hamming threshold: %d\n", config.HammingThreshold)
	fmt.Printf("📈 Jaccard threshold: %.3f\n", config.JaccardThreshold)
	if config.Streaming {
		fmt.Printf("⚡ Streaming mode: enabled (batch size: %d)\n", config.BatchSize)
	}
	fmt.Println()

	if err := performIntersection(config); err != nil {
		log.Fatalf("❌ Intersection failed: %v", err)
	}
}

func performIntersection(config *IntersectConfig) error {
	start := time.Now()

	// Load tokenized datasets
	fmt.Println("📂 Loading tokenized datasets...")
	records1, err := loadTokenizedDatasetCSV(config.Dataset1)
	if err != nil {
		return fmt.Errorf("failed to load dataset1: %w", err)
	}
	fmt.Printf("   ✅ Loaded %d records from dataset1\n", len(records1))

	records2, err := loadTokenizedDatasetCSV(config.Dataset2)
	if err != nil {
		return fmt.Errorf("failed to load dataset2: %w", err)
	}
	fmt.Printf("   ✅ Loaded %d records from dataset2\n", len(records2))

	// Configure fuzzy matcher
	fuzzyConfig := &match.FuzzyMatchConfig{
		HammingThreshold:  config.HammingThreshold,
		JaccardThreshold:  config.JaccardThreshold,
		UseSecureProtocol: false,
	}

	matcher := match.NewFuzzyMatcher(fuzzyConfig)

	// Find intersection
	fmt.Println("🔄 Computing intersection...")
	var results []IntersectionResult
	totalComparisons := 0
	matchesFound := 0
	timestamp := time.Now().Format(time.RFC3339)

	if config.Streaming {
		// Stream processing for large datasets
		results, totalComparisons, matchesFound, err = performStreamingIntersection(
			records1, records2, matcher, timestamp, config.BatchSize)
	} else {
		// In-memory processing
		results, totalComparisons, matchesFound, err = performInMemoryIntersection(
			records1, records2, matcher, timestamp)
	}

	if err != nil {
		return fmt.Errorf("intersection computation failed: %w", err)
	}

	// Save results
	fmt.Println("💾 Saving intersection results...")
	if err := saveIntersectionResultsCSV(results, config.OutputFile); err != nil {
		return fmt.Errorf("failed to save results: %w", err)
	}

	elapsed := time.Since(start)
	fmt.Printf("📊 Results: %d comparisons, %d matches found in %v\n",
		totalComparisons, matchesFound, elapsed)

	return nil
}

func performInMemoryIntersection(records1, records2 []*pprl.Record, matcher *match.FuzzyMatcher, timestamp string) ([]IntersectionResult, int, int, error) {
	var results []IntersectionResult
	totalComparisons := 0
	matchesFound := 0

	for _, record1 := range records1 {
		for _, record2 := range records2 {
			totalComparisons++

			// Perform fuzzy matching
			matchResult, err := matcher.CompareRecords(record1, record2)
			if err != nil {
				continue // Skip comparison on error
			}

			// Convert to intersection result
			result := IntersectionResult{
				ID1:               matchResult.ID1,
				ID2:               matchResult.ID2,
				IsMatch:           matchResult.IsMatch,
				HammingDistance:   matchResult.HammingDistance,
				JaccardSimilarity: matchResult.JaccardSimilarity,
				MatchScore:        matchResult.MatchScore,
				Timestamp:         timestamp,
			}

			// Only store matches or high-scoring candidates
			if result.IsMatch || result.MatchScore > 0.8 {
				results = append(results, result)
			}

			if result.IsMatch {
				matchesFound++
			}

			// Progress reporting
			if totalComparisons%10000 == 0 {
				fmt.Printf("   Progress: %d comparisons, %d matches found\n",
					totalComparisons, matchesFound)
			}
		}
	}

	return results, totalComparisons, matchesFound, nil
}

func performStreamingIntersection(records1, records2 []*pprl.Record, matcher *match.FuzzyMatcher, timestamp string, batchSize int) ([]IntersectionResult, int, int, error) {
	var results []IntersectionResult
	totalComparisons := 0
	matchesFound := 0

	// Process in batches to manage memory
	for i := 0; i < len(records1); i += batchSize {
		end1 := i + batchSize
		if end1 > len(records1) {
			end1 = len(records1)
		}
		batch1 := records1[i:end1]

		for j := 0; j < len(records2); j += batchSize {
			end2 := j + batchSize
			if end2 > len(records2) {
				end2 = len(records2)
			}
			batch2 := records2[j:end2]

			// Process batch intersection
			batchResults, batchComparisons, batchMatches, err := performInMemoryIntersection(
				batch1, batch2, matcher, timestamp)
			if err != nil {
				return nil, 0, 0, err
			}

			results = append(results, batchResults...)
			totalComparisons += batchComparisons
			matchesFound += batchMatches

			fmt.Printf("   Batch [%d-%d] x [%d-%d]: %d comparisons, %d matches\n",
				i, end1-1, j, end2-1, batchComparisons, batchMatches)
		}
	}

	return results, totalComparisons, matchesFound, nil
}

func loadTokenizedDatasetCSV(filename string) ([]*pprl.Record, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("CSV file must have header and at least one data row")
	}

	// Expected header: id,bloom_filter,minhash,timestamp
	headers := rows[0]
	expectedHeaders := []string{"id", "bloom_filter", "minhash", "timestamp"}

	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[strings.ToLower(header)] = i
	}

	// Verify all required headers exist
	for _, required := range expectedHeaders {
		if _, exists := headerMap[required]; !exists {
			return nil, fmt.Errorf("missing required header: %s", required)
		}
	}

	// Convert to PPRL records
	var records []*pprl.Record
	for i, row := range rows[1:] { // Skip header row
		if len(row) < len(expectedHeaders) {
			return nil, fmt.Errorf("row %d has insufficient columns", i+2)
		}

		// Extract values using header map
		id := row[headerMap["id"]]
		bloomFilter := row[headerMap["bloom_filter"]]
		minHashStr := row[headerMap["minhash"]]

		// Parse MinHash if present
		var minHashSig []uint32
		if minHashStr != "" && minHashStr != "null" {
			if err := json.Unmarshal([]byte(minHashStr), &minHashSig); err != nil {
				return nil, fmt.Errorf("failed to decode MinHash for record %s: %w", id, err)
			}
		}

		record := &pprl.Record{
			ID:        id,
			BloomData: bloomFilter,
			MinHash:   minHashSig,
		}
		records = append(records, record)
	}

	return records, nil
}

func saveIntersectionResultsCSV(results []IntersectionResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id1", "id2", "is_match", "hamming_distance", "jaccard_similarity", "match_score", "timestamp"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, result := range results {
		row := []string{
			result.ID1,
			result.ID2,
			strconv.FormatBool(result.IsMatch),
			strconv.FormatUint(uint64(result.HammingDistance), 10),
			strconv.FormatFloat(result.JaccardSimilarity, 'f', 6, 64),
			strconv.FormatFloat(result.MatchScore, 'f', 6, 64),
			result.Timestamp,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func showHelp() {
	fmt.Println("🔍 CohortBridge Intersection Finder")
	fmt.Println("====================================")
	fmt.Println("Find matches between tokenized datasets using privacy-preserving record linkage")
	fmt.Println()

	fmt.Println("📋 USAGE:")
	fmt.Println("  intersect -dataset1=<file1> -dataset2=<file2> [options]")
	fmt.Println()

	fmt.Println("🔧 OPTIONS:")
	fmt.Println("  -dataset1 string     Path to first tokenized dataset file (required)")
	fmt.Println("  -dataset2 string     Path to second tokenized dataset file (required)")
	fmt.Println("  -output string       Output file for intersection results")
	fmt.Println("                       (default: intersection_results.csv)")
	fmt.Println("  -config string       Optional configuration file for advanced settings")
	fmt.Println("  -hamming-threshold   Maximum Hamming distance for match (default: 300)")
	fmt.Println("  -jaccard-threshold   Minimum Jaccard similarity for match (default: 0.8)")
	fmt.Println("  -batch-size int      Processing batch size for streaming (default: 1000)")
	fmt.Println("  -streaming           Enable streaming mode for large datasets")
	fmt.Println("  -help               Show this help message")
	fmt.Println()

	fmt.Println("💡 EXAMPLES:")
	fmt.Println("  # Basic intersection")
	fmt.Println("  intersect -dataset1=tokens_a.csv -dataset2=tokens_b.csv")
	fmt.Println()
	fmt.Println("  # Custom thresholds for stricter matching")
	fmt.Println("  intersect -dataset1=data1.csv -dataset2=data2.csv \\")
	fmt.Println("           -hamming-threshold=50 -jaccard-threshold=0.9")
	fmt.Println()
	fmt.Println("  # Streaming mode for large datasets")
	fmt.Println("  intersect -dataset1=large1.csv -dataset2=large2.csv \\")
	fmt.Println("           -streaming -batch-size=500")
	fmt.Println()

	fmt.Println("📄 INPUT FORMAT:")
	fmt.Println("  CSV files with headers: id,bloom_filter,minhash,timestamp")
	fmt.Println("  • id: Record identifier")
	fmt.Println("  • bloom_filter: Base64 encoded Bloom filter")
	fmt.Println("  • minhash: JSON array of MinHash signature")
	fmt.Println("  • timestamp: Tokenization timestamp")
	fmt.Println()

	fmt.Println("📊 OUTPUT FORMAT:")
	fmt.Println("  CSV file with match results:")
	fmt.Println("  • id1, id2: Record identifiers from both datasets")
	fmt.Println("  • is_match: Boolean indicating if records match")
	fmt.Println("  • hamming_distance: Bit differences between Bloom filters")
	fmt.Println("  • jaccard_similarity: Estimated similarity score (0.0-1.0)")
	fmt.Println("  • match_score: Composite matching score")
	fmt.Println("  • timestamp: When intersection was computed")
}
