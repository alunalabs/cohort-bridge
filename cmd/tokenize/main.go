package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

type TokenizedRecord struct {
	ID          string `json:"id"`
	BloomFilter string `json:"bloom_filter"` // Base64 encoded Bloom filter
	MinHash     string `json:"minhash"`      // Base64 encoded MinHash signature
	Timestamp   string `json:"timestamp"`    // When tokenized
}

type TokenizationConfig struct {
	InputFile           string   `json:"input_file"`
	OutputFile          string   `json:"output_file"`
	Fields              []string `json:"fields"`
	BloomFilterSize     uint32   `json:"bloom_filter_size"`
	BloomHashCount      uint32   `json:"bloom_hash_count"`
	MinHashSignatures   uint32   `json:"minhash_signatures"`
	MinHashPermutations uint32   `json:"minhash_permutations"`
	RandomBitsPercent   float64  `json:"random_bits_percent"`
	QGramLength         int      `json:"qgram_length"`
}

func main() {
	fmt.Println("🔐 PPRL Tokenization Tool")
	fmt.Println("=========================")
	fmt.Println("Converts raw PHI data to privacy-preserving Bloom filter tokens")
	fmt.Println()

	var (
		configFile     = flag.String("config", "", "Configuration file (optional)")
		mainConfigFile = flag.String("main-config", "config.yaml", "Main config file to read field names from")
		inputFile      = flag.String("input", "", "Input CSV file with PHI data")
		outputFile     = flag.String("output", "", "Output file for tokenized data")
		interactive    = flag.Bool("interactive", false, "Use interactive mode")
	)
	flag.Parse()

	// Ensure output directory exists
	if err := os.MkdirAll("out", 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
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

	var tokConfig *TokenizationConfig
	var err error

	if *configFile != "" {
		// Load from config file
		tokConfig, err = loadTokenizationConfig(*configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	} else if *interactive || (*inputFile == "" && *outputFile == "") {
		// Interactive mode
		tokConfig, err = getInteractiveConfig(defaultFields)
		if err != nil {
			log.Fatalf("Interactive configuration failed: %v", err)
		}
	} else {
		// Command line mode
		if *inputFile == "" || *outputFile == "" {
			fmt.Println("❌ Usage:")
			fmt.Println("  tokenize -input data.csv -output tokens.json")
			fmt.Println("  tokenize -input data.csv -output tokens.json -main-config config.yaml")
			fmt.Println("  tokenize -config tokenize_config.yaml")
			fmt.Println("  tokenize -interactive")
			os.Exit(1)
		}
		tokConfig = &TokenizationConfig{
			InputFile:           *inputFile,
			OutputFile:          *outputFile,
			Fields:              defaultFields,
			BloomFilterSize:     1000,
			BloomHashCount:      5,
			MinHashSignatures:   128,
			MinHashPermutations: 1000,
			RandomBitsPercent:   0.0,
			QGramLength:         2,
		}
	}

	fmt.Printf("📋 Tokenization Configuration:\n")
	fmt.Printf("  Input File: %s\n", tokConfig.InputFile)
	fmt.Printf("  Output File: %s\n", tokConfig.OutputFile)
	fmt.Printf("  Fields: %v\n", tokConfig.Fields)
	fmt.Printf("  Bloom Filter Size: %d bits\n", tokConfig.BloomFilterSize)
	fmt.Printf("  Hash Functions: %d\n", tokConfig.BloomHashCount)
	fmt.Printf("  MinHash Signatures: %d\n", tokConfig.MinHashSignatures)
	fmt.Printf("  Q-gram Length: %d\n", tokConfig.QGramLength)
	if tokConfig.RandomBitsPercent > 0 {
		fmt.Printf("  Random Bits: %.1f%%\n", tokConfig.RandomBitsPercent*100)
	}
	fmt.Println()

	// Perform tokenization
	if err := performTokenization(tokConfig); err != nil {
		log.Fatalf("Tokenization failed: %v", err)
	}

	fmt.Println("✅ Tokenization completed successfully!")
	fmt.Printf("📁 Tokenized data saved to: %s\n", tokConfig.OutputFile)
}

func loadTokenizationConfig(filename string) (*TokenizationConfig, error) {
	// For now, create a default config and save it as an example
	// Later this could load from YAML/JSON
	config := &TokenizationConfig{
		InputFile:           "data/patients.csv",
		OutputFile:          "out/tokens.json",
		Fields:              []string{"FIRST", "LAST", "BIRTHDATE", "ZIP"},
		BloomFilterSize:     1000,
		BloomHashCount:      5,
		MinHashSignatures:   128,
		MinHashPermutations: 1000,
		RandomBitsPercent:   0.0,
		QGramLength:         2,
	}

	// Save example config
	exampleFile := strings.Replace(filename, ".yaml", "_example.json", 1)
	if file, err := os.Create(exampleFile); err == nil {
		defer file.Close()
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		encoder.Encode(config)
		fmt.Printf("💡 Example config saved to: %s\n", exampleFile)
	}

	return config, nil
}

func getInteractiveConfig(defaultFields []string) (*TokenizationConfig, error) {
	config := &TokenizationConfig{}

	fmt.Print("📁 Input CSV file path: ")
	fmt.Scanln(&config.InputFile)

	fmt.Print("📁 Output file path (default: out/tokens.json): ")
	var output string
	fmt.Scanln(&output)
	if output == "" {
		config.OutputFile = "out/tokens.json"
	} else {
		config.OutputFile = output
	}

	fmt.Printf("📝 Fields to tokenize (comma-separated, default: %s): ", strings.Join(defaultFields, ","))
	var fieldsStr string
	fmt.Scanln(&fieldsStr)
	if fieldsStr == "" {
		config.Fields = defaultFields
	} else {
		config.Fields = strings.Split(fieldsStr, ",")
		for i, field := range config.Fields {
			config.Fields[i] = strings.TrimSpace(field)
		}
	}

	// Set reasonable defaults
	config.BloomFilterSize = 1000
	config.BloomHashCount = 5
	config.MinHashSignatures = 128
	config.MinHashPermutations = 1000
	config.RandomBitsPercent = 0.0
	config.QGramLength = 2

	return config, nil
}

func performTokenization(config *TokenizationConfig) error {
	// Load CSV data
	fmt.Printf("📖 Loading CSV data from: %s\n", config.InputFile)
	csvDB, err := db.NewCSVDatabase(config.InputFile)
	if err != nil {
		return fmt.Errorf("failed to load CSV: %w", err)
	}

	// Get all records
	allRecords, err := csvDB.List(0, 1000000)
	if err != nil {
		return fmt.Errorf("failed to list records: %w", err)
	}

	fmt.Printf("📊 Processing %d records...\n", len(allRecords))

	var tokenizedRecords []TokenizedRecord
	timestamp := time.Now().UTC().Format(time.RFC3339)

	for i, record := range allRecords {
		if i%100 == 0 {
			fmt.Printf("  Progress: %d/%d records\n", i, len(allRecords))
		}

		// Create Bloom filter
		bf := pprl.NewBloomFilterWithRandomBits(
			config.BloomFilterSize,
			config.BloomHashCount,
			config.RandomBitsPercent,
		)

		// Create MinHash
		mh, err := pprl.NewMinHash(config.MinHashPermutations, config.MinHashSignatures)
		if err != nil {
			return fmt.Errorf("failed to create MinHash: %w", err)
		}

		// Process each configured field
		for _, field := range config.Fields {
			if value, exists := record[field]; exists && value != "" {
				// Normalize field value
				normalized := normalizeField(value)

				// Generate q-grams
				qgrams := generateQGrams(normalized, config.QGramLength)

				// Add q-grams to Bloom filter
				for _, qgram := range qgrams {
					bf.Add([]byte(qgram))
				}
			}
		}

		// Compute MinHash signature
		_, err = mh.ComputeSignature(bf)
		if err != nil {
			return fmt.Errorf("failed to compute MinHash signature: %w", err)
		}

		// Encode to base64
		bloomData, err := pprl.BloomToBase64(bf)
		if err != nil {
			return fmt.Errorf("failed to encode Bloom filter: %w", err)
		}

		minHashData, err := pprl.MinHashToBase64(mh)
		if err != nil {
			return fmt.Errorf("failed to encode MinHash: %w", err)
		}

		// Create tokenized record
		tokenized := TokenizedRecord{
			ID:          record["id"],
			BloomFilter: bloomData,
			MinHash:     minHashData,
			Timestamp:   timestamp,
		}

		tokenizedRecords = append(tokenizedRecords, tokenized)
	}

	fmt.Printf("💾 Saving %d tokenized records...\n", len(tokenizedRecords))

	// Save to output file
	if err := saveTokenizedRecords(tokenizedRecords, config.OutputFile); err != nil {
		return fmt.Errorf("failed to save tokenized records: %w", err)
	}

	// Also save as CSV for compatibility
	csvOutputFile := strings.Replace(config.OutputFile, ".json", ".csv", 1)
	if err := saveTokenizedRecordsCSV(tokenizedRecords, csvOutputFile); err != nil {
		fmt.Printf("⚠️  Warning: Failed to save CSV format: %v\n", err)
	} else {
		fmt.Printf("📁 Also saved as CSV: %s\n", csvOutputFile)
	}

	return nil
}

func saveTokenizedRecords(records []TokenizedRecord, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(records)
}

func saveTokenizedRecordsCSV(records []TokenizedRecord, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id", "bloom_filter", "minhash", "timestamp"}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write records
	for _, record := range records {
		row := []string{
			record.ID,
			record.BloomFilter,
			record.MinHash,
			record.Timestamp,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func normalizeField(value string) string {
	// Convert to lowercase and remove spaces for consistent matching
	return strings.ToLower(strings.ReplaceAll(value, " ", ""))
}

func generateQGrams(text string, q int) []string {
	if len(text) < q {
		return []string{text}
	}

	// Use a map to store unique q-grams
	qgramSet := make(map[string]bool)

	// Add padding for beginning and end
	padded := strings.Repeat("_", q-1) + text + strings.Repeat("_", q-1)

	// Generate q-grams
	for i := 0; i <= len(padded)-q; i++ {
		qgram := padded[i : i+q]
		qgramSet[qgram] = true
	}

	// Convert to slice
	var qgrams []string
	for qgram := range qgramSet {
		qgrams = append(qgrams, qgram)
	}

	return qgrams
}
