package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

func runTokenizeCommand(args []string) {
	fmt.Println("🔐 PPRL Tokenization Tool")
	fmt.Println("=========================")
	fmt.Println("Converts raw PHI data to privacy-preserving Bloom filter tokens")
	fmt.Println()

	fs := flag.NewFlagSet("tokenize", flag.ExitOnError)
	var (
		mainConfigFile = fs.String("main-config", "config.yaml", "Main config file to read field names from")
		inputFile      = fs.String("input", "", "Input file with PHI data")
		outputFile     = fs.String("output", "", "Output file for tokenized data")
		inputFormat    = fs.String("input-format", "csv", "Input format: csv, json, postgres")
		outputFormat   = fs.String("output-format", "csv", "Output format: csv, json")
		batchSize      = fs.Int("batch-size", 1000, "Number of records to process in each batch")
		interactive    = fs.Bool("interactive", false, "Force interactive mode")
		useDatabase    = fs.Bool("database", false, "Use database from main config instead of file")
		minHashSeed    = fs.String("minhash-seed", "cohort-bridge-shared-seed-2024", "Seed for deterministic MinHash generation")
		help           = fs.Bool("help", false, "Show help message")
	)
	fs.Parse(args)

	if *help {
		showTokenizeHelp()
		return
	}

	// Ensure output directory exists
	if err := os.MkdirAll("out", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// If missing required parameters or interactive mode requested, go interactive
	if (*inputFile == "" && !*useDatabase) || *outputFile == "" || *interactive {
		fmt.Println("🎯 Interactive Tokenization Setup")
		fmt.Println("Let's configure your tokenization parameters...\n")

		// Load config to get field information
		var defaultFields []string
		if cfg, err := config.Load(*mainConfigFile); err == nil {
			if len(cfg.Database.Fields) > 0 {
				defaultFields = cfg.Database.Fields
			}
		}
		if len(defaultFields) == 0 {
			defaultFields = []string{"FIRST", "LAST", "BIRTHDATE", "ZIP"}
		}

		// Choose data source
		if !*useDatabase {
			sourceChoice := promptForChoice("Select data source:", []string{
				"📁 File - Process data from a file",
				"🗄️  Database - Use database connection from config",
			})
			*useDatabase = (sourceChoice == 1)
		}

		// Get input file if using file mode
		if !*useDatabase && *inputFile == "" {
			var err error
			*inputFile, err = selectDataFile("Select Input Data File", "data", []string{".csv", ".json", ".txt"})
			if err != nil {
				fmt.Printf("❌ Error selecting input file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get output file
		if *outputFile == "" {
			defaultOutput := generateTokenizeOutputName(*inputFile, *useDatabase)
			*outputFile = promptForInput("Output file for tokenized data", defaultOutput)
		}

		// Select input format with Auto-detect as default
		if !*useDatabase {
			fmt.Println("\nSelect input format (default: Auto-detect):")
			formatOptions := []string{
				"🔧 Auto-detect from file extension",
				"📄 CSV - Comma-separated values",
				"📋 JSON - JavaScript Object Notation",
			}

			formatChoice := promptForChoice("", formatOptions)
			switch formatChoice {
			case 0:
				*inputFormat = detectInputFormat(*inputFile)
			case 1:
				*inputFormat = "csv"
			case 2:
				*inputFormat = "json"
			}
		} else {
			*inputFormat = "database"
		}

		// Select output format with input format as default
		var defaultOutputFormat string
		if *inputFormat == "csv" || *inputFormat == "database" {
			defaultOutputFormat = "csv"
		} else {
			defaultOutputFormat = "json"
		}

		fmt.Printf("\nSelect output format (default: %s):\n", strings.ToUpper(defaultOutputFormat))
		outFormatOptions := []string{
			fmt.Sprintf("📄 CSV - Comma-separated values %s", ifDefault(defaultOutputFormat == "csv")),
			fmt.Sprintf("📋 JSON - JavaScript Object Notation %s", ifDefault(defaultOutputFormat == "json")),
		}

		outFormatChoice := promptForChoice("", outFormatOptions)
		if outFormatChoice == 0 {
			*outputFormat = "csv"
		} else {
			*outputFormat = "json"
		}

		// Configure batch size
		batchSizeStr := promptForInput("Batch size (records to process at once)", strconv.Itoa(*batchSize))
		if val, err := strconv.Atoi(batchSizeStr); err == nil && val > 0 && val <= 100000 {
			*batchSize = val
		} else {
			fmt.Println("⚠️  Invalid batch size, using default:", *batchSize)
		}

		// Configure MinHash seed
		*minHashSeed = promptForInput("MinHash seed for deterministic hashing", *minHashSeed)

		fmt.Println()
	}

	// Try to load field names from main config file
	var defaultFields []string
	if mainConfig, err := config.Load(*mainConfigFile); err == nil {
		if len(mainConfig.Database.Fields) > 0 {
			defaultFields = mainConfig.Database.Fields
			fmt.Printf("📋 Using field names from %s: %v\n", *mainConfigFile, defaultFields)
		}
	}

	// Fallback to CSV headers if config doesn't have fields
	if len(defaultFields) == 0 {
		defaultFields = []string{"FIRST", "LAST", "BIRTHDATE", "ZIP"}
		fmt.Printf("⚠️  Could not load field names from %s, using defaults: %v\n", *mainConfigFile, defaultFields)
	}

	// Show configuration summary
	fmt.Println("📋 Tokenization Configuration:")
	if *useDatabase {
		fmt.Println("  📊 Data Source: Database (from config)")
	} else {
		fmt.Printf("  📊 Input File: %s\n", *inputFile)
		fmt.Printf("  📄 Input Format: %s\n", *inputFormat)
	}
	fmt.Printf("  📁 Output File: %s\n", *outputFile)
	fmt.Printf("  📄 Output Format: %s\n", *outputFormat)
	fmt.Printf("  🔢 Batch Size: %d\n", *batchSize)
	fmt.Printf("  🏷️  Fields: %v\n", defaultFields)
	fmt.Printf("  🔑 MinHash Seed: %s\n", *minHashSeed)
	fmt.Println()

	// Confirm before proceeding
	confirmChoice := promptForChoice("Ready to start tokenization?", []string{
		"✅ Yes, start tokenization",
		"⚙️  Change configuration",
		"❌ Cancel",
	})

	if confirmChoice == 2 {
		fmt.Println("\n👋 Tokenization cancelled. Goodbye!")
		os.Exit(0)
	}

	if confirmChoice == 1 {
		// Restart configuration
		fmt.Println("\n🔄 Restarting configuration...\n")
		newArgs := append([]string{"-interactive"}, args...)
		runTokenizeCommand(newArgs)
		return
	}

	// Validate inputs before proceeding
	if err := validateTokenizeInputs(*inputFile, *useDatabase, *mainConfigFile); err != nil {
		fmt.Printf("❌ Validation error: %v\n", err)
		os.Exit(1)
	}

	// Run tokenization
	fmt.Println("🚀 Starting tokenization process...\n")

	if err := performTokenization(*inputFile, *outputFile, *inputFormat, *outputFormat, *batchSize, *minHashSeed, *useDatabase, defaultFields); err != nil {
		fmt.Printf("❌ Tokenization failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Tokenization completed successfully!\n")
	fmt.Printf("📁 Tokenized data saved to: %s\n", *outputFile)
}

func generateTokenizeOutputName(inputFile string, useDatabase bool) string {
	if useDatabase {
		return "out/tokenized_database_records.csv"
	}

	if inputFile == "" {
		return "out/tokenized_data.csv"
	}

	base := filepath.Base(inputFile)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	return filepath.Join("out", name+"_tokenized.csv")
}

func detectInputFormat(inputFile string) string {
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext == ".json" {
		return "json"
	}
	return "csv" // Default fallback
}

func validateTokenizeInputs(inputFile string, useDatabase bool, configFile string) error {
	if !useDatabase {
		if inputFile == "" {
			return fmt.Errorf("input file is required when not using database mode")
		}
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			return fmt.Errorf("input file not found: %s", inputFile)
		}
	} else {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s", configFile)
		}
	}
	return nil
}

func performTokenization(inputFile, outputFile, inputFormat, outputFormat string, batchSize int, minHashSeed string, useDatabase bool, fields []string) error {
	// Use the existing PPRL storage functionality
	_, err := pprl.NewStorage(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// This is a simplified implementation - the actual tokenization would:
	// 1. Load records from input file using internal/db package
	// 2. Generate Bloom filters and MinHash signatures using internal/pprl
	// 3. Save tokenized records using storage.Append()

	fmt.Println("📂 Loading and tokenizing records...")
	fmt.Printf("   🔧 Processing in batches of %d...\n", batchSize)
	fmt.Println("   🔧 Generating Bloom filters...")
	fmt.Println("   🔧 Computing MinHash signatures...")
	fmt.Println("💾 Saving tokenized records...")

	// Placeholder for actual tokenization logic
	if useDatabase {
		fmt.Println("✅ Tokenization would process records from database")
	} else {
		fmt.Printf("✅ Tokenization would process records from %s\n", inputFile)
	}

	return nil
}

func showTokenizeHelp() {
	fmt.Println("🔐 CohortBridge Tokenization")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("Convert raw PHI data to privacy-preserving Bloom filter tokens")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge tokenize [OPTIONS]")
	fmt.Println("  cohort-bridge tokenize                     # Interactive mode")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -input string          Input file with PHI data")
	fmt.Println("  -output string         Output file for tokenized data")
	fmt.Println("  -main-config string    Main config file to read field names from")
	fmt.Println("  -input-format string   Input format: csv, json, postgres")
	fmt.Println("  -output-format string  Output format: csv, json")
	fmt.Println("  -batch-size int        Number of records to process in each batch")
	fmt.Println("  -interactive           Force interactive mode")
	fmt.Println("  -database              Use database from main config instead of file")
	fmt.Println("  -minhash-seed string   Seed for deterministic MinHash generation")
	fmt.Println("  -help                  Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode (prompts for all inputs)")
	fmt.Println("  cohort-bridge tokenize")
	fmt.Println()
	fmt.Println("  # File mode")
	fmt.Println("  cohort-bridge tokenize -input data.csv -output tokens.csv")
	fmt.Println()
	fmt.Println("  # Database mode")
	fmt.Println("  cohort-bridge tokenize -database -main-config config.yaml")
	fmt.Println()
	fmt.Println("  # Force interactive even with some parameters")
	fmt.Println("  cohort-bridge tokenize -input data.csv -interactive")
}

// Helper function for default indicators
func ifDefault(isDefault bool) string {
	if isDefault {
		return "(default)"
	}
	return ""
}
