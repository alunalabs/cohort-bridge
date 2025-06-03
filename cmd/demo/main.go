// main.go
// Demo application for the HIPAA-compliant, decentralized fuzzy matching system
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/auroradata-ai/cohort-bridge/internal/match"
)

func main() {
	// Command line flags
	var (
		mode        = flag.String("mode", "test", "Mode: test, single, or two-party")
		records1    = flag.Int("records1", 100, "Number of records in dataset 1")
		records2    = flag.Int("records2", 120, "Number of records in dataset 2")
		overlap     = flag.Float64("overlap", 0.3, "Overlap rate between datasets (0.0-1.0)")
		noise       = flag.Float64("noise", 0.1, "Noise rate for data corruption (0.0-1.0)")
		bloomSize   = flag.Int("bloom-size", 1024, "Bloom filter size in bits")
		hashCount   = flag.Int("hash-count", 8, "Number of hash functions for Bloom filter")
		minHashSigs = flag.Int("minhash-sigs", 64, "Number of MinHash signatures")
		hamming     = flag.Int("hamming-threshold", 100, "Hamming distance threshold for matching")
		jaccard     = flag.Float64("jaccard-threshold", 0.7, "Jaccard similarity threshold for matching")
		outputDir   = flag.String("output", "./test_output", "Output directory for test files")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
	)
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stdout)
	}

	switch *mode {
	case "test":
		runTestMode(*records1, *records2, *overlap, *noise,
			uint32(*bloomSize), uint32(*hashCount), uint32(*minHashSigs),
			uint32(*hamming), *jaccard, *outputDir)
	case "single":
		runSinglePartyMode(*outputDir)
	case "two-party":
		runTwoPartyMode(*outputDir)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

// runTestMode runs the comprehensive test harness
func runTestMode(records1, records2 int, overlap, noise float64,
	bloomSize, hashCount, minHashSigs, hamming uint32, jaccard float64, outputDir string) {

	fmt.Println("🧪 Running Secure Fuzzy Matching Test Harness")
	fmt.Println(repeatString("=", 50))

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Configure test harness
	testConfig := &match.TestConfig{
		NumRecords1:       records1,
		NumRecords2:       records2,
		OverlapRate:       overlap,
		NoiseRate:         noise,
		BloomFilterSize:   bloomSize,
		BloomHashCount:    hashCount,
		MinHashSignatures: minHashSigs,
		OutputDir:         outputDir,
	}

	// Configure pipeline
	pipelineConfig := &match.PipelineConfig{
		BlockingConfig: &match.BlockingConfig{
			MaxBucketsPerRecord: 10,
			SimilarityThreshold: 0.5,
		},
		FuzzyMatchConfig: &match.FuzzyMatchConfig{
			HammingThreshold:  hamming,
			JaccardThreshold:  jaccard,
			UseSecureProtocol: false, // Using placeholder for now
		},
		OutputPath:    outputDir + "/results.json",
		EnableStats:   true,
		MaxCandidates: 10000,
	}

	// Create and run test harness
	harness, err := match.NewTestHarness(testConfig)
	if err != nil {
		log.Fatalf("Failed to create test harness: %v", err)
	}

	fmt.Printf("📊 Test Configuration:\n")
	fmt.Printf("  Dataset 1: %d records\n", records1)
	fmt.Printf("  Dataset 2: %d records\n", records2)
	fmt.Printf("  Overlap rate: %.1f%%\n", overlap*100)
	fmt.Printf("  Noise rate: %.1f%%\n", noise*100)
	fmt.Printf("  Bloom filter: %d bits, %d hashes\n", bloomSize, hashCount)
	fmt.Printf("  MinHash signatures: %d\n", minHashSigs)
	fmt.Printf("  Hamming threshold: %d\n", hamming)
	fmt.Printf("  Jaccard threshold: %.2f\n", jaccard)
	fmt.Println()

	// Run the test
	results, err := harness.RunTest(pipelineConfig)
	if err != nil {
		log.Fatalf("Test failed: %v", err)
	}

	// Display results
	displayTestResults(results)

	// Save detailed results
	resultsFile := outputDir + "/detailed_results.json"
	if err := saveResultsToFile(results, resultsFile); err != nil {
		log.Printf("Failed to save results to file: %v", err)
	} else {
		fmt.Printf("📄 Detailed results saved to: %s\n", resultsFile)
	}
}

// runSinglePartyMode demonstrates single-party matching (for debugging)
func runSinglePartyMode(outputDir string) {
	fmt.Println("🔍 Running Single-Party Matching Demo")
	fmt.Println(repeatString("=", 40))
	fmt.Println("This mode demonstrates the matching pipeline on a single dataset")
	// Implementation would load existing data and run single-party matching
}

// runTwoPartyMode demonstrates two-party secure matching
func runTwoPartyMode(outputDir string) {
	fmt.Println("🤝 Running Two-Party Secure Matching Demo")
	fmt.Println(repeatString("=", 45))
	fmt.Println("This mode demonstrates secure matching between two separate parties")
	// Implementation would simulate real two-party protocol
}

// displayTestResults shows a comprehensive summary of test results
func displayTestResults(results *match.TestResults) {
	fmt.Println("🏆 Test Results Summary")
	fmt.Println(repeatString("=", 25))

	// Basic statistics
	fmt.Printf("📈 Matching Statistics:\n")
	fmt.Printf("  Ground truth matches: %d\n", results.GroundTruthCount)
	fmt.Printf("  Candidate pairs generated: %d\n", results.MatchResult.CandidatePairs)
	fmt.Printf("  Total matches found: %d\n", results.MatchResult.TotalMatches)
	fmt.Printf("  Matching buckets: %d\n", results.MatchResult.MatchingBuckets)
	fmt.Println()

	// Evaluation metrics
	eval := results.Evaluation
	fmt.Printf("🎯 Evaluation Metrics:\n")
	fmt.Printf("  True Positives: %d\n", eval.TruePositives)
	fmt.Printf("  False Positives: %d\n", eval.FalsePositives)
	fmt.Printf("  False Negatives: %d\n", eval.FalseNegatives)
	fmt.Printf("  Precision: %.3f\n", eval.Precision)
	fmt.Printf("  Recall: %.3f\n", eval.Recall)
	fmt.Printf("  F1-Score: %.3f\n", eval.F1Score)
	fmt.Println()

	// Performance statistics
	if results.Pipeline1Stats != nil {
		fmt.Printf("⚡ Performance:\n")
		fmt.Printf("  Processing time: %d ms\n", results.Pipeline1Stats.ProcessingTimeMs)
		fmt.Printf("  Records processed: %d\n", results.Pipeline1Stats.TotalRecords)
		if results.Pipeline1Stats.ProcessingTimeMs > 0 {
			throughput := float64(results.Pipeline1Stats.TotalRecords) / (float64(results.Pipeline1Stats.ProcessingTimeMs) / 1000.0)
			fmt.Printf("  Throughput: %.1f records/second\n", throughput)
		}
		fmt.Println()
	}

	// Blocking statistics
	blocking := results.Pipeline1Stats.BlockingStats
	fmt.Printf("🔧 Blocking Statistics:\n")
	fmt.Printf("  Total buckets: %d\n", blocking.TotalBuckets)
	fmt.Printf("  Average bucket size: %.1f\n", blocking.AverageBucketSize)
	fmt.Printf("  Median bucket size: %d\n", blocking.MedianBucketSize)
	fmt.Printf("  Max bucket size: %d\n", blocking.MaxBucketSize)
	fmt.Println()

	// Quality assessment
	assessQuality(eval)
}

// assessQuality provides a qualitative assessment of the matching results
func assessQuality(eval *match.Evaluation) {
	fmt.Printf("🔍 Quality Assessment:\n")

	var precision, recall, f1 string

	if eval.Precision >= 0.9 {
		precision = "Excellent"
	} else if eval.Precision >= 0.8 {
		precision = "Good"
	} else if eval.Precision >= 0.7 {
		precision = "Fair"
	} else {
		precision = "Poor"
	}

	if eval.Recall >= 0.9 {
		recall = "Excellent"
	} else if eval.Recall >= 0.8 {
		recall = "Good"
	} else if eval.Recall >= 0.7 {
		recall = "Fair"
	} else {
		recall = "Poor"
	}

	if eval.F1Score >= 0.9 {
		f1 = "Excellent"
	} else if eval.F1Score >= 0.8 {
		f1 = "Good"
	} else if eval.F1Score >= 0.7 {
		f1 = "Fair"
	} else {
		f1 = "Poor"
	}

	fmt.Printf("  Precision: %s (%.3f)\n", precision, eval.Precision)
	fmt.Printf("  Recall: %s (%.3f)\n", recall, eval.Recall)
	fmt.Printf("  Overall: %s (F1: %.3f)\n", f1, eval.F1Score)

	// Recommendations
	fmt.Printf("\n💡 Recommendations:\n")
	if eval.Precision < 0.8 {
		fmt.Printf("  • Consider increasing Hamming or Jaccard thresholds to reduce false positives\n")
	}
	if eval.Recall < 0.8 {
		fmt.Printf("  • Consider decreasing thresholds or improving blocking to reduce false negatives\n")
	}
	if eval.F1Score < 0.7 {
		fmt.Printf("  • Review noise rates and data quality\n")
		fmt.Printf("  • Consider tuning Bloom filter parameters\n")
	}
}

// saveResultsToFile saves detailed results to a JSON file
func saveResultsToFile(results *match.TestResults, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// Helper function for string repetition (Go doesn't have built-in string multiplication)
func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}
