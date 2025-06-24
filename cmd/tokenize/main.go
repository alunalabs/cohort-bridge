package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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
	InputFormat         string   `json:"input_format"`
	OutputFormat        string   `json:"output_format"`
	BatchSize           int      `json:"batch_size"`
	Fields              []string `json:"fields"`
	BloomFilterSize     uint32   `json:"bloom_filter_size"`
	BloomHashCount      uint32   `json:"bloom_hash_count"`
	MinHashSignatures   uint32   `json:"minhash_signatures"`
	MinHashPermutations uint32   `json:"minhash_permutations"`
	RandomBitsPercent   float64  `json:"random_bits_percent"`
	QGramLength         int      `json:"qgram_length"`
	UseDatabase         bool     `json:"use_database"`
	DatabaseConfig      string   `json:"database_config"`
}

// RecordReader interface for streaming input
type RecordReader interface {
	Read() (map[string]string, error)
	Close() error
}

// RecordWriter interface for streaming output
type RecordWriter interface {
	Write(record TokenizedRecord) error
	Close() error
}

// CSV Reader implementation
type CSVRecordReader struct {
	file    *os.File
	reader  *csv.Reader
	headers []string
}

func (r *CSVRecordReader) Read() (map[string]string, error) {
	record, err := r.reader.Read()
	if err != nil {
		if err == io.EOF {
			return nil, errors.New("EOF")
		}
		return nil, err
	}

	result := make(map[string]string)
	for i, value := range record {
		if i < len(r.headers) {
			result[r.headers[i]] = value
		}
	}
	return result, nil
}

func (r *CSVRecordReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// JSON Reader implementation
type JSONRecordReader struct {
	file    *os.File
	scanner *bufio.Scanner
}

func (r *JSONRecordReader) Read() (map[string]string, error) {
	if !r.scanner.Scan() {
		if r.scanner.Err() != nil {
			return nil, r.scanner.Err()
		}
		return nil, errors.New("EOF")
	}

	var record map[string]interface{}
	if err := json.Unmarshal(r.scanner.Bytes(), &record); err != nil {
		return nil, err
	}

	// Convert all values to strings
	result := make(map[string]string)
	for k, v := range record {
		if v != nil {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

func (r *JSONRecordReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// Database Reader implementation
type DatabaseRecordReader struct {
	database  db.Database
	offset    int
	batchSize int
	records   []map[string]string
	index     int
}

func (r *DatabaseRecordReader) Read() (map[string]string, error) {
	// If we've consumed all records in current batch, fetch next batch
	if r.index >= len(r.records) {
		var err error
		r.records, err = r.database.List(r.offset, r.batchSize)
		if err != nil {
			return nil, err
		}

		if len(r.records) == 0 {
			return nil, errors.New("EOF")
		}

		r.offset += len(r.records)
		r.index = 0
	}

	record := r.records[r.index]
	r.index++
	return record, nil
}

func (r *DatabaseRecordReader) Close() error {
	return nil
}

// JSON Writer implementation
type JSONRecordWriter struct {
	file  *os.File
	first bool
}

func (w *JSONRecordWriter) Write(record TokenizedRecord) error {
	if w.first {
		if _, err := w.file.WriteString("[\n"); err != nil {
			return err
		}
		w.first = false
	} else {
		if _, err := w.file.WriteString(",\n"); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(record, "  ", "  ")
	if err != nil {
		return err
	}

	_, err = w.file.Write(data)
	return err
}

func (w *JSONRecordWriter) Close() error {
	if !w.first {
		if _, err := w.file.WriteString("\n]\n"); err != nil {
			return err
		}
	} else {
		// Empty array case
		if _, err := w.file.WriteString("[]\n"); err != nil {
			return err
		}
	}

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// CSV Writer implementation
type CSVRecordWriter struct {
	file    *os.File
	writer  *csv.Writer
	headers []string
	first   bool
}

func (w *CSVRecordWriter) Write(record TokenizedRecord) error {
	if w.first {
		w.headers = []string{"id", "bloom_filter", "minhash", "timestamp"}
		if err := w.writer.Write(w.headers); err != nil {
			return err
		}
		w.first = false
	}

	row := []string{record.ID, record.BloomFilter, record.MinHash, record.Timestamp}
	return w.writer.Write(row)
}

func (w *CSVRecordWriter) Close() error {
	if w.writer != nil {
		w.writer.Flush()
		if err := w.writer.Error(); err != nil {
			return err
		}
	}

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func main() {
	fmt.Println("🔐 PPRL Tokenization Tool")
	fmt.Println("=========================")
	fmt.Println("Converts raw PHI data to privacy-preserving Bloom filter tokens")
	fmt.Println()

	var (
		configFile     = flag.String("config", "", "Configuration file (optional)")
		mainConfigFile = flag.String("main-config", "config.yaml", "Main config file to read field names from")
		inputFile      = flag.String("input", "", "Input file with PHI data")
		outputFile     = flag.String("output", "", "Output file for tokenized data")
		inputFormat    = flag.String("input-format", "", "Input format: csv, json, postgres (auto-detect if not specified)")
		outputFormat   = flag.String("output-format", "", "Output format: csv, json (auto-detect from extension if not specified)")
		batchSize      = flag.Int("batch-size", 1000, "Number of records to process in each batch")
		interactive    = flag.Bool("interactive", false, "Use interactive mode")
		useDatabase    = flag.Bool("database", false, "Use database from main config instead of file")
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
	} else if *interactive || (*inputFile == "" && *outputFile == "" && !*useDatabase) {
		// Interactive mode
		tokConfig, err = getInteractiveConfig(defaultFields, *useDatabase)
		if err != nil {
			log.Fatalf("Interactive configuration failed: %v", err)
		}
	} else if *useDatabase {
		// Database mode
		if *outputFile == "" {
			*outputFile = "out/tokens.csv"
		}
		tokConfig = &TokenizationConfig{
			InputFile:           "",
			OutputFile:          *outputFile,
			InputFormat:         "postgres",
			OutputFormat:        detectOutputFormat(*outputFile, *outputFormat),
			BatchSize:           *batchSize,
			Fields:              defaultFields,
			BloomFilterSize:     1000,
			BloomHashCount:      5,
			MinHashSignatures:   128,
			MinHashPermutations: 1000,
			RandomBitsPercent:   0.0,
			QGramLength:         2,
			UseDatabase:         true,
			DatabaseConfig:      *mainConfigFile,
		}
	} else {
		// Command line mode
		if *inputFile == "" || *outputFile == "" {
			fmt.Println("❌ Usage:")
			fmt.Println("  tokenize -input data.csv -output tokens.csv")
			fmt.Println("  tokenize -input data.csv -output tokens.csv -main-config config.yaml")
			fmt.Println("  tokenize -config tokenize_config.yaml")
			fmt.Println("  tokenize -interactive")
			fmt.Println("  tokenize -database -main-config postgres_a.yaml -output tokens.csv")
			os.Exit(1)
		}
		tokConfig = &TokenizationConfig{
			InputFile:           *inputFile,
			OutputFile:          *outputFile,
			InputFormat:         detectInputFormat(*inputFile, *inputFormat),
			OutputFormat:        detectOutputFormat(*outputFile, *outputFormat),
			BatchSize:           *batchSize,
			Fields:              defaultFields,
			BloomFilterSize:     1000,
			BloomHashCount:      5,
			MinHashSignatures:   128,
			MinHashPermutations: 1000,
			RandomBitsPercent:   0.0,
			QGramLength:         2,
			UseDatabase:         false,
			DatabaseConfig:      "",
		}
	}

	fmt.Printf("📋 Tokenization Configuration:\n")
	if tokConfig.UseDatabase {
		fmt.Printf("  Database Config: %s\n", tokConfig.DatabaseConfig)
	} else {
		fmt.Printf("  Input File: %s\n", tokConfig.InputFile)
	}
	fmt.Printf("  Output File: %s\n", tokConfig.OutputFile)
	fmt.Printf("  Input Format: %s\n", tokConfig.InputFormat)
	fmt.Printf("  Output Format: %s\n", tokConfig.OutputFormat)
	fmt.Printf("  Batch Size: %d\n", tokConfig.BatchSize)
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

// detectInputFormat determines input format from file extension or explicit format
func detectInputFormat(filename, explicitFormat string) string {
	if explicitFormat != "" {
		return explicitFormat
	}

	if strings.HasSuffix(strings.ToLower(filename), ".json") {
		return "json"
	}
	return "csv" // default
}

// detectOutputFormat determines output format from file extension or explicit format
func detectOutputFormat(filename, explicitFormat string) string {
	if explicitFormat != "" {
		return explicitFormat
	}

	if strings.HasSuffix(strings.ToLower(filename), ".json") {
		return "json"
	}
	return "csv" // default
}

// createInputReader creates appropriate reader based on format
func createInputReader(tokConfig *TokenizationConfig) (RecordReader, error) {
	switch tokConfig.InputFormat {
	case "csv":
		file, err := os.Open(tokConfig.InputFile)
		if err != nil {
			return nil, err
		}

		reader := csv.NewReader(file)
		headers, err := reader.Read()
		if err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to read CSV headers: %w", err)
		}

		return &CSVRecordReader{
			file:    file,
			reader:  reader,
			headers: headers,
		}, nil

	case "json":
		file, err := os.Open(tokConfig.InputFile)
		if err != nil {
			return nil, err
		}

		return &JSONRecordReader{
			file:    file,
			scanner: bufio.NewScanner(file),
		}, nil

	case "postgres":
		cfg, err := config.Load(tokConfig.DatabaseConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}

		database, err := db.GetDatabaseFromConfig(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		fmt.Printf("✅ Connected to %s database\n", cfg.Database.Type)

		return &DatabaseRecordReader{
			database:  database,
			batchSize: tokConfig.BatchSize,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported input format: %s", tokConfig.InputFormat)
	}
}

// createOutputWriter creates appropriate writer based on format
func createOutputWriter(tokConfig *TokenizationConfig) (RecordWriter, error) {
	switch tokConfig.OutputFormat {
	case "json":
		file, err := os.Create(tokConfig.OutputFile)
		if err != nil {
			return nil, err
		}

		return &JSONRecordWriter{
			file:  file,
			first: true,
		}, nil

	case "csv":
		file, err := os.Create(tokConfig.OutputFile)
		if err != nil {
			return nil, err
		}

		return &CSVRecordWriter{
			file:   file,
			writer: csv.NewWriter(file),
			first:  true,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported output format: %s", tokConfig.OutputFormat)
	}
}

// processBatch processes a batch of records and writes them to output
func processBatch(batch []map[string]string, writer RecordWriter, tokConfig *TokenizationConfig, timestamp string, offset int) error {
	fmt.Printf("  Processing batch: %d-%d records\n", offset+1, offset+len(batch))

	for _, record := range batch {
		// Create Bloom filter
		bf := pprl.NewBloomFilterWithRandomBits(
			tokConfig.BloomFilterSize,
			tokConfig.BloomHashCount,
			tokConfig.RandomBitsPercent,
		)

		// Create MinHash
		mh, err := pprl.NewMinHash(tokConfig.MinHashPermutations, tokConfig.MinHashSignatures)
		if err != nil {
			return fmt.Errorf("failed to create MinHash: %w", err)
		}

		// Process each configured field
		for _, field := range tokConfig.Fields {
			if value, exists := record[field]; exists && value != "" {
				// Normalize field value
				normalized := normalizeField(value)

				// Generate q-grams
				qgrams := generateQGrams(normalized, tokConfig.QGramLength)

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

		// Write to output
		if err := writer.Write(tokenized); err != nil {
			return fmt.Errorf("failed to write tokenized record: %w", err)
		}
	}

	return nil
}

func performTokenization(tokConfig *TokenizationConfig) error {
	fmt.Printf("📖 Input format: %s, Output format: %s, Batch size: %d\n",
		tokConfig.InputFormat, tokConfig.OutputFormat, tokConfig.BatchSize)

	// Create input reader
	reader, err := createInputReader(tokConfig)
	if err != nil {
		return fmt.Errorf("failed to create input reader: %w", err)
	}
	defer reader.Close()

	// Create output writer
	writer, err := createOutputWriter(tokConfig)
	if err != nil {
		return fmt.Errorf("failed to create output writer: %w", err)
	}
	defer writer.Close()

	// Process records in batches
	batch := make([]map[string]string, 0, tokConfig.BatchSize)
	totalProcessed := 0
	timestamp := time.Now().UTC().Format(time.RFC3339)

	fmt.Printf("📊 Starting batch processing...\n")

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				// Process final batch
				if len(batch) > 0 {
					if err := processBatch(batch, writer, tokConfig, timestamp, totalProcessed); err != nil {
						return err
					}
					totalProcessed += len(batch)
				}
				break
			}
			return fmt.Errorf("failed to read record: %w", err)
		}

		batch = append(batch, record)

		// Process batch when it's full
		if len(batch) >= tokConfig.BatchSize {
			if err := processBatch(batch, writer, tokConfig, timestamp, totalProcessed); err != nil {
				return err
			}
			totalProcessed += len(batch)
			batch = batch[:0] // Reset batch
		}
	}

	fmt.Printf("✅ Processed %d total records\n", totalProcessed)
	return nil
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

func getInteractiveConfig(defaultFields []string, useDatabase bool) (*TokenizationConfig, error) {
	config := &TokenizationConfig{
		UseDatabase: useDatabase,
	}

	if !useDatabase {
		fmt.Print("📁 Input file path: ")
		fmt.Scanln(&config.InputFile)

		fmt.Print("📝 Input format (csv/json, default: auto-detect): ")
		var inputFmt string
		fmt.Scanln(&inputFmt)
		config.InputFormat = detectInputFormat(config.InputFile, inputFmt)
	} else {
		config.InputFormat = "postgres"
	}

	fmt.Print("📁 Output file path (default: out/tokens.json): ")
	var output string
	fmt.Scanln(&output)
	if output == "" {
		config.OutputFile = "out/tokens.json"
	} else {
		config.OutputFile = output
	}

	fmt.Print("📝 Output format (csv/json, default: auto-detect): ")
	var outputFmt string
	fmt.Scanln(&outputFmt)
	config.OutputFormat = detectOutputFormat(config.OutputFile, outputFmt)

	fmt.Print("📊 Batch size (default: 1000): ")
	var batchStr string
	fmt.Scanln(&batchStr)
	if batchStr == "" {
		config.BatchSize = 1000
	} else {
		fmt.Sscanf(batchStr, "%d", &config.BatchSize)
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
