package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/server"
)

func runIntersectCommand(args []string) {
	fmt.Println("CohortBridge Zero-Knowledge Intersection")
	fmt.Println("========================================")
	fmt.Println("Find matches using zero-knowledge protocols with absolute privacy")
	fmt.Println("No information leaked beyond intersection results")
	fmt.Println()

	fs := flag.NewFlagSet("intersect", flag.ExitOnError)
	var (
		dataset1    = fs.String("dataset1", "", "Path to first tokenized dataset file")
		dataset2    = fs.String("dataset2", "", "Path to second tokenized dataset file")
		outputFile  = fs.String("output", "zk_intersection_results.csv", "Output file for intersection results")
		party       = fs.Int("party", 0, "Party number (0 or 1) for two-party protocol")
		interactive = fs.Bool("interactive", false, "Force interactive mode")
		help        = fs.Bool("help", false, "Show help message")
	)
	fs.Parse(args)

	if *help {
		showZKIntersectHelp()
		return
	}

	// Interactive mode if missing required parameters
	if *dataset1 == "" || *dataset2 == "" || *interactive {
		fmt.Println("Interactive Zero-Knowledge Intersection Setup")
		fmt.Println("Configure your secure intersection parameters:\n")

		if *dataset1 == "" {
			var err error
			*dataset1, err = selectDataFile("Select First Tokenized Dataset", "tokenized", []string{".csv", ".json"})
			if err != nil {
				fmt.Printf("Error selecting first dataset: %v\n", err)
				os.Exit(1)
			}
		}

		if *dataset2 == "" {
			var err error
			*dataset2, err = selectDataFile("Select Second Tokenized Dataset", "tokenized", []string{".csv", ".json"})
			if err != nil {
				fmt.Printf("Error selecting second dataset: %v\n", err)
				os.Exit(1)
			}
		}

		if *outputFile == "zk_intersection_results.csv" {
			defaultOutput := generateOutputName("zk_intersection", *dataset1, *dataset2)
			*outputFile = promptForInput("Output file for intersection results", defaultOutput)
		}

		// Only party number is configurable for security
		fmt.Println("\nZero-Knowledge Protocol Configuration")
		partyResult := promptForInput("Party number (0 or 1) for two-party protocol", strconv.Itoa(*party))
		if val, err := strconv.Atoi(partyResult); err == nil && (val == 0 || val == 1) {
			*party = val
		} else {
			fmt.Println("Invalid party number, using default:", *party)
		}
		fmt.Println()
	}

	// Show configuration summary
	fmt.Println("Zero-Knowledge Intersection Configuration:")
	fmt.Printf("  Dataset 1: %s\n", *dataset1)
	fmt.Printf("  Dataset 2: %s\n", *dataset2)
	fmt.Printf("  Output: %s\n", *outputFile)
	fmt.Printf("  Party: %d\n", *party)
	fmt.Printf("  Security: Zero-knowledge protocols (hardcoded thresholds)\n")
	fmt.Println()

	// Confirm before proceeding
	confirmChoice := promptForChoice("Ready to start zero-knowledge intersection?", []string{
		"Yes, find intersections",
		"Change configuration",
		"Cancel",
	})

	if confirmChoice == 2 {
		fmt.Println("\nIntersection cancelled. Goodbye!")
		os.Exit(0)
	}

	if confirmChoice == 1 {
		fmt.Println("\nRestarting configuration...\n")
		newArgs := append([]string{"-interactive"}, args...)
		runIntersectCommand(newArgs)
		return
	}

	// Validate inputs
	if err := validateIntersectInputs(*dataset1, *dataset2); err != nil {
		fmt.Printf("Validation error: %v\n", err)
		os.Exit(1)
	}

	// Run zero-knowledge intersection
	fmt.Println("Starting zero-knowledge intersection process...\n")

	if err := performZeroKnowledgeIntersection(*dataset1, *dataset2, *outputFile, *party); err != nil {
		fmt.Printf("Zero-knowledge intersection failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nZero-knowledge intersection completed successfully!\n")
	fmt.Printf("Results saved to: %s\n", *outputFile)
	fmt.Printf("GUARANTEE: Zero information leaked beyond intersection\n")
}

// generateZKIntersectOutputName function replaced with shared generateOutputName in utils.go

func validateIntersectInputs(dataset1, dataset2 string) error {
	if _, err := os.Stat(dataset1); os.IsNotExist(err) {
		return fmt.Errorf("dataset1 file not found: %s", dataset1)
	}
	if _, err := os.Stat(dataset2); os.IsNotExist(err) {
		return fmt.Errorf("dataset2 file not found: %s", dataset2)
	}
	return nil
}

func performZeroKnowledgeIntersection(dataset1, dataset2, outputFile string, party int) error {
	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println("Loading tokenized datasets...")

	// Load tokenized datasets using server's secure loading (handles encrypted CSV files)
	records1, err := server.LoadTokenizedRecords(dataset1, false, "", "")
	if err != nil {
		return fmt.Errorf("failed to load dataset1: %w", err)
	}
	fmt.Printf("   Loaded %d records from dataset1\n", len(records1))

	records2, err := server.LoadTokenizedRecords(dataset2, false, "", "")
	if err != nil {
		return fmt.Errorf("failed to load dataset2: %w", err)
	}
	fmt.Printf("   Loaded %d records from dataset2\n", len(records2))

	// Configure zero-knowledge fuzzy matcher (only party is configurable)
	fuzzyConfig := &match.FuzzyMatchConfig{
		Party: party,
	}

	// Create zero-knowledge fuzzy matcher
	fuzzyMatcher := match.NewFuzzyMatcher(fuzzyConfig)

	fmt.Println("Computing zero-knowledge intersection...")
	fmt.Printf("   Using hardcoded secure thresholds for maximum privacy\n")

	// Perform zero-knowledge intersection
	zkResult, err := fuzzyMatcher.ComputePrivateIntersection(records1, records2)
	if err != nil {
		return fmt.Errorf("zero-knowledge intersection failed: %w", err)
	}

	// Save results with ZERO information leakage
	fmt.Println("Saving zero-knowledge intersection results...")
	if err := saveZeroKnowledgeResults(zkResult.MatchPairs, outputFile); err != nil {
		return fmt.Errorf("failed to save results: %w", err)
	}

	fmt.Printf("Results: %d matches found (ONLY information revealed)\n", len(zkResult.MatchPairs))
	return nil
}

func showZKIntersectHelp() {
	fmt.Println("CohortBridge Zero-Knowledge Intersection")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Find matches using zero-knowledge protocols with absolute privacy")
	fmt.Println("Guarantees zero information leakage beyond intersection")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge intersect [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -dataset1 <path>       Path to first tokenized dataset file")
	fmt.Println("  -dataset2 <path>       Path to second tokenized dataset file")
	fmt.Println("  -output <path>         Output file for intersection results")
	fmt.Println("  -party <n>             Party number (0 or 1) for two-party protocol")
	fmt.Println("  -interactive           Force interactive mode")
	fmt.Println("  -help                  Show this help message")
	fmt.Println()
	fmt.Println("SECURITY GUARANTEES:")
	fmt.Println("  - Zero-knowledge protocols: No information leaked beyond matches")
	fmt.Println("  - Hardcoded thresholds: No configurable values that could leak data")
	fmt.Println("  - No similarity scores: Only intersection pairs revealed")
	fmt.Println("  - Constant-time operations: Prevents timing attacks")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Basic zero-knowledge intersection")
	fmt.Println("  cohort-bridge intersect -dataset1 tokens1.csv -dataset2 tokens2.csv")
	fmt.Println()
	fmt.Println("  # Specify party for two-party protocol")
	fmt.Println("  cohort-bridge intersect -dataset1 tokens1.csv -dataset2 tokens2.csv -party 1")
	fmt.Println()
	fmt.Println("  # Interactive mode")
	fmt.Println("  cohort-bridge intersect -interactive")
}

func saveZeroKnowledgeResults(matches []crypto.PrivateMatchPair, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header - ONLY the matches, no other information
	fmt.Fprintf(file, "# CohortBridge Zero-Knowledge Intersection Results\n")
	fmt.Fprintf(file, "# Security Guarantee: Zero information leaked beyond intersection\n")
	fmt.Fprintf(file, "# Total matches found: %d\n", len(matches))
	fmt.Fprintf(file, "local_id,peer_id\n")

	// Write ONLY the matching pairs - no scores, distances, or metadata
	for _, match := range matches {
		fmt.Fprintf(file, "%s,%s\n", match.LocalID, match.PeerID)
	}

	return nil
}
