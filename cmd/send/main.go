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

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/server"
)

type SendConfig struct {
	IntersectionFile string `json:"intersection_file"` // Path to intersection results
	TargetHost       string `json:"target_host"`       // Destination host
	TargetPort       int    `json:"target_port"`       // Destination port
	ConfigFile       string `json:"config_file"`       // Main config file for network settings
	DataFile         string `json:"data_file"`         // Optional raw data file to send matched records
	Mode             string `json:"mode"`              // Send mode: "intersection" or "matched_data"
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
	var (
		intersectionFile = flag.String("intersection", "", "Path to intersection results file (required)")
		targetHost       = flag.String("host", "", "Target host to send data to")
		targetPort       = flag.Int("port", 0, "Target port to send data to")
		configFile       = flag.String("config", "config.yaml", "Configuration file for network settings")
		dataFile         = flag.String("data", "", "Optional raw data file to send matched records")
		mode             = flag.String("mode", "intersection", "Send mode: 'intersection' or 'matched_data'")
		help             = flag.Bool("help", false, "Show help message")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *intersectionFile == "" {
		fmt.Println("Error: Intersection file is required")
		showHelp()
		os.Exit(1)
	}

	// Load configuration if needed for network settings
	var cfg *config.Config
	var err error
	if *configFile != "" {
		cfg, err = config.Load(*configFile)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	// Override with command line options if provided
	sendConfig := &SendConfig{
		IntersectionFile: *intersectionFile,
		ConfigFile:       *configFile,
		DataFile:         *dataFile,
		Mode:             *mode,
	}

	if *targetHost != "" {
		sendConfig.TargetHost = *targetHost
	} else if cfg != nil {
		sendConfig.TargetHost = cfg.Peer.Host
	}

	if *targetPort != 0 {
		sendConfig.TargetPort = *targetPort
	} else if cfg != nil {
		sendConfig.TargetPort = cfg.Peer.Port
	}

	if sendConfig.TargetHost == "" || sendConfig.TargetPort == 0 {
		fmt.Println("Error: Target host and port must be specified")
		showHelp()
		os.Exit(1)
	}

	fmt.Println("📤 CohortBridge Data Sender")
	fmt.Printf("📁 Intersection file: %s\n", sendConfig.IntersectionFile)
	fmt.Printf("🎯 Target: %s:%d\n", sendConfig.TargetHost, sendConfig.TargetPort)
	fmt.Printf("📊 Mode: %s\n", sendConfig.Mode)
	if sendConfig.DataFile != "" {
		fmt.Printf("📂 Data file: %s\n", sendConfig.DataFile)
	}

	// Perform the send operation
	if err := performSend(sendConfig, cfg); err != nil {
		log.Fatalf("Send operation failed: %v", err)
	}

	fmt.Println("✅ Data sent successfully")
}

func performSend(config *SendConfig, cfg *config.Config) error {
	// Load intersection results
	fmt.Println("📂 Loading intersection results...")
	intersectionResults, err := loadIntersectionResultsCSV(config.IntersectionFile)
	if err != nil {
		return fmt.Errorf("failed to load intersection results: %w", err)
	}

	matches := 0
	for _, result := range intersectionResults {
		if result.IsMatch {
			matches++
		}
	}
	fmt.Printf("   ✅ Loaded %d intersection results (%d matches)\n", len(intersectionResults), matches)

	switch config.Mode {
	case "intersection":
		return sendIntersectionResults(intersectionResults, config, cfg)
	case "matched_data":
		return sendMatchedData(intersectionResults, config, cfg)
	default:
		return fmt.Errorf("invalid mode: %s (valid modes: intersection, matched_data)", config.Mode)
	}
}

func sendIntersectionResults(results []IntersectionResult, config *SendConfig, cfg *config.Config) error {
	fmt.Println("📤 Sending intersection results...")

	// Convert intersection results to the format expected by the receiver
	var matchData server.MatchingData
	matchData.Records = make(map[string]string)

	for _, result := range results {
		if result.IsMatch {
			// For intersection results, we just send the match information
			// The receiver can verify this against their own intersection
			matchData.Records[result.ID1] = result.ID2
		}
	}

	fmt.Printf("   📊 Sending %d matched pairs\n", len(matchData.Records))

	if len(matchData.Records) == 0 {
		fmt.Println("   ℹ️  No matches found - skipping network transmission")
		return nil
	}

	// Use the existing server sender functionality with the match data
	if cfg != nil {
		return server.SendIntersectionData(cfg, &matchData)
	}

	// If no config provided, we need host and port from command line
	return fmt.Errorf("configuration file required when host/port not provided via config file")
}

func sendMatchedData(results []IntersectionResult, config *SendConfig, cfg *config.Config) error {
	if config.DataFile == "" {
		return fmt.Errorf("data file is required for matched_data mode")
	}

	fmt.Println("📤 Sending matched raw data...")
	fmt.Printf("📂 Loading data from: %s\n", config.DataFile)

	// Load the raw data file
	rawData, err := loadRawDataFile(config.DataFile)
	if err != nil {
		return fmt.Errorf("failed to load raw data: %w", err)
	}

	// Filter raw data to only include matched records
	var matchedData []map[string]string
	matchedIDs := make(map[string]bool)

	// Collect all matched IDs
	for _, result := range results {
		if result.IsMatch {
			matchedIDs[result.ID1] = true
		}
	}

	// Filter raw data
	for _, record := range rawData {
		if id, exists := record["id"]; exists && matchedIDs[id] {
			matchedData = append(matchedData, record)
		}
	}

	fmt.Printf("   📊 Sending %d matched records from %d total records\n",
		len(matchedData), len(rawData))

	// Send the matched data
	if cfg != nil {
		return server.SendMatchedRecords(cfg, matchedData)
	}

	// If no config provided, we need host and port from command line
	return fmt.Errorf("configuration file required when host/port not provided via config file")
}

func loadIntersectionResultsCSV(filename string) ([]IntersectionResult, error) {
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

	if len(rows) < 1 {
		return nil, fmt.Errorf("CSV file must have at least a header row")
	}

	// Handle empty intersection results (only header, no data rows)
	if len(rows) == 1 {
		return []IntersectionResult{}, nil
	}

	// Expected header: id1,id2,is_match,hamming_distance,jaccard_similarity,match_score,timestamp
	headers := rows[0]
	expectedHeaders := []string{"id1", "id2", "is_match", "hamming_distance", "jaccard_similarity", "match_score", "timestamp"}

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

	var results []IntersectionResult
	for i, row := range rows[1:] { // Skip header row
		if len(row) < len(expectedHeaders) {
			return nil, fmt.Errorf("row %d has insufficient columns", i+2)
		}

		// Parse boolean
		isMatch, err := strconv.ParseBool(row[headerMap["is_match"]])
		if err != nil {
			return nil, fmt.Errorf("invalid is_match value in row %d: %s", i+2, row[headerMap["is_match"]])
		}

		// Parse integer
		hammingDistance, err := strconv.ParseUint(row[headerMap["hamming_distance"]], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid hamming_distance value in row %d: %s", i+2, row[headerMap["hamming_distance"]])
		}

		// Parse floats
		jaccardSimilarity, err := strconv.ParseFloat(row[headerMap["jaccard_similarity"]], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid jaccard_similarity value in row %d: %s", i+2, row[headerMap["jaccard_similarity"]])
		}

		matchScore, err := strconv.ParseFloat(row[headerMap["match_score"]], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid match_score value in row %d: %s", i+2, row[headerMap["match_score"]])
		}

		result := IntersectionResult{
			ID1:               row[headerMap["id1"]],
			ID2:               row[headerMap["id2"]],
			IsMatch:           isMatch,
			HammingDistance:   uint32(hammingDistance),
			JaccardSimilarity: jaccardSimilarity,
			MatchScore:        matchScore,
			Timestamp:         row[headerMap["timestamp"]],
		}
		results = append(results, result)
	}

	return results, nil
}

func loadRawDataFile(filename string) ([]map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}

func showHelp() {
	fmt.Println("CohortBridge Data Sender")
	fmt.Println("========================")
	fmt.Println()
	fmt.Println("Send intersection results or matched data to another CohortBridge receiver.")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  send [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -intersection string")
	fmt.Println("        Path to intersection results file (required)")
	fmt.Println("  -host string")
	fmt.Println("        Target host to send data to")
	fmt.Println("  -port int")
	fmt.Println("        Target port to send data to")
	fmt.Println("  -config string")
	fmt.Println("        Configuration file for network settings (default: config.yaml)")
	fmt.Println("  -data string")
	fmt.Println("        Optional raw data file to send matched records")
	fmt.Println("  -mode string")
	fmt.Println("        Send mode: 'intersection' or 'matched_data' (default: intersection)")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("MODES:")
	fmt.Println("  intersection   - Send only the intersection results (ID pairs)")
	fmt.Println("  matched_data   - Send actual data records for matched pairs")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Send intersection results")
	fmt.Println("  ./send -intersection=intersection_results.csv -host=peer.example.com -port=8080")
	fmt.Println()
	fmt.Println("  # Send matched data using config file")
	fmt.Println("  ./send -intersection=results.csv -data=raw_data.json \\")
	fmt.Println("    -mode=matched_data -config=sender_config.yaml")
	fmt.Println()
	fmt.Println("INPUT FORMATS:")
	fmt.Println()
	fmt.Println("Intersection Results (CSV):")
	fmt.Println("  Headers: id1,id2,is_match,hamming_distance,jaccard_similarity,match_score,timestamp")
	fmt.Println()
	fmt.Println("Raw Data (JSON):")
	fmt.Println("  Array of objects with record data")
	fmt.Println()
	fmt.Println("NOTES:")
	fmt.Println("  - The receiver must be running and listening for connections")
	fmt.Println("  - Network configuration can be specified via config file or command line")
	fmt.Println("  - In 'matched_data' mode, only records with matches are sent")
}
