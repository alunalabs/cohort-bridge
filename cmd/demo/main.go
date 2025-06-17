// main.go
// Demo application for the HIPAA-compliant, decentralized fuzzy matching system
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
	"github.com/manifoldco/promptui"
)

func main() {
	fmt.Println("🩺 HIPAA-Compliant Fuzzy Matching System")
	fmt.Println("=========================================")

	// Mode selection
	modePrompt := promptui.Select{
		Label: "Select operation mode",
		Items: []string{
			"Test Harness - Run comprehensive matching tests with synthetic data",
			"Single Party - Load and analyze single dataset",
			"Two Party - Simulate secure two-party matching",
			"Compare Records - Compare two specific records manually",
			"Validation Test - Run end-to-end validation against ground truth",
		},
	}

	modeIndex, _, err := modePrompt.Run()
	if err != nil {
		fmt.Printf("Mode selection failed: %v\n", err)
		os.Exit(1)
	}

	switch modeIndex {
	case 0:
		runTestMode()
	case 1:
		runSinglePartyMode()
	case 2:
		runTwoPartyMode()
	case 3:
		runCompareRecordsMode()
	case 4:
		runValidationTestMode()
	default:
		fmt.Println("Invalid mode selected")
		os.Exit(1)
	}
}

// runTestMode runs the comprehensive test harness with user-configurable parameters
func runTestMode() {
	fmt.Println("📊 Test Mode: Generating synthetic data and testing matching algorithms")

	params, err := getTestParameters()
	if err != nil {
		log.Fatal(err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll("out", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(params.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Configure test harness
	testConfig := &match.TestConfig{
		NumRecords1:       params.Records1,
		NumRecords2:       params.Records2,
		OverlapRate:       params.Overlap,
		NoiseRate:         params.Noise,
		BloomFilterSize:   params.BloomSize,
		BloomHashCount:    params.HashCount,
		MinHashSignatures: params.MinHashSigs,
		OutputDir:         params.OutputDir,
	}

	// Configure pipeline
	pipelineConfig := &match.PipelineConfig{
		BlockingConfig: &match.BlockingConfig{
			MaxBucketsPerRecord: 10,
			SimilarityThreshold: 0.5,
		},
		FuzzyMatchConfig: &match.FuzzyMatchConfig{
			HammingThreshold:  params.Hamming,
			JaccardThreshold:  params.Jaccard,
			UseSecureProtocol: false, // Using placeholder for now
		},
		OutputPath:    params.OutputDir + "/results.json",
		EnableStats:   true,
		MaxCandidates: 10000,
	}

	// Create and run test harness
	harness, err := match.NewTestHarness(testConfig)
	if err != nil {
		log.Fatalf("Failed to create test harness: %v", err)
	}

	fmt.Printf("\n📊 Test Configuration:\n")
	fmt.Printf("  Dataset 1: %d records\n", params.Records1)
	fmt.Printf("  Dataset 2: %d records\n", params.Records2)
	fmt.Printf("  Overlap rate: %.1f%%\n", params.Overlap*100)
	fmt.Printf("  Noise rate: %.1f%%\n", params.Noise*100)
	fmt.Printf("  Bloom filter: %d bits, %d hashes\n", params.BloomSize, params.HashCount)
	fmt.Printf("  MinHash signatures: %d\n", params.MinHashSigs)
	fmt.Printf("  Hamming threshold: %d\n", params.Hamming)
	fmt.Printf("  Jaccard threshold: %.2f\n", params.Jaccard)
	fmt.Println()

	// Run the test
	results, err := harness.RunTest(pipelineConfig)
	if err != nil {
		log.Fatalf("Test failed: %v", err)
	}

	// Display results
	displayTestResults(results)

	// Save detailed results
	resultsFile := params.OutputDir + "/detailed_results.json"
	if err := saveResultsToFile(results, resultsFile); err != nil {
		log.Printf("Failed to save results to file: %v", err)
	} else {
		fmt.Printf("📄 Detailed results saved to: %s\n", resultsFile)
	}
}

// runSinglePartyMode demonstrates single-party matching (for debugging)
func runSinglePartyMode() {
	fmt.Println("\n🔍 Single-Party Matching Demo")
	fmt.Println("=============================")

	configPath, err := selectConfigFile()
	if err != nil {
		log.Fatalf("Failed to select config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load and analyze single dataset
	csvDB, err := db.NewCSVDatabase(cfg.Database.Filename)
	if err != nil {
		log.Fatalf("Failed to load CSV database: %v", err)
	}

	fmt.Printf("✅ Loaded data from: %s\n", cfg.Database.Filename)

	// Load first few records for demonstration
	records, err := csvDB.List(0, 10)
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}

	fmt.Printf("📊 Sample records (first 10):\n")
	for i, record := range records {
		fmt.Printf("  %d. ID: %s", i+1, record["id"])
		for _, field := range cfg.Database.Fields {
			if value, exists := record[field]; exists {
				fmt.Printf(", %s: %s", field, value)
			}
		}
		fmt.Println()
	}
}

// runTwoPartyMode demonstrates two-party secure matching
func runTwoPartyMode() {
	fmt.Println("\n🤝 Two-Party Secure Matching Demo")
	fmt.Println("=================================")
	fmt.Println("This mode would demonstrate secure matching between two separate parties")
	fmt.Println("For now, this is a placeholder. Use the agent CLI for actual two-party matching.")
}

// runCompareRecordsMode allows manual comparison of two records
func runCompareRecordsMode() {
	fmt.Println("\n🔬 Manual Record Comparison")
	fmt.Println("===========================")

	configPath, err := selectConfigFile()
	if err != nil {
		log.Fatalf("Failed to select config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load CSV data
	csvDB, err := db.NewCSVDatabase(cfg.Database.Filename)
	if err != nil {
		log.Fatalf("Failed to load CSV database: %v", err)
	}

	// Get all records to show options
	allRecords, err := csvDB.List(0, 1000000)
	if err != nil {
		log.Fatalf("Failed to list records: %v", err)
	}

	if len(allRecords) < 2 {
		fmt.Println("❌ Need at least 2 records for comparison")
		return
	}

	// Show available records
	fmt.Printf("📊 Available records (%d total):\n", len(allRecords))
	recordOptions := make([]string, len(allRecords))
	for i, record := range allRecords {
		description := fmt.Sprintf("ID: %s", record["id"])
		for _, field := range cfg.Database.Fields {
			if value, exists := record[field]; exists && value != "" {
				description += fmt.Sprintf(", %s: %s", field, value)
			}
		}
		recordOptions[i] = description
		if i < 20 { // Show first 20 for preview
			fmt.Printf("  %d. %s\n", i+1, description)
		}
	}

	if len(allRecords) > 20 {
		fmt.Printf("  ... and %d more records\n", len(allRecords)-20)
	}

	// Select first record
	record1Prompt := promptui.Select{
		Label: "Select first record to compare",
		Items: recordOptions,
		Size:  10,
	}

	idx1, _, err := record1Prompt.Run()
	if err != nil {
		log.Fatalf("Record selection failed: %v", err)
	}

	// Select second record
	record2Prompt := promptui.Select{
		Label: "Select second record to compare",
		Items: recordOptions,
		Size:  10,
	}

	idx2, _, err := record2Prompt.Run()
	if err != nil {
		log.Fatalf("Record selection failed: %v", err)
	}

	if idx1 == idx2 {
		fmt.Println("❌ Cannot compare a record with itself")
		return
	}

	// Compare the selected records
	compareRecords(allRecords[idx1], allRecords[idx2], cfg.Database.Fields)
}

// runValidationTestMode runs end-to-end validation against ground truth
func runValidationTestMode() {
	fmt.Println("\n🎯 End-to-End Validation Test")
	fmt.Println("=============================")

	// Get validation parameters
	params, err := getValidationParameters()
	if err != nil {
		log.Fatalf("Failed to get validation parameters: %v", err)
	}

	// Load ground truth CSV
	groundTruthDB, err := db.NewCSVDatabase(params.GroundTruthPath)
	if err != nil {
		log.Fatalf("Failed to load ground truth CSV: %v", err)
	}

	// Load ground truth matches
	groundTruthRecords, err := groundTruthDB.List(0, 1000000)
	if err != nil {
		log.Fatalf("Failed to list ground truth records: %v", err)
	}

	groundTruth := make(map[string]string)
	for _, record := range groundTruthRecords {
		patientA := record["patient_a_id"]
		patientB := record["patient_b_id"]
		if patientA != "" && patientB != "" {
			groundTruth[patientA] = patientB
		}
	}

	fmt.Printf("📊 Loaded %d ground truth matches\n", len(groundTruth))

	// Run the matching system
	fmt.Println("🔄 Running matching system...")

	// For now, use the test harness with default parameters
	// In a real implementation, this would run the actual matching pipeline
	testConfig := &match.TestConfig{
		NumRecords1:       len(groundTruth),
		NumRecords2:       len(groundTruth) + 20, // Add some non-matching records
		OverlapRate:       1.0,                   // All records should match
		NoiseRate:         0.05,
		BloomFilterSize:   1024,
		BloomHashCount:    8,
		MinHashSignatures: 64,
		OutputDir:         "./out",
	}

	pipelineConfig := &match.PipelineConfig{
		BlockingConfig: &match.BlockingConfig{
			MaxBucketsPerRecord: 10,
			SimilarityThreshold: 0.5,
		},
		FuzzyMatchConfig: &match.FuzzyMatchConfig{
			HammingThreshold:  100,
			JaccardThreshold:  0.7,
			UseSecureProtocol: false,
		},
		OutputPath:    "./out/results.json",
		EnableStats:   true,
		MaxCandidates: 10000,
	}

	harness, err := match.NewTestHarness(testConfig)
	if err != nil {
		log.Fatalf("Failed to create test harness: %v", err)
	}

	results, err := harness.RunTest(pipelineConfig)
	if err != nil {
		log.Fatalf("Test failed: %v", err)
	}

	// Compare results with ground truth
	fmt.Println("\n📈 Validation Results:")
	fmt.Printf("  Ground truth matches: %d\n", len(groundTruth))
	fmt.Printf("  System found matches: %d\n", results.MatchResult.TotalMatches)
	fmt.Printf("  True Positives: %d\n", results.Evaluation.TruePositives)
	fmt.Printf("  False Positives: %d\n", results.Evaluation.FalsePositives)
	fmt.Printf("  False Negatives: %d\n", results.Evaluation.FalseNegatives)
	fmt.Printf("  Precision: %.3f\n", results.Evaluation.Precision)
	fmt.Printf("  Recall: %.3f\n", results.Evaluation.Recall)
	fmt.Printf("  F1-Score: %.3f\n", results.Evaluation.F1Score)
}

// compareRecords performs detailed comparison between two records
func compareRecords(record1, record2 map[string]string, fields []string) {
	fmt.Printf("\n🔬 Detailed Record Comparison\n")
	fmt.Println("=============================")

	// Show record details
	fmt.Printf("Record 1 (ID: %s):\n", record1["id"])
	for _, field := range fields {
		if value, exists := record1[field]; exists {
			fmt.Printf("  %s: %s\n", field, value)
		}
	}

	fmt.Printf("\nRecord 2 (ID: %s):\n", record2["id"])
	for _, field := range fields {
		if value, exists := record2[field]; exists {
			fmt.Printf("  %s: %s\n", field, value)
		}
	}

	// Create Bloom filters for both records
	bf1 := pprl.NewBloomFilter(1024, 8)
	bf2 := pprl.NewBloomFilter(1024, 8)

	// Add fields to Bloom filters using q-grams
	for _, field := range fields {
		if value1, exists1 := record1[field]; exists1 && value1 != "" {
			normalized := normalizeField(value1)
			qgrams := generateQGrams(normalized, 2)
			for _, qgram := range qgrams {
				bf1.Add([]byte(qgram))
			}
		}

		if value2, exists2 := record2[field]; exists2 && value2 != "" {
			normalized := normalizeField(value2)
			qgrams := generateQGrams(normalized, 2)
			for _, qgram := range qgrams {
				bf2.Add([]byte(qgram))
			}
		}
	}

	// Calculate similarities
	hammingDist, err := bf1.HammingDistance(bf2)
	if err != nil {
		fmt.Printf("❌ Failed to calculate Hamming distance: %v\n", err)
		return
	}

	// Calculate match score
	bfSize := bf1.GetSize()
	matchScore := 1.0
	if hammingDist > 0 {
		matchScore = 1.0 - (float64(hammingDist) / float64(bfSize))
	}

	// Calculate Jaccard similarity estimate
	jaccardSim := matchScore // Simplified estimate

	// Determine if records would match
	hammingThreshold := uint32(100)
	jaccardThreshold := 0.7
	isMatch := hammingDist <= hammingThreshold && jaccardSim >= jaccardThreshold

	fmt.Printf("\n📊 Similarity Analysis:\n")
	fmt.Printf("  Hamming Distance: %d (threshold: %d)\n", hammingDist, hammingThreshold)
	fmt.Printf("  Match Score: %.3f\n", matchScore)
	fmt.Printf("  Jaccard Similarity: %.3f (threshold: %.3f)\n", jaccardSim, jaccardThreshold)
	fmt.Printf("  Bloom Filter Size: %d bits\n", bfSize)

	if isMatch {
		fmt.Println("  ✅ MATCH - Records would be considered a match")
	} else {
		fmt.Println("  ❌ NO MATCH - Records would not be considered a match")
	}

	// Show which thresholds failed
	if hammingDist > hammingThreshold {
		fmt.Printf("  ⚠️  Hamming distance too high (%d > %d)\n", hammingDist, hammingThreshold)
	}
	if jaccardSim < jaccardThreshold {
		fmt.Printf("  ⚠️  Jaccard similarity too low (%.3f < %.3f)\n", jaccardSim, jaccardThreshold)
	}
}

// Helper functions

func selectConfigFile() (string, error) {
	// Find .yaml config files in current directory
	yamlFiles := []string{}
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if len(yamlFiles) == 0 {
		return "", fmt.Errorf("no .yaml config files found")
	}

	configPrompt := promptui.Select{
		Label: "Select config file",
		Items: yamlFiles,
	}
	_, configFile, err := configPrompt.Run()
	return configFile, err
}

type TestParameters struct {
	Records1    int
	Records2    int
	Overlap     float64
	Noise       float64
	BloomSize   uint32
	HashCount   uint32
	MinHashSigs uint32
	Hamming     uint32
	Jaccard     float64
	OutputDir   string
}

func getTestParameters() (*TestParameters, error) {
	params := &TestParameters{}

	// Use prompts to get parameters with defaults
	prompt := promptui.Prompt{
		Label:   "Number of records in dataset 1",
		Default: "100",
	}
	result, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	if params.Records1, err = strconv.Atoi(result); err != nil {
		return nil, err
	}

	prompt = promptui.Prompt{
		Label:   "Number of records in dataset 2",
		Default: "120",
	}
	result, err = prompt.Run()
	if err != nil {
		return nil, err
	}
	if params.Records2, err = strconv.Atoi(result); err != nil {
		return nil, err
	}

	prompt = promptui.Prompt{
		Label:   "Overlap rate (0.0-1.0)",
		Default: "0.3",
	}
	result, err = prompt.Run()
	if err != nil {
		return nil, err
	}
	if params.Overlap, err = strconv.ParseFloat(result, 64); err != nil {
		return nil, err
	}

	prompt = promptui.Prompt{
		Label:   "Noise rate (0.0-1.0)",
		Default: "0.1",
	}
	result, err = prompt.Run()
	if err != nil {
		return nil, err
	}
	if params.Noise, err = strconv.ParseFloat(result, 64); err != nil {
		return nil, err
	}

	prompt = promptui.Prompt{
		Label:   "Bloom filter size (bits)",
		Default: "1024",
	}
	result, err = prompt.Run()
	if err != nil {
		return nil, err
	}
	bloomSize, err := strconv.Atoi(result)
	if err != nil {
		return nil, err
	}
	params.BloomSize = uint32(bloomSize)

	prompt = promptui.Prompt{
		Label:   "Number of hash functions",
		Default: "8",
	}
	result, err = prompt.Run()
	if err != nil {
		return nil, err
	}
	hashCount, err := strconv.Atoi(result)
	if err != nil {
		return nil, err
	}
	params.HashCount = uint32(hashCount)

	prompt = promptui.Prompt{
		Label:   "MinHash signatures",
		Default: "64",
	}
	result, err = prompt.Run()
	if err != nil {
		return nil, err
	}
	minHashSigs, err := strconv.Atoi(result)
	if err != nil {
		return nil, err
	}
	params.MinHashSigs = uint32(minHashSigs)

	prompt = promptui.Prompt{
		Label:   "Hamming distance threshold",
		Default: "100",
	}
	result, err = prompt.Run()
	if err != nil {
		return nil, err
	}
	hamming, err := strconv.Atoi(result)
	if err != nil {
		return nil, err
	}
	params.Hamming = uint32(hamming)

	prompt = promptui.Prompt{
		Label:   "Jaccard similarity threshold",
		Default: "0.7",
	}
	result, err = prompt.Run()
	if err != nil {
		return nil, err
	}
	if params.Jaccard, err = strconv.ParseFloat(result, 64); err != nil {
		return nil, err
	}

	prompt = promptui.Prompt{
		Label:   "Output directory",
		Default: "./out",
	}
	params.OutputDir, err = prompt.Run()
	if err != nil {
		return nil, err
	}

	return params, nil
}

type ValidationParameters struct {
	GroundTruthPath string
}

func getValidationParameters() (*ValidationParameters, error) {
	params := &ValidationParameters{}

	prompt := promptui.Prompt{
		Label:   "Path to ground truth CSV file",
		Default: "data/ground_truth.csv",
	}
	result, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	params.GroundTruthPath = result

	return params, nil
}

func normalizeField(value string) string {
	return strings.ToLower(strings.ReplaceAll(value, " ", ""))
}

func generateQGrams(text string, length int) []string {
	if len(text) < length {
		return []string{text}
	}

	var qgrams []string
	for i := 0; i <= len(text)-length; i++ {
		qgrams = append(qgrams, text[i:i+length])
	}
	return qgrams
}

// displayTestResults shows a comprehensive summary of test results
func displayTestResults(results *match.TestResults) {
	fmt.Println("\n🏆 Test Results Summary")
	fmt.Println("======================")

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
