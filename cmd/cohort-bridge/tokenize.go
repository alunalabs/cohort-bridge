package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

func runTokenizeCommand(args []string) {
	fmt.Println("PPRL Tokenization Tool")
	fmt.Println("======================")
	fmt.Println("Converts raw PHI data to privacy-preserving Bloom filter tokens")
	fmt.Println("Files are encrypted by default for maximum security")
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
		minHashSeed    = fs.String("minhash-seed", "0PsRm4KNmgRSY8ynApUtpXjeO19S7OUE", "Seed for deterministic MinHash generation")
		encryptionKey  = fs.String("encryption-key", "", "32-byte hex encryption key (auto-generated if empty)")
		noEncryption   = fs.Bool("no-encryption", false, "Disable encryption (not recommended for production)")
		force          = fs.Bool("force", false, "Skip confirmation prompts and run automatically")
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
		fmt.Println("Interactive Tokenization Setup")
		fmt.Println("Configure your tokenization parameters...")

		// Load config to get field information
		var defaultFields []string
		if cfg, err := config.Load(*mainConfigFile); err == nil {
			if len(cfg.Database.Fields) > 0 {
				defaultFields = cfg.Database.Fields
			}
		}
		if len(defaultFields) == 0 {
			defaultFields = []string{"FIRST", "LAST", "BIRTHDATE", "GENDER", "ZIP"}
		}

		// Choose data source
		if !*useDatabase {
			sourceChoice := promptForChoice("Select data source:", []string{
				"File - Process data from a file",
				"Database - Use database connection from config",
			})
			*useDatabase = (sourceChoice == 1)
		}

		// Get input file if using file mode
		if !*useDatabase && *inputFile == "" {
			var err error
			*inputFile, err = selectDataFile("Select Input Data File", "data", []string{".csv", ".json", ".txt"})
			if err != nil {
				fmt.Printf("Error selecting input file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get output file
		if *outputFile == "" {
			defaultOutput := generateOutputName("tokenized", *inputFile)
			*outputFile = promptForInput("Output file for tokenized data", defaultOutput)
		}

		// Configure encryption settings
		if !*noEncryption {
			fmt.Println("\nEncryption Configuration:")
			encryptChoice := promptForChoice("Encryption key source:", []string{
				"Auto-generate new key (recommended)",
				"Provide custom key (32-byte hex)",
				"Disable encryption (not recommended)",
			})

			switch encryptChoice {
			case 0:
				*encryptionKey = "" // Will be auto-generated
			case 1:
				customKey := promptForInput("Enter 32-byte hex encryption key", "")
				if len(customKey) != 64 {
					fmt.Println("Invalid key length, auto-generating instead...")
					*encryptionKey = ""
				} else {
					*encryptionKey = customKey
				}
			case 2:
				*noEncryption = true
				fmt.Println("Encryption disabled - files will be stored in plaintext!")
			}
		}

		// Output .enc file if it's encrypted
		if !*noEncryption {
			*outputFile = *outputFile + ".enc"
		}

		// Select input format with Auto-detect as default
		if !*useDatabase {
			fmt.Println("\nSelect input format (default: Auto-detect):")
			formatOptions := []string{
				"Auto-detect from file extension",
				"CSV - Comma-separated values",
				"JSON - JavaScript Object Notation",
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
			fmt.Sprintf("CSV - Comma-separated values %s", ifDefault(defaultOutputFormat == "csv")),
			fmt.Sprintf("JSON - JavaScript Object Notation %s", ifDefault(defaultOutputFormat == "json")),
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
			fmt.Println("Invalid batch size, using default:", *batchSize)
		}

		// Configure MinHash seed
		*minHashSeed = promptForInput("MinHash seed for deterministic hashing", *minHashSeed)

		fmt.Println()
	}

	// Try to load field names from main config file or CSV headers
	var defaultFields []string
	var normalizationConfig map[string]crypto.NormalizationMethod

	// If using CSV file input, read headers from CSV first
	if !*useDatabase && *inputFormat == "csv" && *inputFile != "" {
		csvFields, err := readCSVHeaders(*inputFile)
		if err == nil && len(csvFields) > 0 {
			defaultFields = csvFields
			fmt.Printf("Using field names from CSV headers: %v\n", defaultFields)
		}
	}

	// If no CSV headers found, try to load from config file
	if len(defaultFields) == 0 {
		if mainConfig, err := config.Load(*mainConfigFile); err == nil {
			if len(mainConfig.Database.Fields) > 0 {
				// Parse fields to extract field names and normalization
				defaultFields, normalizationConfig = parseFieldsWithNormalization(mainConfig.Database.Fields)
				fmt.Printf("Using field names from %s: %v\n", *mainConfigFile, defaultFields)
				if len(normalizationConfig) > 0 {
					fmt.Printf("Using normalization config: %v\n", normalizationConfig)
				}
			}
		}
	}

	// Fallback to defaults if no fields found
	if len(defaultFields) == 0 {
		defaultFields = []string{"FIRST", "LAST", "BIRTHDATE", "GENDER", "ZIP"}
		fmt.Printf("Could not load field names from config or CSV, using defaults: %v\n", defaultFields)
	}

	// Generate encryption key if needed
	var finalEncryptionKey string
	var keyFile string
	if !*noEncryption {
		if *encryptionKey == "" {
			// Auto-generate key
			key := make([]byte, 32)
			if _, err := rand.Read(key); err != nil {
				fmt.Printf("ERROR: Failed to generate encryption key: %v\n", err)
				os.Exit(1)
			}
			finalEncryptionKey = hex.EncodeToString(key)
			keyFile = generateKeyFileName(*outputFile)
		} else {
			finalEncryptionKey = *encryptionKey
		}

		// Automatically add .enc extension if encryption is enabled and not already present
		if !strings.HasSuffix(strings.ToLower(*outputFile), ".enc") {
			*outputFile = *outputFile + ".enc"
			// Update keyFile path if it was generated
			if keyFile != "" {
				keyFile = generateKeyFileName(*outputFile)
			}
		}
	}

	// Show configuration summary
	fmt.Println("Tokenization Configuration:")
	if *useDatabase {
		fmt.Println("  Data Source: Database (from config)")
	} else {
		fmt.Printf("  Input File: %s\n", *inputFile)
		fmt.Printf("  Input Format: %s\n", *inputFormat)
	}
	fmt.Printf("  Output File: %s\n", *outputFile)
	fmt.Printf("  Output Format: %s\n", *outputFormat)
	fmt.Printf("  Batch Size: %d\n", *batchSize)
	fmt.Printf("  Fields: %v\n", defaultFields)
	fmt.Printf("  MinHash Seed: %s\n", *minHashSeed)

	if !*noEncryption {
		fmt.Printf("  Encryption: AES-256-GCM (enabled)\n")
		if keyFile != "" {
			fmt.Printf("  Key Storage: %s\n", keyFile)
		} else {
			fmt.Printf("  Key Source: Custom provided\n")
		}
	} else {
		fmt.Printf("  Encryption: Disabled\n")
	}
	fmt.Println()

	// Confirm before proceeding (unless force flag is set)
	if !*force {
		confirmChoice := promptForChoice("Ready to start tokenization?", []string{
			"Yes, start tokenization",
			"Change configuration",
			"Cancel",
		})

		if confirmChoice == 2 {
			fmt.Println("\nTokenization cancelled. Goodbye!")
			os.Exit(0)
		}

		if confirmChoice == 1 {
			// Restart configuration
			fmt.Println("\nRestarting configuration...")
			newArgs := append([]string{"-interactive"}, args...)
			runTokenizeCommand(newArgs)
			return
		}
	} else {
		fmt.Println("Starting tokenization process automatically (force mode)...")
	}

	// Validate inputs before proceeding
	if err := validateTokenizeInputs(*inputFile, *useDatabase, *mainConfigFile); err != nil {
		fmt.Printf("ERROR: Validation error: %v\n", err)
		os.Exit(1)
	}

	// Run tokenization
	fmt.Println("Starting tokenization process...")

	if err := performTokenization(*inputFile, *outputFile, *inputFormat, *outputFormat, *batchSize, *minHashSeed, *useDatabase, defaultFields, finalEncryptionKey, keyFile, *noEncryption, normalizationConfig); err != nil {
		fmt.Printf("ERROR: Tokenization failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nTokenization completed successfully!\n")
	if !*noEncryption {
		fmt.Printf("Encrypted data saved to: %s\n", *outputFile)
		if keyFile != "" {
			fmt.Printf("Encryption key saved to: %s\n", keyFile)
			fmt.Printf("IMPORTANT: Save your encryption key securely! Without it, your data cannot be decrypted.\n")
		}
	} else {
		fmt.Printf("Tokenized data saved to: %s\n", *outputFile)
	}
}

// generateTokenizeOutputName function replaced with shared generateOutputName in utils.go

func generateKeyFileName(outputFile string) string {
	base := strings.TrimSuffix(outputFile, filepath.Ext(outputFile))
	return base + ".key"
}

func detectInputFormat(inputFile string) string {
	ext := strings.ToLower(filepath.Ext(inputFile))
	if ext == ".json" {
		return "json"
	}
	return "csv" // Default fallback
}

// readCSVHeaders reads the first line of a CSV file and returns the column headers
func readCSVHeaders(csvFile string) ([]string, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Clean up headers (trim whitespace and convert to uppercase)
	var cleanHeaders []string
	for _, header := range headers {
		cleaned := strings.TrimSpace(strings.ToUpper(header))
		if cleaned != "" {
			cleanHeaders = append(cleanHeaders, cleaned)
		}
	}

	return cleanHeaders, nil
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

func performTokenization(inputFile, outputFile, inputFormat, outputFormat string, batchSize int, minHashSeed string, useDatabase bool, fields []string, encryptionKey, keyFile string, noEncryption bool, normalizationConfig map[string]crypto.NormalizationMethod) error {
	if useDatabase {
		return fmt.Errorf("database mode not yet implemented - please use file mode")
	}

	// Load records from input file
	fmt.Println("Loading records from input file...")

	var allRecords []map[string]string

	if inputFormat == "csv" {
		// Use CSV database to load records
		csvDB, err := db.NewCSVDatabase(inputFile)
		if err != nil {
			return fmt.Errorf("failed to open CSV file: %w", err)
		}

		// Get all records from CSV
		allRecords, err = csvDB.List(0, 100000) // Load all records (up to 100k)
		if err != nil {
			return fmt.Errorf("failed to read records: %w", err)
		}
	} else {
		return fmt.Errorf("input format %s not yet implemented - please use CSV", inputFormat)
	}

	fmt.Printf("   Loaded %d records\n", len(allRecords))

	// Create output file
	fmt.Println("Creating output file...")

	if outputFormat == "csv" {
		return performCSVTokenization(allRecords, outputFile, fields, batchSize, minHashSeed, encryptionKey, keyFile, noEncryption, normalizationConfig)
	} else {
		return fmt.Errorf("output format %s not yet implemented - please use CSV", outputFormat)
	}
}

func performCSVTokenization(allRecords []map[string]string, outputFile string, fields []string, batchSize int, minHashSeed string, encryptionKey, keyFile string, noEncryption bool, normalizationConfig map[string]crypto.NormalizationMethod) error {
	// Determine if we need to encrypt
	var tempFile string
	var finalOutputFile string

	if !noEncryption {
		// Create temporary unencrypted file first
		tempFile = outputFile + ".tmp"
		finalOutputFile = outputFile
		outputFile = tempFile // Write to temp file first
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

	// PPRL configuration for tokenization
	recordConfig := &pprl.RecordConfig{
		BloomSize:    1000, // 1000 bits
		BloomHashes:  5,    // 5 hash functions
		MinHashSize:  100,  // 100-element signature
		QGramLength:  2,    // 2-grams
		QGramPadding: "$",  // Padding character
		NoiseLevel:   0.01, // 1% noise
	}

	fmt.Println("Processing records in batches...")
	fmt.Printf("   Batch size: %d\n", batchSize)
	fmt.Println("   Generating Bloom filters...")
	fmt.Println("   Computing MinHash signatures...")

	processedCount := 0
	totalRecords := len(allRecords)

	for i := 0; i < totalRecords; i += batchSize {
		end := i + batchSize
		if end > totalRecords {
			end = totalRecords
		}

		batch := allRecords[i:end]
		fmt.Printf("   Processing batch %d/%d (%d records)\n",
			(i/batchSize)+1,
			(totalRecords+batchSize-1)/batchSize,
			len(batch))

		for _, record := range batch {
			// Extract field values for this record
			var fieldValues []string
			for _, field := range fields {
				if value, exists := record[field]; exists && value != "" {
					// Apply normalization if configured
					var normalizedValue string
					if normalizationConfig != nil {
						if method, hasNorm := normalizationConfig[field]; hasNorm {
							normalizedValue = crypto.NormalizeField(value, method)
						} else {
							// No specific normalization configured, apply basic normalization
							normalizedValue = crypto.NormalizeField(value, "")
						}
					} else {
						// Basic normalization fallback
						normalizedValue = crypto.NormalizeField(value, "")
					}

					if normalizedValue != "" {
						fieldValues = append(fieldValues, normalizedValue)
					}
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

			// Create MinHash and compute signature from the Bloom filter
			mh, err := pprl.NewMinHash(recordConfig.BloomSize, recordConfig.MinHashSize)
			if err != nil {
				return fmt.Errorf("failed to create MinHash for %s: %w", recordID, err)
			}

			// Compute the signature directly from the Bloom filter
			_, err = mh.ComputeSignature(bf)
			if err != nil {
				return fmt.Errorf("failed to compute MinHash signature for %s: %w", recordID, err)
			}

			// Convert to CSV format with actual record ID
			timestamp := time.Now().Format("2006-01-02T15:04:05Z")

			// Encode the complete MinHash object to base64
			minHashBase64, err := mh.ToBase64()
			if err != nil {
				return fmt.Errorf("failed to encode MinHash to base64 for %s: %w", recordID, err)
			}

			// Write the tokenized record to CSV with the actual record ID
			row := []string{
				recordID, // Include the actual record ID
				pprlRecord.BloomData,
				minHashBase64,
				timestamp,
			}

			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write record to CSV: %w", err)
			}

			processedCount++
		}
	}

	// Close the file to ensure all data is written
	writer.Flush()
	outputCSV.Close()

	fmt.Printf("Successfully tokenized %d records\n", processedCount)

	// Handle encryption if enabled
	if !noEncryption {
		fmt.Println("Encrypting output file...")

		// Save encryption key to file if keyFile is specified
		if keyFile != "" {
			if err := saveKeyToFile(encryptionKey, keyFile); err != nil {
				// Cleanup temp file before returning error
				os.Remove(tempFile)
				return fmt.Errorf("failed to save encryption key: %w", err)
			}
			fmt.Printf("   Encryption key saved to: %s\n", keyFile)
		}

		// Encrypt the file
		if err := encryptFile(tempFile, finalOutputFile, encryptionKey); err != nil {
			// Cleanup temp file before returning error
			os.Remove(tempFile)
			return fmt.Errorf("failed to encrypt output file: %w", err)
		}

		// Secure cleanup of temporary file
		if err := secureDeleteFile(tempFile); err != nil {
			fmt.Printf("Warning: failed to securely delete temporary file: %v\n", err)
		}

		fmt.Printf("   File encrypted successfully with AES-256-GCM\n")
	}

	return nil
}

// encryptFile encrypts a file using AES-256-GCM
func encryptFile(inputFile, outputFile, keyHex string) error {
	// Decode hex key
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("invalid encryption key format: %w", err)
	}

	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	// Read plaintext file
	plaintext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write to output file with restricted permissions
	if err := os.WriteFile(outputFile, ciphertext, 0600); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return nil
}

// saveKeyToFile saves the encryption key to a file with restricted permissions
func saveKeyToFile(keyHex, keyFile string) error {
	keyData := fmt.Sprintf("# CohortBridge Encryption Key\n# Generated: %s\n# WARNING: Keep this key secure! Without it, your data cannot be decrypted.\n\n%s\n",
		time.Now().Format("2006-01-02 15:04:05"), keyHex)

	if err := os.WriteFile(keyFile, []byte(keyData), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// secureDeleteFile attempts to securely delete a file by overwriting it before removal
func secureDeleteFile(filename string) error {
	// Get file size
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	// Open file for writing
	file, err := os.OpenFile(filename, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	// Overwrite with random data
	size := info.Size()
	randomData := make([]byte, size)
	if _, err := rand.Read(randomData); err == nil {
		file.Write(randomData)
		file.Sync()
	}

	// Close and remove
	file.Close()
	return os.Remove(filename)
}

func showTokenizeHelp() {
	fmt.Println("CohortBridge Tokenization")
	fmt.Println("===========================")
	fmt.Println()
	fmt.Println("Convert raw PHI data to privacy-preserving Bloom filter tokens")
	fmt.Println("Files are encrypted by default using AES-256-GCM")
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
	fmt.Println("  -encryption-key string 32-byte hex encryption key (auto-generated if empty)")
	fmt.Println("  -no-encryption         Disable encryption (not recommended for production)")
	fmt.Println("  -force                 Skip confirmation prompts and run automatically")
	fmt.Println("  -help                  Show this help message")
	fmt.Println()
	fmt.Println("ENCRYPTION:")
	fmt.Println("  By default, output files are encrypted with AES-256-GCM.")
	fmt.Println("  - If no key is provided, one is auto-generated and saved")
	fmt.Println("  - Keep your encryption key safe! Data cannot be recovered without it")
	fmt.Println("  - Use -no-encryption to disable (not recommended for production)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode (prompts for all inputs)")
	fmt.Println("  cohort-bridge tokenize")
	fmt.Println()
	fmt.Println("  # File mode with auto-generated encryption")
	fmt.Println("  cohort-bridge tokenize -input data.csv -output tokens.csv.enc")
	fmt.Println()
	fmt.Println("  # Use custom encryption key")
	fmt.Println("  cohort-bridge tokenize -input data.csv -encryption-key a1b2c3d4e5f6789...")
	fmt.Println()
	fmt.Println("  # Automatic mode (skip confirmations)")
	fmt.Println("  cohort-bridge tokenize -input data.csv -output tokens.csv.enc -force")
	fmt.Println("  cohort-bridge tokenize -database -main-config config.yaml -force")
	fmt.Println()
	fmt.Println("  # Database mode")
	fmt.Println("  cohort-bridge tokenize -database -main-config config.yaml")
	fmt.Println()
	fmt.Println("  # Disable encryption (not recommended)")
	fmt.Println("  cohort-bridge tokenize -input data.csv -no-encryption")
	fmt.Println()
	fmt.Println("DECRYPT:")
	fmt.Println("  To decrypt an encrypted file:")
	fmt.Println("  cohort-bridge decrypt -input tokens.csv.enc -key path/to/file.key")
	fmt.Println("  cohort-bridge decrypt -input tokens.csv.enc -key-hex a1b2c3d4e5f6789...")
}

// Helper function for default indicators
// ifDefault function moved to utils.go

// DecryptFile decrypts a file encrypted with encryptFile
func DecryptFile(inputFile, outputFile, keyHex string) error {
	// Decode hex key
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return fmt.Errorf("invalid encryption key format: %w", err)
	}

	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	// Read encrypted file
	ciphertext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce and ciphertext
	if len(ciphertext) < gcm.NonceSize() {
		return fmt.Errorf("ciphertext too short")
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("failed to decrypt file (wrong key or corrupted data): %w", err)
	}

	// Write decrypted file
	if err := os.WriteFile(outputFile, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}

// LoadKeyFromFile loads an encryption key from a key file
func LoadKeyFromFile(keyFile string) (string, error) {
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read key file: %w", err)
	}

	// Extract hex key from file (skip comments)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			// Validate hex format
			if len(line) == 64 {
				if _, err := hex.DecodeString(line); err == nil {
					return line, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no valid encryption key found in file")
}

func runDecryptCommand(args []string) {
	fmt.Println("File Decryption Tool")
	fmt.Println("=======================")
	fmt.Println("Decrypt encrypted tokenized files")
	fmt.Println()

	fs := flag.NewFlagSet("decrypt", flag.ExitOnError)
	var (
		inputFile   = fs.String("input", "", "Encrypted input file")
		outputFile  = fs.String("output", "", "Decrypted output file")
		keyFile     = fs.String("key", "", "Path to encryption key file")
		keyHex      = fs.String("key-hex", "", "Encryption key as hex string")
		interactive = fs.Bool("interactive", false, "Force interactive mode")
		force       = fs.Bool("force", false, "Skip confirmation prompts")
		help        = fs.Bool("help", false, "Show help message")
	)
	fs.Parse(args)

	if *help {
		showDecryptHelp()
		return
	}

	// If missing required parameters or interactive mode requested, go interactive
	if *inputFile == "" || (*keyFile == "" && *keyHex == "") || *outputFile == "" || *interactive {
		fmt.Println("Interactive Decryption Setup")
		fmt.Println("Let's configure your decryption parameters...")

		// Get input file
		if *inputFile == "" {
			var err error
			*inputFile, err = selectDataFile("Select Encrypted File", "out", []string{".enc", ".encrypted"})
			if err != nil {
				fmt.Printf("ERROR: Error selecting input file: %v\n", err)
				os.Exit(1)
			}
		}

		// Get output file
		if *outputFile == "" {
			defaultOutput := generateDecryptOutputName(*inputFile)
			*outputFile = promptForInput("Output file for decrypted data", defaultOutput)
		}

		// Get encryption key
		if *keyFile == "" && *keyHex == "" {
			keyChoice := promptForChoice("How would you like to provide the encryption key?", []string{
				"Key file - Load from .key file",
				"Manual entry - Enter hex key directly",
			})

			if keyChoice == 0 {
				// Key file
				var err error
				*keyFile, err = selectDataFile("Select Key File", "out", []string{".key"})
				if err != nil {
					fmt.Printf("ERROR: Error selecting key file: %v\n", err)
					os.Exit(1)
				}
			} else {
				// Manual entry
				*keyHex = promptForInput("Enter 64-character hex encryption key", "")
				if len(*keyHex) != 64 {
					fmt.Printf("ERROR: Invalid key length. Expected 64 characters, got %d\n", len(*keyHex))
					os.Exit(1)
				}
			}
		}
	}

	// Load key from file if specified
	var finalKeyHex string
	if *keyFile != "" {
		var err error
		finalKeyHex, err = LoadKeyFromFile(*keyFile)
		if err != nil {
			fmt.Printf("ERROR: Failed to load key from file: %v\n", err)
			os.Exit(1)
		}
	} else {
		finalKeyHex = *keyHex
	}

	// Validate key format
	if len(finalKeyHex) != 64 {
		fmt.Printf("ERROR: Invalid key format. Expected 64 hex characters, got %d\n", len(finalKeyHex))
		os.Exit(1)
	}

	// Show configuration summary
	fmt.Println("Decryption Configuration:")
	fmt.Printf(" Input File: %s\n", *inputFile)
	fmt.Printf(" Output File: %s\n", *outputFile)
	if *keyFile != "" {
		fmt.Printf("  Key Source: File (%s)\n", *keyFile)
	} else {
		fmt.Printf("  Key Source: Manual entry\n")
	}
	fmt.Println()

	// Confirm before proceeding (unless force flag is set)
	if !*force {
		confirmChoice := promptForChoice("Ready to decrypt file?", []string{
			"Yes, decrypt now",
			"Change configuration",
			"Cancel",
		})

		if confirmChoice == 2 {
			fmt.Println("\nDecryption cancelled. Goodbye!")
			os.Exit(0)
		}

		if confirmChoice == 1 {
			// Restart configuration
			fmt.Println("\nRestarting configuration...\n")
			newArgs := append([]string{"-interactive"}, args...)
			runDecryptCommand(newArgs)
			return
		}
	} else {
		fmt.Println("Starting decryption process automatically (force mode)...")
	}

	// Validate input file exists
	if _, err := os.Stat(*inputFile); os.IsNotExist(err) {
		fmt.Printf("Input file not found: %s\n", *inputFile)
		os.Exit(1)
	}

	// Run decryption
	fmt.Println("Decrypting file...")

	if err := DecryptFile(*inputFile, *outputFile, finalKeyHex); err != nil {
		fmt.Printf("ERROR: Decryption failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nDecryption completed successfully!\n")
	fmt.Printf("Decrypted data saved to: %s\n", *outputFile)
	fmt.Printf("You can now view the tokenized data in plaintext format\n")
}

func generateDecryptOutputName(inputFile string) string {
	// Remove .enc extension if present
	if strings.HasSuffix(inputFile, ".enc") {
		return strings.TrimSuffix(inputFile, ".enc")
	}
	if strings.HasSuffix(inputFile, ".encrypted") {
		return strings.TrimSuffix(inputFile, ".encrypted")
	}

	// Add _decrypted suffix
	ext := filepath.Ext(inputFile)
	base := strings.TrimSuffix(inputFile, ext)
	return base + "_decrypted" + ext
}

func showDecryptHelp() {
	fmt.Println("CohortBridge File Decryption")
	fmt.Println("===============================")
	fmt.Println()
	fmt.Println("Decrypt files encrypted by the tokenize command")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge decrypt [OPTIONS]")
	fmt.Println("  cohort-bridge decrypt                          # Interactive mode")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -input string          Encrypted input file")
	fmt.Println("  -output string         Decrypted output file")
	fmt.Println("  -key string            Path to encryption key file")
	fmt.Println("  -key-hex string        Encryption key as 64-character hex string")
	fmt.Println("  -interactive           Force interactive mode")
	fmt.Println("  -force                 Skip confirmation prompts")
	fmt.Println("  -help                  Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode")
	fmt.Println("  cohort-bridge decrypt")
	fmt.Println()
	fmt.Println("  # Using key file")
	fmt.Println("  cohort-bridge decrypt -input tokens.csv.enc -key tokens.key")
	fmt.Println()
	fmt.Println("  # Using hex key directly")
	fmt.Println("  cohort-bridge decrypt -input tokens.csv.enc -key-hex a1b2c3d4e5f6789...")
	fmt.Println()
	fmt.Println("  # Specify output file")
	fmt.Println("  cohort-bridge decrypt -input tokens.csv.enc -key tokens.key -output readable.csv")
	fmt.Println()
	fmt.Println("  # Force mode (no confirmations)")
	fmt.Println("  cohort-bridge decrypt -input tokens.csv.enc -key tokens.key -force")
	fmt.Println()
	fmt.Println("NOTE:")
	fmt.Println("  You must have the correct encryption key to decrypt the file.")
	fmt.Println("  Keys are either saved as .key files or provided manually.")
}

// parseFieldsWithNormalization parses fields array that may contain normalization specs
// Input: ["name:FIRST", "LAST", "date:BIRTHDATE", "ZIP"]
// Output: (["FIRST", "LAST", "BIRTHDATE", "GENDER", "ZIP"], {"FIRST": "name", "BIRTHDATE": "date"})
func parseFieldsWithNormalization(fields []string) ([]string, map[string]crypto.NormalizationMethod) {
	var fieldNames []string
	normalizationConfig := make(map[string]crypto.NormalizationMethod)

	for _, field := range fields {
		// Check if field contains normalization specification (method:field)
		if strings.Contains(field, ":") {
			parts := strings.Split(field, ":")
			if len(parts) == 2 {
				method := strings.ToLower(strings.TrimSpace(parts[0]))
				fieldName := strings.TrimSpace(parts[1])

				// Add field name
				fieldNames = append(fieldNames, fieldName)

				// Add normalization method if supported
				switch method {
				case "name":
					normalizationConfig[fieldName] = crypto.NormName
				case "date":
					normalizationConfig[fieldName] = crypto.NormDate
				case "gender":
					normalizationConfig[fieldName] = crypto.NormGender
				case "zip":
					normalizationConfig[fieldName] = crypto.NormZip
				}
			} else {
				// Invalid format, just use as field name
				fieldNames = append(fieldNames, field)
			}
		} else {
			// Plain field name without normalization
			fieldNames = append(fieldNames, field)
		}
	}

	return fieldNames, normalizationConfig
}
