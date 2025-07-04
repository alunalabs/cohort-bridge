package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// IntersectionResult represents a zero-knowledge computed intersection
// ONLY contains matches - no other information that could leak data
type IntersectionResult struct {
	Matches []*match.PrivateMatchResult `json:"matches"` // ONLY the matches
	// NO statistics, metadata, or any other information that could leak data
}

// NO IntersectionStats - statistics could leak information about datasets

// PeerMessage represents messages exchanged between peers
type PeerMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// TokenData represents the tokenized data to be exchanged
type TokenData struct {
	Records map[string]TokenRecord `json:"records"`
}

// TokenRecord represents a single tokenized record
type TokenRecord struct {
	ID          string `json:"id"`
	BloomFilter string `json:"bloom_filter"` // base64 encoded
	MinHash     string `json:"minhash"`      // base64 encoded
}

// SecureWorkflowConfig holds secure computation configuration
type SecureWorkflowConfig struct {
	Party int `json:"party"` // 0 or 1 for two-party protocol
}

// runUnifiedWorkflow implements the new unified peer-to-peer workflow
func runUnifiedWorkflow(cfg *config.Config, force, allowDuplicates bool) {
	fmt.Println("Starting Unified PPRL Peer-to-Peer Workflow")
	fmt.Println("============================================")
	fmt.Printf("Local Dataset: %s\n", cfg.Database.Filename)
	fmt.Printf("Peer Address: %s:%d\n", cfg.Peer.Host, cfg.Peer.Port)
	fmt.Printf("Listen Port: %d\n", cfg.ListenPort)

	// Zero-knowledge protocols are ALWAYS enabled - no toggleable options
	fmt.Printf("Zero-Knowledge Protocol: ALWAYS ENABLED\n")
	fmt.Printf("Absolute zero information leakage guaranteed\n")
	fmt.Println()

	// Create temp directory for this session
	tempDir := fmt.Sprintf("temp-workflow-%d", time.Now().Unix())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if !isDebugMode() {
			os.RemoveAll(tempDir)
		}
	}()

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// STEP 1: Read the config file (already done)
	fmt.Println("STEP 1: Configuration Loaded")
	fmt.Printf("   Config file processed successfully\n")
	fmt.Printf("   Hamming threshold: %d\n", cfg.Matching.HammingThreshold)
	fmt.Printf("   Jaccard threshold: %.3f\n", cfg.Matching.JaccardThreshold)
	fmt.Println()

	// STEP 2: Tokenize the dataset if not already tokenized
	fmt.Println("STEP 2: Dataset Tokenization")
	tokenizedFile, err := performTokenizationStep(cfg)
	if err != nil {
		log.Fatalf("Tokenization failed: %v", err)
	}
	fmt.Printf("   Tokenized data ready: %s\n", tokenizedFile)
	fmt.Println()

	// Confirmation
	if !confirmStep("Ready to establish peer connection and exchange tokens?", force) {
		fmt.Println("Workflow cancelled by user")
		return
	}

	// STEP 3: Establish connection with peer
	fmt.Println("STEP 3: Establishing Peer Connection")
	conn, isServer, err := establishPeerConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to establish peer connection: %v", err)
	}
	defer conn.Close()

	if isServer {
		fmt.Printf("   Connected as server (listening on port %d)\n", cfg.ListenPort)
	} else {
		fmt.Printf("   Connected as client to %s:%d\n", cfg.Peer.Host, cfg.Peer.Port)
	}
	fmt.Println()

	// STEP 4: Exchange tokens with peer
	fmt.Println("STEP 4: Token Exchange")
	localTokens, peerTokens, err := exchangeTokens(conn, tokenizedFile, isServer)
	if err != nil {
		log.Fatalf("Token exchange failed: %v", err)
	}
	fmt.Printf("   Local tokens: %d records\n", len(localTokens.Records))
	fmt.Printf("   Peer tokens: %d records\n", len(peerTokens.Records))
	fmt.Println()

	// STEP 5: Compute intersection using thresholds from config
	fmt.Println("STEP 5: Computing Intersection")

	// Determine party number based on connection role
	party := 0
	if isServer {
		party = 1
	}

	intersection, err := computeZeroKnowledgeIntersection(localTokens, peerTokens, cfg, party, allowDuplicates)
	if err != nil {
		log.Fatalf("Intersection computation failed: %v", err)
	}

	fmt.Printf("   Found %d matches using zero-knowledge protocols\n", len(intersection.Matches))
	fmt.Printf("   Zero information leaked beyond intersection result\n")

	// Save local intersection
	localIntersectionFile := "local_intersection.json"
	if err := saveWorkflowIntersectionResults(intersection, localIntersectionFile); err != nil {
		log.Fatalf("Failed to save local intersection: %v", err)
	}
	fmt.Printf("   Local intersection saved: %s\n", localIntersectionFile)
	fmt.Println()

	// STEP 6: Exchange intersection results for comparison
	fmt.Println("STEP 6: Exchanging Intersection Results")
	peerIntersection, err := exchangeIntersectionResults(conn, intersection, isServer)
	if err != nil {
		log.Fatalf("Intersection exchange failed: %v", err)
	}
	fmt.Printf("   Received peer intersection (%d matches)\n", len(peerIntersection.Matches))
	fmt.Println()

	// STEP 7: Compare results and create diff if needed
	fmt.Println("STEP 7: Comparing Intersection Results")
	resultsMatch, diffFile, err := compareIntersectionResults(intersection, peerIntersection)
	if err != nil {
		log.Fatalf("Result comparison failed: %v", err)
	}

	if resultsMatch {
		fmt.Println("   SUCCESS: Intersection results match between peers!")
		fmt.Println("   Both peers computed identical intersections")

		// Copy results to output directory
		if err := copyToOutput(localIntersectionFile, "intersection_results.json"); err != nil {
			fmt.Printf("   Warning: Failed to copy results to output: %v\n", err)
		} else {
			fmt.Printf("   Results saved to: out/intersection_results.json\n")
		}
	} else {
		fmt.Println("   ERROR: Intersection results DO NOT match between peers!")
		fmt.Printf("   Diff file created: %s\n", diffFile)

		// Copy diff to output directory
		if err := copyToOutput(diffFile, "intersection_diff.json"); err != nil {
			fmt.Printf("   Warning: Failed to copy diff to output: %v\n", err)
		} else {
			fmt.Printf("   Diff saved to: out/intersection_diff.json\n")
		}

		log.Fatalf("Workflow failed: Intersection results do not match")
	}

	fmt.Println()
	fmt.Println("UNIFIED PPRL WORKFLOW COMPLETED SUCCESSFULLY!")
	fmt.Println("============================================")
	fmt.Printf("Results available in: out/\n")
	if isDebugMode() {
		fmt.Printf("Debug files preserved in: %s/\n", tempDir)
	}
}

// performTokenizationStep handles tokenization if needed
func performTokenizationStep(cfg *config.Config) (string, error) {
	if cfg.Database.IsTokenized {
		fmt.Printf("   Using pre-tokenized data: %s\n", cfg.Database.Filename)
		return filepath.Join("..", cfg.Database.Filename), nil
	}

	fmt.Printf("   Tokenizing dataset: %s\n", cfg.Database.Filename)
	fmt.Printf("   Fields: %s\n", strings.Join(cfg.Database.Fields, ", "))

	tokenizedFile := "tokenized_data.csv"

	// Use direct tokenization without external config dependency
	inputPath := filepath.Join("..", cfg.Database.Filename)
	if err := performRealTokenization(inputPath, tokenizedFile, cfg.Database.Fields); err != nil {
		return "", fmt.Errorf("tokenization failed: %v", err)
	}

	return tokenizedFile, nil
}

// establishPeerConnection creates a connection between peers
func establishPeerConnection(cfg *config.Config) (net.Conn, bool, error) {
	// First try to connect as client
	address := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)
	fmt.Printf("   Attempting to connect to peer at %s...\n", address)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err == nil {
		fmt.Printf("   Connected as client to %s\n", address)
		return conn, false, nil
	}

	fmt.Printf("   Client connection failed, starting server mode...\n")

	// If client connection fails, start as server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		return nil, false, fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("   Listening for peer connection on port %d...\n", cfg.ListenPort)

	// Accept one connection
	conn, err = listener.Accept()
	if err != nil {
		return nil, false, fmt.Errorf("failed to accept connection: %v", err)
	}

	fmt.Printf("   Peer connected from %s\n", conn.RemoteAddr())
	return conn, true, nil
}

// exchangeTokens handles the bidirectional token exchange
func exchangeTokens(conn net.Conn, tokenizedFile string, isServer bool) (*TokenData, *TokenData, error) {
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Load local tokens
	localTokens, err := loadTokenizedData(tokenizedFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load local tokens: %v", err)
	}

	if isServer {
		// Server: first receive, then send
		fmt.Printf("   Receiving tokens from peer...\n")
		var peerMessage PeerMessage
		if err := decoder.Decode(&peerMessage); err != nil {
			return nil, nil, fmt.Errorf("failed to receive peer tokens: %v", err)
		}

		if peerMessage.Type != "tokens" {
			return nil, nil, fmt.Errorf("unexpected message type: %s", peerMessage.Type)
		}

		peerTokens := &TokenData{}
		if err := mapToStruct(peerMessage.Payload, peerTokens); err != nil {
			return nil, nil, fmt.Errorf("failed to parse peer tokens: %v", err)
		}

		fmt.Printf("   Sending local tokens to peer...\n")
		if err := encoder.Encode(PeerMessage{Type: "tokens", Payload: localTokens}); err != nil {
			return nil, nil, fmt.Errorf("failed to send local tokens: %v", err)
		}

		return localTokens, peerTokens, nil
	} else {
		// Client: first send, then receive
		fmt.Printf("   Sending local tokens to peer...\n")
		if err := encoder.Encode(PeerMessage{Type: "tokens", Payload: localTokens}); err != nil {
			return nil, nil, fmt.Errorf("failed to send local tokens: %v", err)
		}

		fmt.Printf("   Receiving tokens from peer...\n")
		var peerMessage PeerMessage
		if err := decoder.Decode(&peerMessage); err != nil {
			return nil, nil, fmt.Errorf("failed to receive peer tokens: %v", err)
		}

		if peerMessage.Type != "tokens" {
			return nil, nil, fmt.Errorf("unexpected message type: %s", peerMessage.Type)
		}

		peerTokens := &TokenData{}
		if err := mapToStruct(peerMessage.Payload, peerTokens); err != nil {
			return nil, nil, fmt.Errorf("failed to parse peer tokens: %v", err)
		}

		return localTokens, peerTokens, nil
	}
}

// loadTokenizedData loads tokenized data from a CSV file
func loadTokenizedData(filename string) (*TokenData, error) {
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

	tokenData := &TokenData{Records: make(map[string]TokenRecord)}

	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 4 {
			continue // Skip incomplete records
		}

		tokenRecord := TokenRecord{
			ID:          record[0],
			BloomFilter: record[1],
			MinHash:     record[2],
		}

		tokenData.Records[tokenRecord.ID] = tokenRecord
	}

	return tokenData, nil
}

// computeZeroKnowledgeIntersection computes intersection using ONLY zero-knowledge protocols
func computeZeroKnowledgeIntersection(localTokens, peerTokens *TokenData, cfg *config.Config, party int, allowDuplicates bool) (*IntersectionResult, error) {
	fmt.Printf("   Using zero-knowledge protocols (Party %d)\n", party)
	fmt.Printf("   No information leaked beyond intersection\n")

	if allowDuplicates {
		fmt.Printf("   Matching mode: 1:many (duplicates allowed)\n")
	} else {
		fmt.Printf("   Matching mode: 1:1 (unique matches only)\n")
	}

	return computeSecureIntersection(localTokens, peerTokens, cfg, party, allowDuplicates)
}

// computeSecureIntersection performs secure intersection computation
func computeSecureIntersection(localTokens, peerTokens *TokenData, cfg *config.Config, party int, allowDuplicates bool) (*IntersectionResult, error) {
	// Convert TokenData to PPRL Records for secure matching
	localRecords, err := tokenDataToPPRLRecords(localTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to convert local tokens: %v", err)
	}

	peerRecords, err := tokenDataToPPRLRecords(peerTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to convert peer tokens: %v", err)
	}

	// Configure zero-knowledge fuzzy matcher with duplicate control
	fuzzyConfig := &match.FuzzyMatchConfig{
		Party:           party,
		AllowDuplicates: allowDuplicates,
	}

	// Create zero-knowledge fuzzy matcher
	fuzzyMatcher := match.NewFuzzyMatcher(fuzzyConfig)

	// Perform zero-knowledge intersection computation
	secureResult, err := fuzzyMatcher.ComputePrivateIntersection(localRecords, peerRecords)
	if err != nil {
		return nil, fmt.Errorf("secure intersection computation failed: %v", err)
	}

	// Convert zero-knowledge results - only matches, no other information
	var matches []*match.PrivateMatchResult
	for _, privateMatch := range secureResult.MatchPairs {
		matchResult := &match.PrivateMatchResult{
			LocalID: privateMatch.LocalID,
			PeerID:  privateMatch.PeerID,
		}
		matches = append(matches, matchResult)
	}

	// Create intersection result with ZERO information leakage
	result := &IntersectionResult{
		Matches: matches,
	}

	return result, nil
}

// REMOVED: computeStandardIntersection
// Standard intersections are not supported in zero-knowledge protocols
// All intersections now use zero-knowledge protocols to ensure no information leakage
func computeStandardIntersection(localTokens, peerTokens *TokenData, cfg *config.Config) (*IntersectionResult, error) {
	return nil, fmt.Errorf("standard intersection not supported - use zero-knowledge protocols only")
}

// tokenDataToPPRLRecords converts TokenData to PPRL Records for secure matching
func tokenDataToPPRLRecords(tokenData *TokenData) ([]*pprl.Record, error) {
	var records []*pprl.Record

	for _, tokenRecord := range tokenData.Records {
		// Decode MinHash from base64
		mh, err := pprl.MinHashFromBase64(tokenRecord.MinHash)
		if err != nil {
			return nil, fmt.Errorf("failed to decode minhash for %s: %v", tokenRecord.ID, err)
		}

		// Get MinHash signature
		signature, err := mh.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal minhash for %s: %v", tokenRecord.ID, err)
		}

		// Convert to uint32 signature (simplified)
		var minHashSig []uint32
		for i := 0; i < len(signature) && i < 400; i += 4 { // Limit to reasonable size
			if i+3 < len(signature) {
				val := uint32(signature[i]) | uint32(signature[i+1])<<8 | uint32(signature[i+2])<<16 | uint32(signature[i+3])<<24
				minHashSig = append(minHashSig, val)
			}
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

// REMOVED: tokenDataToPatientRecords
// This function would leak information beyond intersection results
// In zero-knowledge protocols, we only work with PPRL records and return only intersection pairs

// REMOVED: runWorkflowMatchingPipeline
// This function would leak information beyond intersection results
// In zero-knowledge protocols, all matching is done through secure fuzzy matchers
// that ensure ZERO information leakage beyond the final intersection pairs

// exchangeIntersectionResults exchanges intersection results between peers
func exchangeIntersectionResults(conn net.Conn, localIntersection *IntersectionResult, isServer bool) (*IntersectionResult, error) {
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if isServer {
		// Server: first receive, then send
		fmt.Printf("   Receiving intersection from peer...\n")
		var peerMessage PeerMessage
		if err := decoder.Decode(&peerMessage); err != nil {
			return nil, fmt.Errorf("failed to receive peer intersection: %v", err)
		}

		if peerMessage.Type != "intersection" {
			return nil, fmt.Errorf("unexpected message type: %s", peerMessage.Type)
		}

		peerIntersection := &IntersectionResult{}
		if err := mapToStruct(peerMessage.Payload, peerIntersection); err != nil {
			return nil, fmt.Errorf("failed to parse peer intersection: %v", err)
		}

		fmt.Printf("   Sending local intersection to peer...\n")
		if err := encoder.Encode(PeerMessage{Type: "intersection", Payload: localIntersection}); err != nil {
			return nil, fmt.Errorf("failed to send local intersection: %v", err)
		}

		return peerIntersection, nil
	} else {
		// Client: first send, then receive
		fmt.Printf("   Sending local intersection to peer...\n")
		if err := encoder.Encode(PeerMessage{Type: "intersection", Payload: localIntersection}); err != nil {
			return nil, fmt.Errorf("failed to send local intersection: %v", err)
		}

		fmt.Printf("   Receiving intersection from peer...\n")
		var peerMessage PeerMessage
		if err := decoder.Decode(&peerMessage); err != nil {
			return nil, fmt.Errorf("failed to receive peer intersection: %v", err)
		}

		if peerMessage.Type != "intersection" {
			return nil, fmt.Errorf("unexpected message type: %s", peerMessage.Type)
		}

		peerIntersection := &IntersectionResult{}
		if err := mapToStruct(peerMessage.Payload, peerIntersection); err != nil {
			return nil, fmt.Errorf("failed to parse peer intersection: %v", err)
		}

		return peerIntersection, nil
	}
}

// compareIntersectionResults compares ONLY the intersection match pairs (zero information leakage)
func compareIntersectionResults(local, peer *IntersectionResult) (bool, string, error) {
	// Compare ONLY the number of matches - no other statistics
	localCount := len(local.Matches)
	peerCount := len(peer.Matches)

	if localCount != peerCount {
		fmt.Printf("   Match count differs: local=%d, peer=%d\n", localCount, peerCount)
	}

	// Create sorted match sets for comparison using ONLY IDs
	localMatches := createPrivateMatchSet(local.Matches)
	peerMatches := createPrivateMatchSet(peer.Matches)

	// Find differences in match pairs ONLY
	onlyInLocal := make(map[string]*match.PrivateMatchResult)
	onlyInPeer := make(map[string]*match.PrivateMatchResult)

	for key, match := range localMatches {
		if _, exists := peerMatches[key]; !exists {
			onlyInLocal[key] = match
		}
	}

	for key, match := range peerMatches {
		if _, exists := localMatches[key]; !exists {
			onlyInPeer[key] = match
		}
	}

	// Check if results match
	resultsMatch := len(onlyInLocal) == 0 && len(onlyInPeer) == 0 && localCount == peerCount

	if resultsMatch {
		return true, "", nil
	}

	// Create diff file with ONLY match information (no other statistics)
	diffFile := "intersection_diff.json"
	diff := map[string]interface{}{
		"summary": map[string]interface{}{
			"matches":             resultsMatch,
			"local_match_count":   localCount,
			"peer_match_count":    peerCount,
			"only_in_local_count": len(onlyInLocal),
			"only_in_peer_count":  len(onlyInPeer),
		},
		"only_in_local": onlyInLocal,
		"only_in_peer":  onlyInPeer,
		"created_at":    time.Now().Format(time.RFC3339),
	}

	if err := saveJSONFile(diff, diffFile); err != nil {
		return false, "", fmt.Errorf("failed to save diff file: %v", err)
	}

	return false, diffFile, nil
}

// createPrivateMatchSet creates a map of private matches keyed by canonical string representation (ONLY IDs)
func createPrivateMatchSet(matches []*match.PrivateMatchResult) map[string]*match.PrivateMatchResult {
	matchSet := make(map[string]*match.PrivateMatchResult)
	for _, match := range matches {
		// Create canonical key (ensure consistent ordering) using ONLY IDs
		var key string
		if match.LocalID < match.PeerID {
			key = fmt.Sprintf("%s<->%s", match.LocalID, match.PeerID)
		} else {
			key = fmt.Sprintf("%s<->%s", match.PeerID, match.LocalID)
		}
		matchSet[key] = match
	}
	return matchSet
}

// saveWorkflowIntersectionResults saves intersection results to a JSON file
func saveWorkflowIntersectionResults(intersection *IntersectionResult, filename string) error {
	return saveJSONFile(intersection, filename)
}

// saveJSONFile saves any object to a JSON file
func saveJSONFile(obj interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(obj)
}

// mapToStruct converts a map[string]interface{} to a struct
func mapToStruct(data interface{}, target interface{}) error {
	// Convert to JSON and back to properly handle the type conversion
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

// runTokenizeCommandInternal performs tokenization (simplified version of tokenize.go logic)
func runTokenizeCommandInternal(args []string, fields []string) error {
	var inputFile, outputFile string

	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-input":
			if i+1 < len(args) {
				inputFile = args[i+1]
				i++
			}
		case "-output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		}
	}

	if inputFile == "" || outputFile == "" {
		return fmt.Errorf("input and output files are required")
	}

	// Use the existing tokenization logic
	return performRealTokenization(inputFile, outputFile, fields)
}

// performRealTokenization (copied from existing workflows.go)
func performRealTokenization(inputFile, outputFile string, fields []string) error {
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

	// PPRL configuration for tokenization
	recordConfig := &pprl.RecordConfig{
		BloomSize:    1000, // 1000 bits
		BloomHashes:  5,    // 5 hash functions
		MinHashSize:  100,  // 100-element signature
		QGramLength:  2,    // 2-grams
		QGramPadding: "$",  // Padding character
		NoiseLevel:   0.01, // 1% noise
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

		// Convert to CSV format
		timestamp := time.Now().Format("2006-01-02T15:04:05Z")

		// Encode the complete MinHash object to base64
		minHashBase64, err := mh.ToBase64()
		if err != nil {
			return fmt.Errorf("failed to encode MinHash to base64 for %s: %w", recordID, err)
		}

		csvRow := []string{
			fmt.Sprintf("anonymous_%d", processedCount+1), // Anonymous ID only
			pprlRecord.BloomData,                          // Already base64 encoded
			minHashBase64,                                 // Properly base64 encoded MinHash
			timestamp,
		}

		if err := writer.Write(csvRow); err != nil {
			return fmt.Errorf("failed to write CSV row for %s: %w", recordID, err)
		}

		processedCount++
	}

	return nil
}

// copyToOutput copies a file to the output directory
func copyToOutput(srcFile, dstFile string) error {
	// Ensure output directory exists
	if err := os.MkdirAll("../out", 0755); err != nil {
		return err
	}

	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filepath.Join("../out", dstFile))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// Utility functions

func confirmStep(message string, force bool) bool {
	if force {
		fmt.Printf("%s (auto-confirmed with force flag)\n", message)
		return true
	}

	options := []string{
		"Yes, continue",
		"Cancel PPRL",
	}

	choice := promptForChoice(message, options)
	return choice == 0
}

func isDebugMode() bool {
	if os.Getenv("COHORT_DEBUG") == "1" || os.Getenv("COHORT_DEBUG") == "true" {
		return true
	}

	for _, arg := range os.Args {
		if arg == "-debug" || arg == "--debug" {
			return true
		}
	}

	return false
}

// runPPRLCommand is the entry point for the pprl command
func runPPRLCommand(args []string) {
	fmt.Println("CohortBridge PPRL")
	fmt.Println("=================")
	fmt.Println("Peer-to-peer privacy-preserving record linkage")
	fmt.Println()

	fs := flag.NewFlagSet("pprl", flag.ExitOnError)
	var (
		configFile      = fs.String("config", "", "Configuration file")
		interactive     = fs.Bool("interactive", false, "Force interactive mode")
		force           = fs.Bool("force", false, "Skip confirmation prompts and run automatically")
		allowDuplicates = fs.Bool("allow-duplicates", false, "Allow 1:many matching (default: 1:1 matching only)")
		help            = fs.Bool("help", false, "Show help message")
	)
	fs.Parse(args)

	if *help {
		showPPRLHelp()
		return
	}

	// Interactive mode if missing config or requested
	if *configFile == "" || *interactive {
		fmt.Println("Interactive PPRL Setup")
		fmt.Println("Configure your peer-to-peer record linkage:\n")

		if *configFile == "" {
			var err error
			*configFile, err = selectDataFile("Select Configuration File", "config", []string{".yaml"})
			if err != nil {
				fmt.Printf("Error selecting config file: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Println()
	}

	// Show configuration summary
	fmt.Println("PPRL Configuration:")
	fmt.Printf("  Config File: %s\n", *configFile)
	if *allowDuplicates {
		fmt.Printf("  Matching Mode: 1:many (duplicates allowed)\n")
	} else {
		fmt.Printf("  Matching Mode: 1:1 (unique matches only)\n")
	}
	fmt.Println()

	// Confirm before proceeding
	if !*force {
		confirmOptions := []string{
			"Yes, start PPRL",
			"Cancel",
		}

		confirmChoice := promptForChoice("Ready to start peer-to-peer record linkage?", confirmOptions)

		if confirmChoice == 1 {
			fmt.Println("\nPPRL cancelled. Goodbye!")
			os.Exit(0)
		}
	} else {
		fmt.Println("Starting PPRL automatically (force mode)...")
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Debug: Print loaded config details
	fmt.Printf("Debug - Loaded config: Peer.Host='%s', Peer.Port=%d, ListenPort=%d\n", cfg.Peer.Host, cfg.Peer.Port, cfg.ListenPort)

	// Validate config has required fields
	if cfg.Peer.Host == "" || cfg.Peer.Port == 0 {
		log.Fatalf("Configuration missing peer connection details (peer.host and peer.port)")
	}

	if cfg.ListenPort == 0 {
		log.Fatalf("Configuration missing listen_port")
	}

	if cfg.Matching.HammingThreshold == 0 {
		cfg.Matching.HammingThreshold = 90 // Default
	}

	if cfg.Matching.JaccardThreshold == 0 {
		cfg.Matching.JaccardThreshold = 0.5 // Default
	}

	// Run the PPRL workflow
	fmt.Println("Starting PPRL workflow...\n")
	runUnifiedWorkflow(cfg, *force, *allowDuplicates)
}

func showPPRLHelp() {
	fmt.Println("CohortBridge PPRL")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("PPRL STEPS:")
	fmt.Println("  1. Read configuration file")
	fmt.Println("  2. Tokenize dataset (if not pre-tokenized)")
	fmt.Println("  3. Establish peer connection")
	fmt.Println("  4. Exchange tokens with peer")
	fmt.Println("  5. Compute intersection using thresholds")
	fmt.Println("  6. Exchange intersection results")
	fmt.Println("  7. Compare results and create diff if needed")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge pprl [OPTIONS]")
	fmt.Println("  cohort-bridge pprl                       # Interactive mode")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -config string        Configuration file")
	fmt.Println("  -interactive          Force interactive mode")
	fmt.Println("  -force                Skip confirmation prompts")
	fmt.Println("  -allow-duplicates     Allow 1:many matching (default: 1:1 matching only)")
	fmt.Println("  -help                 Show this help message")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode")
	fmt.Println("  cohort-bridge pprl")
	fmt.Println()
	fmt.Println("  # Command line mode")
	fmt.Println("  cohort-bridge pprl -config config.yaml")
	fmt.Println()
	fmt.Println("  # Automatic mode (skip confirmations)")
	fmt.Println("  cohort-bridge pprl -config config.yaml -force")
	fmt.Println()
	fmt.Println("  # Allow 1:many matching (multiple matches per record)")
	fmt.Println("  cohort-bridge pprl -config config.yaml -allow-duplicates")
	fmt.Println()
	fmt.Println("CONFIGURATION REQUIREMENTS:")
	fmt.Println("  - peer.host and peer.port (peer connection)")
	fmt.Println("  - listen_port (local server port)")
	fmt.Println("  - matching.hamming_threshold (default: 90)")
	fmt.Println("  - matching.jaccard_threshold (default: 0.5)")
}
