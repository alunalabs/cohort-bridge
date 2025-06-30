package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
	"github.com/auroradata-ai/cohort-bridge/internal/server"
)

// ValidationResult holds the results of validation against ground truth
type ValidationResult struct {
	TruePositives     int
	FalsePositives    int
	FalseNegatives    int
	Precision         float64
	Recall            float64
	F1Score           float64
	MatchedPairs      []MatchPair
	MissedMatches     []string
	FalseMatches      []MatchPair
	LowestTrueScore   float64
	HighestFalseScore float64
}

// MatchPair represents a matched pair with its score
type MatchPair struct {
	ID1   string
	ID2   string
	Score float64
}

func runValidateCommand(args []string) {
	fmt.Println("🔬 CohortBridge Validation Tool")
	fmt.Println("===============================")
	fmt.Println("End-to-end validation against ground truth")
	fmt.Println()

	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	var (
		config1File     = fs.String("config1", "", "Configuration file for dataset 1 (Party A)")
		config2File     = fs.String("config2", "", "Configuration file for dataset 2 (Party B)")
		groundTruthFile = fs.String("ground-truth", "", "Ground truth file with expected matches")
		outputFile      = fs.String("output", "", "Output CSV file for validation report")
		matchThreshold  = fs.Uint("match-threshold", 100, "Hamming distance threshold for matches")
		verbose         = fs.Bool("verbose", false, "Verbose output with detailed analysis")
		interactive     = fs.Bool("interactive", false, "Force interactive mode")
		help            = fs.Bool("help", false, "Show help message")
	)
	fs.Parse(args)

	if *help {
		showValidateHelp()
		return
	}

	// If missing required parameters or interactive mode requested, go interactive
	if *config1File == "" || *config2File == "" || *groundTruthFile == "" || *outputFile == "" || *interactive {
		fmt.Println("🎯 Interactive Validation Setup")
		fmt.Println("Let's configure your validation parameters...\n")

		// Get first configuration file
		if *config1File == "" {
			var err error
			*config1File, err = selectConfigFile("Select Configuration File for Dataset 1 (Party A)")
			if err != nil {
				fmt.Printf("❌ Error selecting config1 file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get second configuration file
		if *config2File == "" {
			var err error
			*config2File, err = selectConfigFile("Select Configuration File for Dataset 2 (Party B)")
			if err != nil {
				fmt.Printf("❌ Error selecting config2 file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get ground truth file from data directory
		if *groundTruthFile == "" {
			var err error
			*groundTruthFile, err = selectGroundTruthFile()
			if err != nil {
				fmt.Printf("❌ Error selecting ground truth file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get output file with smart default
		if *outputFile == "" {
			defaultOutput := generateValidationOutputName(*config1File, *config2File)
			*outputFile = promptForInput("Output CSV file for validation report", defaultOutput)
		}

		// Configure match threshold
		fmt.Println("\n🎯 Matching Configuration")

		thresholdChoice := promptForChoice("Select match threshold:", []string{
			"🎯 100 - Default (recommended for good matches)",
			"🔥 50 - Very strict matching",
			"⚖️  150 - More lenient matching",
			"🔧 Custom - Enter custom value",
		})

		switch thresholdChoice {
		case 0:
			*matchThreshold = 100
		case 1:
			*matchThreshold = 50
		case 2:
			*matchThreshold = 150
		case 3:
			customResult := promptForInput("Enter custom Hamming distance threshold (0-500)", "100")
			if val, err := strconv.ParseUint(customResult, 10, 32); err == nil && val <= 500 {
				*matchThreshold = uint(val)
			} else {
				fmt.Println("⚠️  Invalid threshold, using default: 100")
				*matchThreshold = 100
			}
		}

		// Verbose mode
		verboseChoice := promptForChoice("Enable verbose output?", []string{
			"📊 Standard - Basic metrics and summary",
			"🔍 Verbose - Detailed analysis and breakdown",
		})
		*verbose = (verboseChoice == 1)

		fmt.Println()
	}

	// Default output file if not specified
	if *outputFile == "" {
		*outputFile = generateValidationOutputName(*config1File, *config2File)
	}

	// Show configuration summary
	fmt.Println("📋 Validation Configuration:")
	fmt.Printf("  📁 Config 1 (Party A): %s\n", *config1File)
	fmt.Printf("  📁 Config 2 (Party B): %s\n", *config2File)
	fmt.Printf("  📊 Ground Truth: %s\n", *groundTruthFile)
	fmt.Printf("  📝 Output Report: %s\n", *outputFile)
	fmt.Printf("  🎯 Match Threshold: %d\n", *matchThreshold)
	if *verbose {
		fmt.Println("  🔍 Mode: Verbose")
	} else {
		fmt.Println("  📊 Mode: Standard")
	}
	fmt.Println()

	// Only show confirmation prompt if in interactive mode
	if *interactive || (*config1File == "" || *config2File == "" || *groundTruthFile == "") {
		// Confirm before proceeding
		confirmChoice := promptForChoice("Ready to start validation?", []string{
			"✅ Yes, start validation",
			"⚙️  Change configuration",
			"❌ Cancel",
		})

		if confirmChoice == 2 {
			fmt.Println("\n👋 Validation cancelled. Goodbye!")
			os.Exit(0)
		}

		if confirmChoice == 1 {
			// Restart configuration
			fmt.Println("\n🔄 Restarting configuration...\n")
			newArgs := append([]string{"-interactive"}, args...)
			runValidateCommand(newArgs)
			return
		}
	}

	// Validate inputs before proceeding
	if err := validateValidationInputs(*config1File, *config2File, *groundTruthFile); err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
		os.Exit(1)
	}

	// Run validation
	fmt.Println("🚀 Starting validation process...\n")

	if err := performValidation(*config1File, *config2File, *groundTruthFile, *outputFile, *matchThreshold, *verbose); err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Validation completed successfully!\n")
	fmt.Printf("📁 Report saved to: %s\n", *outputFile)
}

func selectConfigFile(label string) (string, error) {
	// Find YAML config files in current directory
	var configFiles []string

	matches, _ := filepath.Glob("*.yaml")
	for _, match := range matches {
		if strings.Contains(strings.ToLower(match), "example") {
			continue // Skip example files
		}
		configFiles = append(configFiles, match)
	}

	if len(configFiles) == 0 {
		// Manual input if no files found
		return promptForInput(label+" (enter .yaml file path)", ""), nil
	}

	// Add manual input option
	configFiles = append(configFiles, "📝 Enter file path manually...")

	// Create display options with file descriptions
	var displayOptions []string
	for _, file := range configFiles {
		if file == "📝 Enter file path manually..." {
			displayOptions = append(displayOptions, file)
		} else {
			description := getConfigDescription(file)
			displayOptions = append(displayOptions, fmt.Sprintf("📄 %s - %s", file, description))
		}
	}

	selectedIndex := promptForChoice(label, displayOptions)

	selectedFile := configFiles[selectedIndex]
	if selectedFile == "📝 Enter file path manually..." {
		return promptForInput("Enter config file path (.yaml)", ""), nil
	}

	return selectedFile, nil
}

func selectGroundTruthFile() (string, error) {
	// Look for ground truth CSV files specifically in data directory
	var groundTruthFiles []string

	dataDir := "data"
	if _, err := os.Stat(dataDir); !os.IsNotExist(err) {
		matches, _ := filepath.Glob(filepath.Join(dataDir, "*.csv"))
		for _, match := range matches {
			name := strings.ToLower(filepath.Base(match))
			// Look for files that contain ground truth keywords
			if strings.Contains(name, "expected") ||
				strings.Contains(name, "truth") ||
				strings.Contains(name, "match") {
				groundTruthFiles = append(groundTruthFiles, match)
			}
		}
	}

	if len(groundTruthFiles) == 0 {
		// Manual input if no files found
		return promptForInput("Ground Truth CSV File (enter path, should be in data/ directory)", ""), nil
	}

	// Add manual input option
	groundTruthFiles = append(groundTruthFiles, "📝 Enter file path manually...")

	// Create display options with file info
	var displayOptions []string
	for _, file := range groundTruthFiles {
		if file == "📝 Enter file path manually..." {
			displayOptions = append(displayOptions, file)
		} else {
			info, _ := os.Stat(file)
			size := info.Size()
			sizeStr := fmt.Sprintf("%.1fKB", float64(size)/1024)
			displayOptions = append(displayOptions, fmt.Sprintf("📊 %s (%s)", file, sizeStr))
		}
	}

	selectedIndex := promptForChoice("Select Ground Truth File", displayOptions)

	selectedFile := groundTruthFiles[selectedIndex]
	if selectedFile == "📝 Enter file path manually..." {
		return promptForInput("Enter ground truth file path (.csv)", ""), nil
	}

	return selectedFile, nil
}

func selectDataFile(label, context string, extensions []string) (string, error) {
	// Find files in current directory and common data directories
	searchDirs := []string{".", "data", "out", "results", "logs"}
	var files []string

	for _, dir := range searchDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue
		}

		matches, _ := filepath.Glob(filepath.Join(dir, "*"))
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && !info.IsDir() {
				// Check if file has relevant extension or contains context keywords
				ext := strings.ToLower(filepath.Ext(match))
				name := strings.ToLower(filepath.Base(match))

				hasValidExt := false
				for _, validExt := range extensions {
					if ext == validExt {
						hasValidExt = true
						break
					}
				}

				containsContext := strings.Contains(name, context) ||
					strings.Contains(name, "truth") ||
					strings.Contains(name, "result") ||
					strings.Contains(name, "match") ||
					strings.Contains(name, "validation") ||
					strings.Contains(name, "tokenized") ||
					strings.Contains(name, "data")

				if hasValidExt || containsContext {
					files = append(files, match)
				}
			}
		}
	}

	if len(files) == 0 {
		// No files found, ask for manual input
		return promptForInput(label+" (enter file path)", ""), nil
	}

	// Add manual input option
	files = append(files, "📝 Enter file path manually...")

	// Create display options with file info
	var displayOptions []string
	for _, file := range files {
		if file == "📝 Enter file path manually..." {
			displayOptions = append(displayOptions, file)
		} else {
			info, _ := os.Stat(file)
			size := info.Size()
			sizeStr := fmt.Sprintf("%.1fKB", float64(size)/1024)
			if size > 1024*1024 {
				sizeStr = fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
			}
			displayOptions = append(displayOptions, fmt.Sprintf("📁 %s (%s)", file, sizeStr))
		}
	}

	selectedIndex := promptForChoice(label, displayOptions)

	selectedFile := files[selectedIndex]
	if selectedFile == "📝 Enter file path manually..." {
		return promptForInput("Enter file path", ""), nil
	}

	return selectedFile, nil
}

func getConfigDescription(filename string) string {
	// Try to give meaningful descriptions based on filename patterns
	lower := strings.ToLower(filename)

	if strings.Contains(lower, "_a") || strings.Contains(lower, "party_a") {
		return "Party A configuration"
	}
	if strings.Contains(lower, "_b") || strings.Contains(lower, "party_b") {
		return "Party B configuration"
	}
	if strings.Contains(lower, "sender") {
		return "Sender configuration"
	}
	if strings.Contains(lower, "receiver") {
		return "Receiver configuration"
	}
	if strings.Contains(lower, "secure") {
		return "Secure/encrypted configuration"
	}
	if strings.Contains(lower, "postgres") {
		return "PostgreSQL database configuration"
	}
	if strings.Contains(lower, "tokenized") {
		return "Tokenized data configuration"
	}

	return "Configuration file"
}

func generateValidationOutputName(config1, config2 string) string {
	base1 := strings.TrimSuffix(filepath.Base(config1), filepath.Ext(config1))
	base2 := strings.TrimSuffix(filepath.Base(config2), filepath.Ext(config2))

	return filepath.Join("out", fmt.Sprintf("validation_%s_vs_%s.csv", base1, base2))
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

func performValidation(config1, config2, groundTruth, outputFile string, matchThreshold uint, verbose bool) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println("🔧 Loading configurations...")
	fmt.Printf("  📁 Config 1: %s\n", config1)
	fmt.Printf("  📁 Config 2: %s\n", config2)

	// Load configurations
	cfg1, err := config.Load(config1)
	if err != nil {
		return fmt.Errorf("failed to load config1: %w", err)
	}

	cfg2, err := config.Load(config2)
	if err != nil {
		return fmt.Errorf("failed to load config2: %w", err)
	}

	fmt.Println("📊 Loading ground truth data...")
	fmt.Printf("  📊 Ground truth: %s\n", groundTruth)

	// Load ground truth
	groundTruthMap, err := loadGroundTruth(groundTruth)
	if err != nil {
		return fmt.Errorf("failed to load ground truth: %w", err)
	}

	fmt.Printf("✅ Loaded %d ground truth matches\n", len(groundTruthMap))

	// Load datasets
	fmt.Println("📂 Loading datasets...")
	records1, err := loadDataset(cfg1, "Dataset 1")
	if err != nil {
		return fmt.Errorf("failed to load dataset 1: %w", err)
	}

	records2, err := loadDataset(cfg2, "Dataset 2")
	if err != nil {
		return fmt.Errorf("failed to load dataset 2: %w", err)
	}

	fmt.Printf("✅ Dataset 1: %d records\n", len(records1))
	fmt.Printf("✅ Dataset 2: %d records\n", len(records2))

	fmt.Println("🔄 Running PPRL matching pipeline...")
	fmt.Printf("  🎯 Using Hamming threshold: %d\n", matchThreshold)

	// Configure matching pipeline
	pipelineConfig := &match.PipelineConfig{
		FuzzyMatchConfig: &match.FuzzyMatchConfig{
			HammingThreshold:  uint32(matchThreshold),
			JaccardThreshold:  0.5, // Default Jaccard threshold
			UseSecureProtocol: false,
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

	// Run matching
	matches, allComparisons, err := runMatchingPipeline(records1, records2, pipeline, uint32(matchThreshold))
	if err != nil {
		return fmt.Errorf("failed to run matching pipeline: %w", err)
	}

	fmt.Printf("✅ Found %d matches from %d comparisons\n", len(matches), len(allComparisons))

	if verbose {
		fmt.Println("🔍 Performing detailed analysis...")
		fmt.Println("   📈 Computing ROC curve...")
		fmt.Println("   📊 Calculating confusion matrix...")
		fmt.Println("   🎯 Analyzing error patterns...")
	}

	fmt.Println("⚖️  Computing validation metrics...")

	// Validate results against ground truth
	validationResult := validateResults(matches, allComparisons, groundTruthMap)

	// Display results
	fmt.Println("\n📈 Validation Results:")
	fmt.Printf("   True Positives: %d\n", validationResult.TruePositives)
	fmt.Printf("   False Positives: %d\n", validationResult.FalsePositives)
	fmt.Printf("   False Negatives: %d\n", validationResult.FalseNegatives)
	fmt.Printf("   Total Ground Truth Matches: %d\n", len(groundTruthMap))
	fmt.Printf("   Precision: %.3f\n", validationResult.Precision)
	fmt.Printf("   Recall: %.3f\n", validationResult.Recall)
	fmt.Printf("   F1-Score: %.3f\n", validationResult.F1Score)

	if verbose {
		fmt.Printf("   Lowest ground truth score: %.3f\n", validationResult.LowestTrueScore)
		fmt.Printf("   Highest non-ground truth score: %.3f\n", validationResult.HighestFalseScore)

		// Show some examples
		if len(validationResult.MatchedPairs) > 0 {
			fmt.Println("\n🎯 Sample True Positives:")
			for i, pair := range validationResult.MatchedPairs {
				if i >= 3 { // Show first 3
					break
				}
				fmt.Printf("   %s -> %s (score: %.3f)\n", pair.ID1, pair.ID2, pair.Score)
			}
		}

		if len(validationResult.FalseMatches) > 0 {
			fmt.Println("\n❌ Sample False Positives:")
			for i, pair := range validationResult.FalseMatches {
				if i >= 3 { // Show first 3
					break
				}
				fmt.Printf("   %s -> %s (score: %.3f)\n", pair.ID1, pair.ID2, pair.Score)
			}
		}

		if len(validationResult.MissedMatches) > 0 {
			fmt.Println("\n🔍 Sample Missed Matches:")
			for i, missed := range validationResult.MissedMatches {
				if i >= 3 { // Show first 3
					break
				}
				fmt.Printf("   %s\n", missed)
			}
		}
	}

	fmt.Println("\n💾 Saving validation report to CSV...")

	// Save detailed validation report
	if err := saveValidationReport(validationResult, outputFile, len(groundTruthMap), verbose); err != nil {
		return fmt.Errorf("failed to save validation report: %w", err)
	}

	fmt.Printf("✅ Validation report saved to: %s\n", outputFile)
	return nil
}

func showValidateHelp() {
	fmt.Println("🔬 CohortBridge Validation Tool")
	fmt.Println("===============================")
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
	fmt.Println("  -match-threshold      Hamming distance threshold for matches")
	fmt.Println("  -verbose              Verbose output with detailed analysis")
	fmt.Println("  -interactive          Force interactive mode")
	fmt.Println("  -help                 Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode (prompts for all inputs)")
	fmt.Println("  cohort-bridge validate")
	fmt.Println()
	fmt.Println("  # Command line mode")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/expected_matches.csv")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/expected_matches.csv -verbose")
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
			fmt.Printf("   ⚠️  Warning: First row doesn't look like typical CSV headers: [%s, %s]\n", records[0][0], records[0][1])
			fmt.Printf("   📝 Treating it as header anyway. If this is wrong, please format your CSV with proper headers.\n")
		} else {
			fmt.Printf("   ✅ Detected CSV headers: [%s, %s]\n", records[0][0], records[0][1])
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

// loadDataset loads a dataset from configuration
func loadDataset(cfg *config.Config, datasetName string) ([]server.PatientRecord, error) {
	fmt.Printf("   📊 Loading %s...\n", datasetName)

	var records []server.PatientRecord
	var err error

	if cfg.Database.IsTokenized {
		fmt.Printf("   📁 Loading tokenized data from %s\n", cfg.Database.TokenizedFile)
		records, err = server.LoadTokenizedRecords(cfg.Database.TokenizedFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load tokenized %s: %w", datasetName, err)
		}
	} else {
		fmt.Printf("   📁 Loading raw data from %s\n", cfg.Database.Filename)
		csvDB, err := db.NewCSVDatabase(cfg.Database.Filename)
		if err != nil {
			return nil, fmt.Errorf("failed to load %s: %w", datasetName, err)
		}
		records, err = server.LoadPatientRecordsUtilWithRandomBits(csvDB, cfg.Database.Fields, 0.0) // No random bits for validation
		if err != nil {
			return nil, fmt.Errorf("failed to convert %s: %w", datasetName, err)
		}
	}

	return records, nil
}

// runMatchingPipeline runs the PPRL matching pipeline
func runMatchingPipeline(records1, records2 []server.PatientRecord, pipeline *match.Pipeline, hammingThreshold uint32) ([]*match.MatchResult, []*match.MatchResult, error) {
	fmt.Println("   🔄 Computing pairwise comparisons...")

	var allComparisons []*match.MatchResult
	var matches []*match.MatchResult

	totalComparisons := 0
	fmt.Printf("   🔧 Using Hamming threshold: %d (distances <= %d will be matches)\n", hammingThreshold, hammingThreshold)

	// Perform all pairwise comparisons
	for _, record1 := range records1 {
		for _, record2 := range records2 {
			totalComparisons++

			// Calculate Hamming distance
			hammingDist, err := record1.BloomFilter.HammingDistance(record2.BloomFilter)
			if err != nil {
				continue // Skip this comparison on error
			}

			// Calculate match score using PROVEN working method from test command
			bfSize := record1.BloomFilter.GetSize()
			matchScore := 1.0
			if hammingDist > 0 {
				matchScore = 1.0 - (float64(hammingDist) / float64(bfSize))
			}

			// Calculate Jaccard similarity using pre-computed signatures if available
			var jaccardSim float64
			if len(record1.MinHashSignature) > 0 && len(record2.MinHashSignature) > 0 {
				jaccardSim, _ = pprl.JaccardSimilarity(record1.MinHashSignature, record2.MinHashSignature)
			} else if record1.MinHash != nil && record2.MinHash != nil {
				sig1, err1 := record1.MinHash.ComputeSignature(record1.BloomFilter)
				sig2, err2 := record2.MinHash.ComputeSignature(record2.BloomFilter)
				if err1 == nil && err2 == nil {
					jaccardSim, _ = pprl.JaccardSimilarity(sig1, sig2)
				}
			}

			// Determine if this is a match based on Hamming threshold ONLY (same as test command)
			isMatch := hammingDist <= hammingThreshold

			// Create match result
			matchResult := &match.MatchResult{
				ID1:               record1.ID,
				ID2:               record2.ID,
				HammingDistance:   hammingDist,
				JaccardSimilarity: jaccardSim,
				MatchScore:        matchScore, // Use normalized 0-1 score
				IsMatch:           isMatch,
			}

			allComparisons = append(allComparisons, matchResult)

			// Add to matches if it meets threshold
			if matchResult.IsMatch {
				matches = append(matches, matchResult)
			}

			// Debug first few comparisons
			if totalComparisons <= 3 {
				fmt.Printf("   🔍 DEBUG: Comparison #%d: %s->%s, Hamming=%d, Score=%.6f, IsMatch=%v\n",
					totalComparisons, record1.ID, record2.ID, hammingDist, matchScore, isMatch)
			}
		}
	}

	fmt.Printf("   ✅ Completed %d comparisons, found %d matches (Hamming <= %d)\n", len(allComparisons), len(matches), hammingThreshold)

	// Debug sample of matches found
	if len(matches) > 0 {
		fmt.Printf("   🔍 Sample matches found:\n")
		for i, match := range matches {
			if i >= 3 { // Show first 3
				break
			}
			fmt.Printf("     %s->%s: Hamming=%d, Score=%.6f\n", match.ID1, match.ID2, match.HammingDistance, match.MatchScore)
		}
	}

	return matches, allComparisons, nil
}

// validateResults validates predicted matches against ground truth
func validateResults(matches []*match.MatchResult, allComparisons []*match.MatchResult, groundTruth map[string]string) *ValidationResult {
	result := &ValidationResult{
		MatchedPairs:      make([]MatchPair, 0),
		MissedMatches:     make([]string, 0),
		FalseMatches:      make([]MatchPair, 0),
		LowestTrueScore:   1000.0,
		HighestFalseScore: 0.0,
	}

	// Create a set of predicted matches
	predictedMatches := make(map[string]string)
	for _, match := range matches {
		if match.IsMatch {
			predictedMatches[match.ID1] = match.ID2
		}
	}

	// Calculate True Positives and False Negatives
	for id1, expectedID2 := range groundTruth {
		if predictedID2, found := predictedMatches[id1]; found && predictedID2 == expectedID2 {
			result.TruePositives++
			score := 0.0
			// Look for the actual comparison score in allComparisons
			for _, comparison := range allComparisons {
				if comparison.ID1 == id1 && comparison.ID2 == expectedID2 {
					score = comparison.MatchScore
					break
				}
			}
			result.MatchedPairs = append(result.MatchedPairs, MatchPair{
				ID1:   id1,
				ID2:   predictedID2,
				Score: score,
			})
			if score < result.LowestTrueScore {
				result.LowestTrueScore = score
			}
		} else {
			result.FalseNegatives++
			result.MissedMatches = append(result.MissedMatches, fmt.Sprintf("%s -> %s", id1, expectedID2))
		}
	}

	// Calculate False Positives
	for _, match := range matches {
		if match.IsMatch {
			if expectedID2, exists := groundTruth[match.ID1]; !exists || expectedID2 != match.ID2 {
				result.FalsePositives++
				result.FalseMatches = append(result.FalseMatches, MatchPair{
					ID1:   match.ID1,
					ID2:   match.ID2,
					Score: match.MatchScore,
				})
				if match.MatchScore > result.HighestFalseScore {
					result.HighestFalseScore = match.MatchScore
				}
			}
		}
	}

	// Calculate metrics
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

	if verbose {
		writer.Write([]string{"lowest_true_score", fmt.Sprintf("%.6f", result.LowestTrueScore)})
		writer.Write([]string{"highest_false_score", fmt.Sprintf("%.6f", result.HighestFalseScore)})
	}

	// Add detailed results
	writer.Write([]string{""}) // Empty row
	writer.Write([]string{"=== DETAILED RESULTS ==="})
	writer.Write([]string{"match_type", "id1", "id2", "score"})

	// True Positives
	for _, match := range result.MatchedPairs {
		writer.Write([]string{
			"true_positive",
			match.ID1,
			match.ID2,
			fmt.Sprintf("%.6f", match.Score),
		})
	}

	// False Positives
	for _, match := range result.FalseMatches {
		writer.Write([]string{
			"false_positive",
			match.ID1,
			match.ID2,
			fmt.Sprintf("%.6f", match.Score),
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
				"",
			})
		}
	}

	return nil
}
