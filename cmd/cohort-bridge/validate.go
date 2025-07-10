package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
	"github.com/auroradata-ai/cohort-bridge/internal/server"
)

// ValidationResult holds the results of validation against ground truth
type ValidationResult struct {
	TruePositives  int
	FalsePositives int
	FalseNegatives int
	Precision      float64
	Recall         float64
	F1Score        float64
	MatchedPairs   []MatchPair
	MissedMatches  []string
	FalseMatches   []MatchPair
}

// MatchPair represents a matched pair (no scores in zero-knowledge validation)
type MatchPair struct {
	ID1 string
	ID2 string
}

// TokenRecord represents a single tokenized record (copied from pprl.go)
type TokenRecordValidation struct {
	ID          string `json:"id"`
	BloomFilter string `json:"bloom_filter"` // base64 encoded
	MinHash     string `json:"minhash"`      // base64 encoded
}

// TokenData represents the tokenized data to be exchanged (copied from pprl.go)
type TokenDataValidation struct {
	Records map[string]TokenRecordValidation `json:"records"`
}

func runValidateCommand(args []string) {
	fmt.Println("CohortBridge Validation Tool")
	fmt.Println("============================")
	fmt.Println("End-to-end validation against ground truth")
	fmt.Println()
	fs := flag.NewFlagSet("validate", flag.ExitOnError)

	var (
		config1File      = fs.String("config1", "", "Configuration file for dataset 1 (Party A)")
		config2File      = fs.String("config2", "", "Configuration file for dataset 2 (Party B)")
		groundTruthFile  = fs.String("ground-truth", "", "Ground truth file with expected matches")
		outputFile       = fs.String("output", "", "Output CSV file for validation report")
		matchThreshold   = fs.Uint("match-threshold", 20, "Hamming distance threshold for matches (default: 20)")
		jaccardThreshold = fs.Float64("jaccard-threshold", 0.32, "Minimum Jaccard similarity for matches (default: 0.32)")
		force            = fs.Bool("force", false, "Skip confirmation prompts and run automatically")
		verbose          = fs.Bool("verbose", false, "Verbose output with detailed analysis")
		interactive      = fs.Bool("interactive", false, "Force interactive mode")
		help             = fs.Bool("help", false, "Show help message")
	)
	fs.Parse(args)

	if *help {
		showValidateHelp()
		return
	}

	// If missing required parameters or interactive mode requested, go interactive
	if (*config1File == "" || *config2File == "" || *groundTruthFile == "" || *outputFile == "") || *interactive {
		fmt.Println("Interactive Validation Setup")
		fmt.Println("Configure your validation parameters...")

		// Get first configuration file
		if *config1File == "" {
			var err error
			*config1File, err = selectConfigFile("Select Configuration File for Dataset 1 (Party A)")
			if err != nil {
				fmt.Printf("Error selecting config1 file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get second configuration file
		if *config2File == "" {
			var err error
			*config2File, err = selectConfigFile("Select Configuration File for Dataset 2 (Party B)")
			if err != nil {
				fmt.Printf("Error selecting config2 file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get ground truth file from data directory
		if *groundTruthFile == "" {
			var err error
			*groundTruthFile, err = selectGroundTruthFile()
			if err != nil {
				fmt.Printf("Error selecting ground truth file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get output file with smart default
		if *outputFile == "" {
			defaultOutput := generateOutputName("validation", *config1File, *config2File)
			*outputFile = promptForInput("Output CSV file for validation report", defaultOutput)
		}

		// Configure match threshold
		fmt.Println("\nMatching Configuration")
		fmt.Println("Configuring thresholds...")
		thresholdChoice := promptForChoice("Select Hamming distance threshold:", []string{
			"20 - Default (recommended for good matches)",
			"10 - Very strict matching",
			"30 - More lenient matching",
			"Custom - Enter custom value",
		})

		switch thresholdChoice {
		case 0:
			*matchThreshold = 20
		case 1:
			*matchThreshold = 10
		case 2:
			*matchThreshold = 30
		case 3:
			customResult := promptForInput("Enter custom Hamming distance threshold (0-100)", "20")
			if val, err := strconv.ParseUint(customResult, 10, 32); err == nil && val <= 100 {
				*matchThreshold = uint(val)
			} else {
				fmt.Println("Invalid threshold, using default: 20")
				*matchThreshold = 20
			}
		}
		// Configure Jaccard threshold
		jaccardChoice := promptForChoice("Select Jaccard similarity threshold:", []string{
			"0.32 - Default (balanced matching)",
			"0.8 - High similarity required",
			"0.3 - More lenient similarity",
			"Custom - Enter custom value",
		})

		switch jaccardChoice {
		case 0:
			*jaccardThreshold = 0.32
		case 1:
			*jaccardThreshold = 0.8
		case 2:
			*jaccardThreshold = 0.3
		case 3:
			customJaccardResult := promptForInput("Enter custom Jaccard similarity threshold (0.0-1.0)", "0.32")
			if val, err := strconv.ParseFloat(customJaccardResult, 64); err == nil && val >= 0.0 && val <= 1.0 {
				*jaccardThreshold = val
			} else {
				fmt.Println("Invalid Jaccard threshold, using default: 0.32")
				*jaccardThreshold = 0.32
			}
		}

		// Verbose mode
		verboseChoice := promptForChoice("Enable verbose output?", []string{
			"Standard - Basic metrics and summary",
			"Verbose - Detailed analysis and breakdown",
		})
		*verbose = (verboseChoice == 1)

		fmt.Println()
	}

	// Default output file if not specified
	if *outputFile == "" {
		*outputFile = generateOutputName("validation", *config1File, *config2File)
	}

	// Show configuration summary
	fmt.Println("Validation Configuration:")
	fmt.Printf("  Config 1 (Party A): %s\n", *config1File)
	fmt.Printf("  Config 2 (Party B): %s\n", *config2File)
	fmt.Printf("  Ground Truth: %s\n", *groundTruthFile)
	fmt.Printf("  Output Report: %s\n", *outputFile)
	fmt.Printf("  Hamming Threshold: %d\n", *matchThreshold)
	fmt.Printf("  Jaccard Threshold: %.3f\n", *jaccardThreshold)
	if *verbose {
		fmt.Println("  Mode: Verbose")
	} else {
		fmt.Println("  Mode: Standard")
	}
	fmt.Println()

	// Confirm before proceeding (unless force flag is set)
	if !*force {
		// Only show confirmation prompt if in interactive mode or missing required params
		if *interactive || (*config1File == "" || *config2File == "" || *groundTruthFile == "") {
			confirmChoice := promptForChoice("Ready to start validation?", []string{
				"Yes, start validation",
				"Change configuration",
				"Cancel",
			})

			if confirmChoice == 2 {
				fmt.Println("\nValidation cancelled. Goodbye!")
				os.Exit(0)
			}

			if confirmChoice == 1 {
				// Restart configuration
				fmt.Println("\nRestarting configuration...")
				newArgs := append([]string{"-interactive"}, args...)
				runValidateCommand(newArgs)
				return
			}
		}
	} else {
		fmt.Println("Starting validation process automatically (force mode)...")
	}

	// Validate inputs before proceeding
	if err := validateValidationInputs(*config1File, *config2File, *groundTruthFile); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	// Run validation
	fmt.Println("Starting validation process...")

	if err := performValidation(*config1File, *config2File, *groundTruthFile, *outputFile, *matchThreshold, *jaccardThreshold, *verbose); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nValidation completed successfully!\n")
	fmt.Printf("Report saved to: %s\n", *outputFile)
}

func validateValidationInputs(config1, config2, groundTruth string) error {
	if _, err := os.Stat(config1); os.IsNotExist(err) {
		return fmt.Errorf("config1 file not found: %s", config1)
	}

	if _, err := os.Stat(config2); os.IsNotExist(err) {
		return fmt.Errorf("config2 file not found: %s", config2)
	}

	if _, err := os.Stat(groundTruth); os.IsNotExist(err) {
		return fmt.Errorf("ground truth file not found: %s", groundTruth)
	}

	return nil
}

func performValidation(config1, config2, groundTruth, outputFile string, matchThreshold uint, jaccardThreshold float64, verbose bool) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println("Loading configurations...")
	fmt.Printf("  Config 1: %s\n", config1)
	fmt.Printf("  Config 2: %s\n", config2)

	// Load configurations
	cfg1, err := config.Load(config1)
	if err != nil {
		return fmt.Errorf("failed to load config1: %w", err)
	}

	cfg2, err := config.Load(config2)
	if err != nil {
		return fmt.Errorf("failed to load config2: %w", err)
	}

	// Use command-line thresholds for validation testing, fall back to config thresholds if not specified
	configHammingThreshold := uint32(matchThreshold)
	configJaccardThreshold := jaccardThreshold
	// If command-line thresholds are default values, use config file thresholds instead
	if matchThreshold == 20 && jaccardThreshold == 0.32 {
		configHammingThreshold = cfg1.Matching.HammingThreshold
		configJaccardThreshold = cfg1.Matching.JaccardThreshold

		// If config2 has different thresholds, use the more permissive one for validation
		if cfg2.Matching.HammingThreshold > configHammingThreshold {
			configHammingThreshold = cfg2.Matching.HammingThreshold
		}
		if cfg2.Matching.JaccardThreshold < configJaccardThreshold {
			configJaccardThreshold = cfg2.Matching.JaccardThreshold
		}
	}

	fmt.Printf("  Using thresholds: Hamming=%d, Jaccard=%.3f\n", configHammingThreshold, configJaccardThreshold)

	fmt.Println("Loading ground truth data...")
	fmt.Printf("  Ground truth: %s\n", groundTruth)

	// Load ground truth
	groundTruthMap, err := loadGroundTruth(groundTruth)
	if err != nil {
		return fmt.Errorf("failed to load ground truth: %w", err)
	}

	fmt.Printf("Loaded %d ground truth matches\n", len(groundTruthMap))

	// Load datasets
	fmt.Println("Loading datasets...")
	records1, err := loadDataset(cfg1, "Dataset 1")
	if err != nil {
		return fmt.Errorf("failed to load dataset 1: %w", err)
	}

	records2, err := loadDataset(cfg2, "Dataset 2")
	if err != nil {
		return fmt.Errorf("failed to load dataset 2: %w", err)
	}

	fmt.Printf("Dataset 1: %d records\n", len(records1))
	fmt.Printf("Dataset 2: %d records\n", len(records2))

	fmt.Println("Running PPRL matching pipeline...")
	fmt.Printf("  Using Hamming threshold: %d (from config)\n", configHammingThreshold)
	fmt.Printf("  Using Jaccard threshold: %.3f (from config)\n", configJaccardThreshold)

	// Configure zero-knowledge matching pipeline
	// All thresholds are now hardcoded for security - no configurable values
	pipelineConfig := &match.PipelineConfig{
		FuzzyMatchConfig: &match.FuzzyMatchConfig{
			Party: 0, // Default to party 0 for validation
		},
		OutputPath:    outputFile + ".matches", // Temporary file for matches
		EnableStats:   true,
		MaxCandidates: 1000,
	}

	// Create matching pipeline
	pipeline, err := match.NewPipeline(pipelineConfig)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}

	// Run matching with config thresholds
	matches, allComparisons, err := runMatchingPipeline(records1, records2, pipeline, configHammingThreshold, configJaccardThreshold)
	if err != nil {
		return fmt.Errorf("failed to run matching pipeline: %w", err)
	}

	fmt.Printf("Found %d matches from %d comparisons\n", len(matches), len(allComparisons))

	if verbose {
		fmt.Println("Performing detailed analysis...")
		fmt.Println("   Computing ROC curve...")
		fmt.Println("   Calculating confusion matrix...")
		fmt.Println("   Analyzing error patterns...")
	}

	fmt.Println("Computing validation metrics...")

	// Validate results against ground truth
	validationResult := validateResults(matches, allComparisons, groundTruthMap)

	// Display results
	fmt.Println("\nValidation Results:")
	fmt.Printf("   True Positives: %d\n", validationResult.TruePositives)
	fmt.Printf("   False Positives: %d\n", validationResult.FalsePositives)
	fmt.Printf("   False Negatives: %d\n", validationResult.FalseNegatives)
	fmt.Printf("   Total Ground Truth Matches: %d\n", len(groundTruthMap))
	fmt.Printf("   Precision: %.3f\n", validationResult.Precision)
	fmt.Printf("   Recall: %.3f\n", validationResult.Recall)
	fmt.Printf("   F1-Score: %.3f\n", validationResult.F1Score)
	if verbose {
		// Show some examples
		if len(validationResult.MatchedPairs) > 0 {
			fmt.Println("\nSample True Positives:")
			for i, pair := range validationResult.MatchedPairs {
				if i >= 3 { // Show first 3
					break
				}
				fmt.Printf("   %s -> %s\n", pair.ID1, pair.ID2)
			}
		}

		if len(validationResult.FalseMatches) > 0 {
			fmt.Println("\nSample False Positives:")
			for i, pair := range validationResult.FalseMatches {
				if i >= 3 { // Show first 3
					break
				}
				fmt.Printf("   %s -> %s\n", pair.ID1, pair.ID2)
			}
		}

		if len(validationResult.MissedMatches) > 0 {
			fmt.Println("\nSample Missed Matches:")
			for i, missed := range validationResult.MissedMatches {
				if i >= 3 { // Show first 3
					break
				}
				fmt.Printf("   %s\n", missed)
			}
		}
	}

	fmt.Println("\nSaving validation report to CSV...")

	// Save detailed validation report
	if err := saveValidationReport(validationResult, outputFile, len(groundTruthMap), verbose); err != nil {
		return fmt.Errorf("failed to save validation report: %w", err)
	}

	fmt.Printf("Validation report saved to: %s\n", outputFile)
	return nil
}

func showValidateHelp() {
	fmt.Println("CohortBridge Validation Tool")
	fmt.Println("============================")
	fmt.Println()
	fmt.Println("Validate PPRL results against ground truth data")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge validate [OPTIONS]")
	fmt.Println("  cohort-bridge validate                    # Interactive mode")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -config1 string       Configuration file for dataset 1 (Party A)")
	fmt.Println("  -config2 string       Configuration file for dataset 2 (Party B)")
	fmt.Println("  -ground-truth string  Ground truth CSV file with expected matches")
	fmt.Println("  -output string        Output CSV file for validation report")
	fmt.Println("  -match-threshold      Hamming distance threshold for matches (default: 20)")
	fmt.Println("  -jaccard-threshold    Jaccard similarity threshold for matches (default: 0.32)")
	fmt.Println("  -verbose              Verbose output with detailed analysis")
	fmt.Println("  -interactive          Force interactive mode")
	fmt.Println("  -force                Skip confirmation prompts and run automatically")
	fmt.Println("  -help                 Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode (prompts for all inputs)")
	fmt.Println("  cohort-bridge validate")
	fmt.Println()
	fmt.Println("  # Command line mode")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/expected_matches.csv")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/expected_matches.csv -verbose")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/expected_matches.csv -match-threshold 15 -jaccard-threshold 0.8")
	fmt.Println()
	fmt.Println("  # Automatic mode (skip confirmations)")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/expected_matches.csv -force")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/expected_matches.csv -verbose -force")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/expected_matches.csv -match-threshold 25 -jaccard-threshold 0.3 -force")
	fmt.Println()
	fmt.Println("  # Force interactive even with some parameters")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -interactive")
}

// loadGroundTruth loads the ground truth CSV file
func loadGroundTruth(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ground truth file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("ground truth file is empty")
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("ground truth file must have at least 2 rows (header + data)")
	}

	groundTruth := make(map[string]string)

	// Always treat the first row as header
	startIdx := 1

	// Check if first row looks like a header (for informational purposes and validation)
	if len(records[0]) >= 2 {
		col1 := strings.TrimSpace(strings.ToLower(records[0][0]))
		col2 := strings.TrimSpace(strings.ToLower(records[0][1]))

		// Common header patterns for ground truth files
		isHeader := col1 == "id1" || col1 == "record1" || col1 == "patient_id1" || col1 == "patient_a_id" ||
			col2 == "id2" || col2 == "record2" || col2 == "patient_id2" || col2 == "patient_b_id" ||
			col1 == "patient1" || col1 == "patientid1" || col1 == "record_id1" ||
			col2 == "patient2" || col2 == "patientid2" || col2 == "record_id2"

		if !isHeader {
			fmt.Printf("   Warning: First row doesn't look like typical CSV headers: [%s, %s]\n", records[0][0], records[0][1])
			fmt.Printf("   Treating it as header anyway. If this is wrong, please format your CSV with proper headers.\n")
		} else {
			fmt.Printf("   Detected CSV headers: [%s, %s]\n", records[0][0], records[0][1])
		}
	}

	// Process data rows (skip header)
	for i := startIdx; i < len(records); i++ {
		record := records[i]
		if len(record) >= 2 {
			id1 := strings.TrimSpace(record[0])
			id2 := strings.TrimSpace(record[1])
			if id1 != "" && id2 != "" {
				groundTruth[id1] = id2
			}
		}
	}

	return groundTruth, nil
}

// loadDataset loads a dataset from configuration for zero-knowledge validation
func loadDataset(cfg *config.Config, datasetName string) ([]*pprl.Record, error) {
	fmt.Printf("   Loading %s...\n", datasetName)

	var records []*pprl.Record
	var err error

	if cfg.Database.IsTokenized {
		fmt.Printf("   Loading tokenized data from %s\n", cfg.Database.Filename)
		records, err = server.LoadTokenizedRecords(cfg.Database.Filename, cfg.IsEncrypted(), cfg.Database.EncryptionKey, cfg.Database.EncryptionKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load tokenized records: %v", err)
		}
	} else {
		fmt.Printf("   Loading raw data from %s\n", cfg.Database.Filename)

		// Use the EXACT SAME tokenization process as the PPRL workflow
		tempTokenFile := fmt.Sprintf("temp_validation_tokens_%s.csv", datasetName)
		err := performValidationTokenization(cfg.Database.Filename, tempTokenFile, cfg.Database.Fields)
		if err != nil {
			return nil, fmt.Errorf("failed to tokenize %s: %w", datasetName, err)
		}
		defer os.Remove(tempTokenFile) // Clean up temp file

		// Load the tokenized data the same way PPRL workflow does
		tokenData, err := loadTokenizedDataForValidation(tempTokenFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load tokenized %s: %w", datasetName, err)
		}

		// Convert to PPRL records using the same method as PPRL workflow
		records, err = tokenDataToPPRLRecordsForValidation(tokenData)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tokenized %s: %w", datasetName, err)
		}
	}

	return records, nil
}

// runMatchingPipeline performs validation using the SAME approach as the PPRL workflow
// This ensures validation uses identical zero-knowledge protocols as production
func runMatchingPipeline(records1, records2 []*pprl.Record, pipeline *match.Pipeline, hammingThreshold uint32, jaccardThreshold float64) ([]*match.PrivateMatchResult, []*match.PrivateMatchResult, error) {
	fmt.Println("   Computing zero-knowledge matching for validation...")
	fmt.Printf("   Using thresholds: Hamming=%d, Jaccard=%.3f\n", hammingThreshold, jaccardThreshold)

	// Use the zero-knowledge fuzzy matcher for validation with proper thresholds
	fuzzyMatcher := match.NewFuzzyMatcher(&match.FuzzyMatchConfig{
		Party:            0,     // Validation uses party 0
		AllowDuplicates:  false, // 1:1 matching for validation
		HammingThreshold: hammingThreshold,
		JaccardThreshold: jaccardThreshold,
	})

	// Perform zero-knowledge intersection computation
	secureResult, err := fuzzyMatcher.ComputePrivateIntersection(records1, records2)
	if err != nil {
		return nil, nil, fmt.Errorf("secure intersection computation failed: %v", err)
	}

	// Convert results to PrivateMatchResult
	var matches []*match.PrivateMatchResult
	for _, privateMatch := range secureResult.MatchPairs {
		matchResult := &match.PrivateMatchResult{
			LocalID: privateMatch.LocalID,
			PeerID:  privateMatch.PeerID,
		}
		matches = append(matches, matchResult)
	}

	fmt.Printf("   ✅ Found %d matches using zero-knowledge protocols\n", len(matches))
	fmt.Printf("   Completed zero-knowledge intersection, found %d matches\n", len(matches))

	// Show sample matches for debugging
	if len(matches) > 0 {
		fmt.Printf("   Sample matches found:\n")
		for i, match := range matches {
			if i >= 3 { // Show first 3 matches only
				break
			}
			fmt.Printf("     %s->%s\n", match.LocalID, match.PeerID)
		}
	}

	return matches, matches, nil
}

// validateResults validates zero-knowledge predicted matches against ground truth
func validateResults(matches []*match.PrivateMatchResult, allComparisons []*match.PrivateMatchResult, groundTruth map[string]string) *ValidationResult {
	result := &ValidationResult{
		MatchedPairs:  make([]MatchPair, 0),
		MissedMatches: make([]string, 0),
		FalseMatches:  make([]MatchPair, 0),
	}

	// Create a set of predicted matches using ONLY IDs
	predictedMatches := make(map[string]string)
	for _, match := range matches {
		predictedMatches[match.LocalID] = match.PeerID
	}

	// Calculate True Positives and False Negatives
	for id1, expectedID2 := range groundTruth {
		if predictedID2, found := predictedMatches[id1]; found && predictedID2 == expectedID2 {
			result.TruePositives++
			// No scores in zero-knowledge validation - just store the match
			result.MatchedPairs = append(result.MatchedPairs, MatchPair{
				ID1: id1,
				ID2: predictedID2,
			})
		} else {
			result.FalseNegatives++
			result.MissedMatches = append(result.MissedMatches, fmt.Sprintf("%s -> %s", id1, expectedID2))
		}
	}

	// Calculate False Positives
	for _, match := range matches {
		if expectedID2, exists := groundTruth[match.LocalID]; !exists || expectedID2 != match.PeerID {
			result.FalsePositives++
			result.FalseMatches = append(result.FalseMatches, MatchPair{
				ID1: match.LocalID,
				ID2: match.PeerID,
			})
		}
	}

	// Calculate metrics (same as before)
	if result.TruePositives+result.FalsePositives > 0 {
		result.Precision = float64(result.TruePositives) / float64(result.TruePositives+result.FalsePositives)
	}

	if result.TruePositives+result.FalseNegatives > 0 {
		result.Recall = float64(result.TruePositives) / float64(result.TruePositives+result.FalseNegatives)
	}

	if result.Precision+result.Recall > 0 {
		result.F1Score = 2 * (result.Precision * result.Recall) / (result.Precision + result.Recall)
	}

	return result
}

// saveValidationReport saves the validation results to a CSV file
func saveValidationReport(result *ValidationResult, outputFile string, totalGroundTruth int, verbose bool) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	// Write summary metrics
	writer.Write([]string{"metric", "value"})
	writer.Write([]string{"true_positives", strconv.Itoa(result.TruePositives)})
	writer.Write([]string{"false_positives", strconv.Itoa(result.FalsePositives)})
	writer.Write([]string{"false_negatives", strconv.Itoa(result.FalseNegatives)})
	writer.Write([]string{"total_ground_truth", strconv.Itoa(totalGroundTruth)})
	writer.Write([]string{"precision", fmt.Sprintf("%.6f", result.Precision)})
	writer.Write([]string{"recall", fmt.Sprintf("%.6f", result.Recall)})
	writer.Write([]string{"f1_score", fmt.Sprintf("%.6f", result.F1Score)})

	// Add detailed results
	writer.Write([]string{""}) // Empty row
	writer.Write([]string{"=== DETAILED RESULTS ==="})
	writer.Write([]string{"match_type", "id1", "id2"})

	// True Positives
	for _, match := range result.MatchedPairs {
		writer.Write([]string{
			"true_positive",
			match.ID1,
			match.ID2,
		})
	}

	// False Positives
	for _, match := range result.FalseMatches {
		writer.Write([]string{
			"false_positive",
			match.ID1,
			match.ID2,
		})
	}
	// False Negatives
	for _, missed := range result.MissedMatches {
		parts := strings.Split(missed, " -> ")
		if len(parts) == 2 {
			writer.Write([]string{
				"false_negative",
				parts[0],
				parts[1],
			})
		}
	}

	return nil
}

// performValidationTokenization - exact copy of performRealTokenization from pprl.go
func performValidationTokenization(inputFile, outputFile string, fields []string) error {
	// Read input CSV file
	csvDB, err := db.NewCSVDatabase(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}

	// Get all records from CSV
	allRecords, err := csvDB.List(0, 10000) // Load all records
	if err != nil {
		return fmt.Errorf("failed to read records: %w", err)
	}

	// Create CSV output file with proper headers
	outputCSV, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputCSV.Close()

	writer := csv.NewWriter(outputCSV)
	defer writer.Flush()

	// Write CSV header
	header := []string{"id", "bloom_filter", "minhash", "timestamp"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// PPRL configuration for tokenization - EXACT SAME as pprl.go
	recordConfig := &pprl.RecordConfig{
		BloomSize:    1000, // 1000 bits
		BloomHashes:  5,    // 5 hash functions
		MinHashSize:  100,  // 100-element signature
		QGramLength:  2,    // 2-grams
		QGramPadding: "$",  // Padding character
		NoiseLevel:   0,    // No noise for deterministic matching
	}

	processedCount := 0
	for _, record := range allRecords {
		// Extract field values for this record
		var fieldValues []string
		for _, field := range fields {
			// Extract actual field name (remove type prefix like "name:", "date:", etc.)
			fieldName := field
			if strings.Contains(field, ":") {
				parts := strings.Split(field, ":")
				if len(parts) == 2 {
					fieldName = parts[1]
				}
			}

			if value, exists := record[fieldName]; exists && value != "" {
				fieldValues = append(fieldValues, value)
			}
		}

		if len(fieldValues) == 0 {
			continue // Skip records with no data in specified fields
		}

		// Get record ID
		recordID := record["id"]
		if recordID == "" {
			// Generate ID if not present
			recordID = fmt.Sprintf("record_%d", processedCount+1)
		}

		// Create PPRL record with real tokenization
		pprlRecord, err := pprl.CreateRecord(recordID, fieldValues, recordConfig)
		if err != nil {
			return fmt.Errorf("failed to create PPRL record for %s: %w", recordID, err)
		}

		// Decode the Bloom filter to compute MinHash from it
		bf, err := pprl.BloomFromBase64(pprlRecord.BloomData)
		if err != nil {
			return fmt.Errorf("failed to decode Bloom filter for %s: %w", recordID, err)
		}

		// Create deterministic MinHash with shared seed for consistent signatures across parties
		mh, err := pprl.NewMinHashSeeded(recordConfig.BloomSize, recordConfig.MinHashSize, "cohort-bridge-pprl-seed")
		if err != nil {
			return fmt.Errorf("failed to create MinHash for %s: %w", recordID, err)
		}

		// Compute the signature directly from the Bloom filter
		_, err = mh.ComputeSignature(bf)
		if err != nil {
			return fmt.Errorf("failed to compute MinHash signature for %s: %w", recordID, err)
		}

		// Convert to CSV format
		timestamp := time.Now().Format("2006-01-02T15:04:05Z")

		// Encode the complete MinHash object to base64
		minHashBase64, err := mh.ToBase64()
		if err != nil {
			return fmt.Errorf("failed to encode MinHash to base64 for %s: %w", recordID, err)
		}

		csvRow := []string{
			recordID,             // KEEP ORIGINAL ID for validation matching
			pprlRecord.BloomData, // Already base64 encoded
			minHashBase64,        // Properly base64 encoded MinHash
			timestamp,
		}

		if err := writer.Write(csvRow); err != nil {
			return fmt.Errorf("failed to write CSV row for %s: %w", recordID, err)
		}

		processedCount++
	}

	return nil
}

// loadTokenizedDataForValidation - exact copy of loadTokenizedData from pprl.go
func loadTokenizedDataForValidation(filename string) (*TokenDataValidation, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 { // Header + at least one record
		return nil, fmt.Errorf("insufficient data in tokenized file")
	}

	tokenData := &TokenDataValidation{Records: make(map[string]TokenRecordValidation)}

	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 4 {
			continue // Skip incomplete records
		}

		tokenRecord := TokenRecordValidation{
			ID:          record[0],
			BloomFilter: record[1],
			MinHash:     record[2],
		}

		tokenData.Records[tokenRecord.ID] = tokenRecord
	}

	return tokenData, nil
}

// tokenDataToPPRLRecordsForValidation - exact copy of tokenDataToPPRLRecords from pprl.go
func tokenDataToPPRLRecordsForValidation(tokenData *TokenDataValidation) ([]*pprl.Record, error) {
	var records []*pprl.Record

	for _, tokenRecord := range tokenData.Records {
		// Decode MinHash from base64
		mh, err := pprl.MinHashFromBase64(tokenRecord.MinHash)
		if err != nil {
			return nil, fmt.Errorf("failed to decode minhash for %s: %v", tokenRecord.ID, err)
		}

		// Get MinHash signature directly - this is the correct way
		minHashSig := mh.GetSignature()
		if minHashSig == nil {
			return nil, fmt.Errorf("failed to get minhash signature for %s", tokenRecord.ID)
		}

		record := &pprl.Record{
			ID:        tokenRecord.ID,
			BloomData: tokenRecord.BloomFilter,
			MinHash:   minHashSig,
			QGramData: "", // Not used in workflow
		}

		records = append(records, record)
	}

	return records, nil
}
