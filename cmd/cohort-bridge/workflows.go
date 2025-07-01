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
	"github.com/auroradata-ai/cohort-bridge/internal/server"
)

// IntersectionResult represents a computed intersection
type IntersectionResult struct {
	Matches []*match.MatchResult `json:"matches"`
	Stats   IntersectionStats    `json:"stats"`
}

// IntersectionStats provides statistics about the intersection
type IntersectionStats struct {
	TotalComparisons int     `json:"total_comparisons"`
	MatchCount       int     `json:"match_count"`
	MatchRate        float64 `json:"match_rate"`
	ComputedAt       string  `json:"computed_at"`
}

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

// runUnifiedWorkflow implements the new unified peer-to-peer workflow
func runUnifiedWorkflow(cfg *config.Config, force bool) {
	fmt.Println("🔄 Starting Unified PPRL Peer-to-Peer Workflow")
	fmt.Println("==============================================")
	fmt.Printf("Local Dataset: %s\n", cfg.Database.Filename)
	fmt.Printf("Peer Address: %s:%d\n", cfg.Peer.Host, cfg.Peer.Port)
	fmt.Printf("Listen Port: %d\n", cfg.ListenPort)
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
	fmt.Println("📋 STEP 1: Configuration Loaded")
	fmt.Printf("   ✓ Config file processed successfully\n")
	fmt.Printf("   ✓ Hamming threshold: %d\n", cfg.Matching.HammingThreshold)
	fmt.Printf("   ✓ Jaccard threshold: %.3f\n", cfg.Matching.JaccardThreshold)
	fmt.Println()

	// STEP 2: Tokenize the dataset if not already tokenized
	fmt.Println("🔧 STEP 2: Dataset Tokenization")
	tokenizedFile, err := performTokenizationStep(cfg)
	if err != nil {
		log.Fatalf("❌ Tokenization failed: %v", err)
	}
	fmt.Printf("   ✅ Tokenized data ready: %s\n", tokenizedFile)
	fmt.Println()

	// Confirmation
	if !confirmStep("Ready to establish peer connection and exchange tokens?", force) {
		fmt.Println("👋 Workflow cancelled by user")
		return
	}

	// STEP 3: Establish connection with peer
	fmt.Println("📡 STEP 3: Establishing Peer Connection")
	conn, isServer, err := establishPeerConnection(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to establish peer connection: %v", err)
	}
	defer conn.Close()

	if isServer {
		fmt.Printf("   ✅ Connected as server (listening on port %d)\n", cfg.ListenPort)
	} else {
		fmt.Printf("   ✅ Connected as client to %s:%d\n", cfg.Peer.Host, cfg.Peer.Port)
	}
	fmt.Println()

	// STEP 4: Exchange tokens with peer
	fmt.Println("🔄 STEP 4: Token Exchange")
	localTokens, peerTokens, err := exchangeTokens(conn, tokenizedFile, isServer)
	if err != nil {
		log.Fatalf("❌ Token exchange failed: %v", err)
	}
	fmt.Printf("   ✅ Local tokens: %d records\n", len(localTokens.Records))
	fmt.Printf("   ✅ Peer tokens: %d records\n", len(peerTokens.Records))
	fmt.Println()

	// STEP 5: Compute intersection using thresholds from config
	fmt.Println("🔍 STEP 5: Computing Intersection")
	intersection, err := computeIntersection(localTokens, peerTokens, cfg)
	if err != nil {
		log.Fatalf("❌ Intersection computation failed: %v", err)
	}
	fmt.Printf("   ✅ Found %d matches from %d comparisons\n",
		intersection.Stats.MatchCount, intersection.Stats.TotalComparisons)

	// Save local intersection
	localIntersectionFile := "local_intersection.json"
	if err := saveWorkflowIntersectionResults(intersection, localIntersectionFile); err != nil {
		log.Fatalf("❌ Failed to save local intersection: %v", err)
	}
	fmt.Printf("   ✅ Local intersection saved: %s\n", localIntersectionFile)
	fmt.Println()

	// STEP 6: Exchange intersection results for comparison
	fmt.Println("🔄 STEP 6: Exchanging Intersection Results")
	peerIntersection, err := exchangeIntersectionResults(conn, intersection, isServer)
	if err != nil {
		log.Fatalf("❌ Intersection exchange failed: %v", err)
	}
	fmt.Printf("   ✅ Received peer intersection (%d matches)\n", peerIntersection.Stats.MatchCount)
	fmt.Println()

	// STEP 7: Compare results and create diff if needed
	fmt.Println("⚖️  STEP 7: Comparing Intersection Results")
	resultsMatch, diffFile, err := compareIntersectionResults(intersection, peerIntersection)
	if err != nil {
		log.Fatalf("❌ Result comparison failed: %v", err)
	}

	if resultsMatch {
		fmt.Println("   ✅ SUCCESS: Intersection results match between peers!")
		fmt.Println("   🎉 Both peers computed identical intersections")

		// Copy results to output directory
		if err := copyToOutput(localIntersectionFile, "intersection_results.json"); err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to copy results to output: %v\n", err)
		} else {
			fmt.Printf("   📁 Results saved to: out/intersection_results.json\n")
		}
	} else {
		fmt.Println("   ❌ ERROR: Intersection results DO NOT match between peers!")
		fmt.Printf("   📋 Diff file created: %s\n", diffFile)

		// Copy diff to output directory
		if err := copyToOutput(diffFile, "intersection_diff.json"); err != nil {
			fmt.Printf("   ⚠️  Warning: Failed to copy diff to output: %v\n", err)
		} else {
			fmt.Printf("   📁 Diff saved to: out/intersection_diff.json\n")
		}

		log.Fatalf("❌ Workflow failed: Intersection results do not match")
	}

	fmt.Println()
	fmt.Println("🎉 UNIFIED PPRL WORKFLOW COMPLETED SUCCESSFULLY!")
	fmt.Println("==============================================")
	fmt.Printf("📁 Results available in: out/\n")
	if isDebugMode() {
		fmt.Printf("🐛 Debug files preserved in: %s/\n", tempDir)
	}
}

// performTokenizationStep handles tokenization if needed
func performTokenizationStep(cfg *config.Config) (string, error) {
	if cfg.Database.IsTokenized {
		fmt.Printf("   ✓ Using pre-tokenized data: %s\n", cfg.Database.Filename)
		return filepath.Join("..", cfg.Database.Filename), nil
	}

	fmt.Printf("   🔧 Tokenizing dataset: %s\n", cfg.Database.Filename)
	fmt.Printf("   📋 Fields: %s\n", strings.Join(cfg.Database.Fields, ", "))

	tokenizedFile := "tokenized_data.csv"

	// Use the same tokenization logic as the tokenize command
	tokenizeArgs := []string{
		"-input", filepath.Join("..", cfg.Database.Filename),
		"-output", tokenizedFile,
		"-main-config", filepath.Join("..", "config.yaml"),
		"-force",         // Skip confirmations
		"-no-encryption", // For simplicity in workflow mode
	}

	if err := runTokenizeCommandInternal(tokenizeArgs, cfg.Database.Fields); err != nil {
		return "", fmt.Errorf("tokenization failed: %v", err)
	}

	return tokenizedFile, nil
}

// establishPeerConnection creates a connection between peers
func establishPeerConnection(cfg *config.Config) (net.Conn, bool, error) {
	// First try to connect as client
	address := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)
	fmt.Printf("   🔄 Attempting to connect to peer at %s...\n", address)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err == nil {
		fmt.Printf("   ✅ Connected as client to %s\n", address)
		return conn, false, nil
	}

	fmt.Printf("   ⚠️  Client connection failed, starting server mode...\n")

	// If client connection fails, start as server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		return nil, false, fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("   🔄 Listening for peer connection on port %d...\n", cfg.ListenPort)

	// Accept one connection
	conn, err = listener.Accept()
	if err != nil {
		return nil, false, fmt.Errorf("failed to accept connection: %v", err)
	}

	fmt.Printf("   ✅ Peer connected from %s\n", conn.RemoteAddr())
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
		fmt.Printf("   🔄 Receiving tokens from peer...\n")
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

		fmt.Printf("   📨 Sending local tokens to peer...\n")
		if err := encoder.Encode(PeerMessage{Type: "tokens", Payload: localTokens}); err != nil {
			return nil, nil, fmt.Errorf("failed to send local tokens: %v", err)
		}

		return localTokens, peerTokens, nil
	} else {
		// Client: first send, then receive
		fmt.Printf("   📨 Sending local tokens to peer...\n")
		if err := encoder.Encode(PeerMessage{Type: "tokens", Payload: localTokens}); err != nil {
			return nil, nil, fmt.Errorf("failed to send local tokens: %v", err)
		}

		fmt.Printf("   🔄 Receiving tokens from peer...\n")
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

// computeIntersection computes the intersection using the same logic as validate.go
func computeIntersection(localTokens, peerTokens *TokenData, cfg *config.Config) (*IntersectionResult, error) {
	fmt.Printf("   🔧 Using Hamming threshold: %d\n", cfg.Matching.HammingThreshold)
	fmt.Printf("   📈 Using Jaccard threshold: %.3f\n", cfg.Matching.JaccardThreshold)

	// Convert TokenData to PatientRecords for matching
	localRecords, err := tokenDataToPatientRecords(localTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to convert local tokens: %v", err)
	}

	peerRecords, err := tokenDataToPatientRecords(peerTokens)
	if err != nil {
		return nil, fmt.Errorf("failed to convert peer tokens: %v", err)
	}

	// Perform matching using the same logic as validate.go
	matches, allComparisons, err := runWorkflowMatchingPipeline(
		localRecords,
		peerRecords,
		uint32(cfg.Matching.HammingThreshold),
		cfg.Matching.JaccardThreshold,
	)
	if err != nil {
		return nil, fmt.Errorf("matching pipeline failed: %v", err)
	}

	// Create intersection result
	result := &IntersectionResult{
		Matches: matches,
		Stats: IntersectionStats{
			TotalComparisons: len(allComparisons),
			MatchCount:       len(matches),
			MatchRate:        float64(len(matches)) / float64(len(allComparisons)),
			ComputedAt:       time.Now().Format(time.RFC3339),
		},
	}

	return result, nil
}

// tokenDataToPatientRecords converts TokenData to PatientRecords for matching
func tokenDataToPatientRecords(tokenData *TokenData) ([]server.PatientRecord, error) {
	var records []server.PatientRecord

	for _, tokenRecord := range tokenData.Records {
		// Decode Bloom filter from base64
		bf, err := pprl.BloomFromBase64(tokenRecord.BloomFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to decode bloom filter for %s: %v", tokenRecord.ID, err)
		}

		// Decode MinHash from base64
		mh, err := pprl.MinHashFromBase64(tokenRecord.MinHash)
		if err != nil {
			return nil, fmt.Errorf("failed to decode minhash for %s: %v", tokenRecord.ID, err)
		}

		record := server.PatientRecord{
			ID:          tokenRecord.ID,
			BloomFilter: bf,
			MinHash:     mh,
		}

		records = append(records, record)
	}

	return records, nil
}

// runWorkflowMatchingPipeline performs the intersection computation (copied from validate.go)
func runWorkflowMatchingPipeline(records1, records2 []server.PatientRecord, hammingThreshold uint32, jaccardThreshold float64) ([]*match.MatchResult, []*match.MatchResult, error) {
	fmt.Println("   🔄 Computing pairwise comparisons...")

	var allComparisons []*match.MatchResult
	var matches []*match.MatchResult

	totalComparisons := 0

	// Perform all pairwise comparisons
	for _, record1 := range records1 {
		for _, record2 := range records2 {
			totalComparisons++

			// Calculate Hamming distance
			hammingDist, err := record1.BloomFilter.HammingDistance(record2.BloomFilter)
			if err != nil {
				continue // Skip this comparison on error
			}

			// Calculate match score
			bfSize := record1.BloomFilter.GetSize()
			matchScore := 1.0
			if hammingDist > 0 {
				matchScore = 1.0 - (float64(hammingDist) / float64(bfSize))
			}

			// Calculate Jaccard similarity
			var jaccardSim float64
			if record1.MinHash != nil && record2.MinHash != nil {
				sig1, err1 := record1.MinHash.ComputeSignature(record1.BloomFilter)
				sig2, err2 := record2.MinHash.ComputeSignature(record2.BloomFilter)
				if err1 == nil && err2 == nil {
					jaccardSim, _ = pprl.JaccardSimilarity(sig1, sig2)
				}
			}

			// Determine if this is a match using BOTH thresholds
			isMatch := hammingDist <= hammingThreshold && jaccardSim >= jaccardThreshold

			// Create match result
			matchResult := &match.MatchResult{
				ID1:               record1.ID,
				ID2:               record2.ID,
				HammingDistance:   hammingDist,
				JaccardSimilarity: jaccardSim,
				MatchScore:        matchScore,
				IsMatch:           isMatch,
			}

			allComparisons = append(allComparisons, matchResult)

			// Add to matches if it meets threshold
			if matchResult.IsMatch {
				matches = append(matches, matchResult)
			}
		}
	}

	fmt.Printf("   ✅ Completed %d comparisons, found %d matches\n", len(allComparisons), len(matches))
	return matches, allComparisons, nil
}

// exchangeIntersectionResults exchanges intersection results between peers
func exchangeIntersectionResults(conn net.Conn, localIntersection *IntersectionResult, isServer bool) (*IntersectionResult, error) {
	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if isServer {
		// Server: first receive, then send
		fmt.Printf("   🔄 Receiving intersection from peer...\n")
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

		fmt.Printf("   📨 Sending local intersection to peer...\n")
		if err := encoder.Encode(PeerMessage{Type: "intersection", Payload: localIntersection}); err != nil {
			return nil, fmt.Errorf("failed to send local intersection: %v", err)
		}

		return peerIntersection, nil
	} else {
		// Client: first send, then receive
		fmt.Printf("   📨 Sending local intersection to peer...\n")
		if err := encoder.Encode(PeerMessage{Type: "intersection", Payload: localIntersection}); err != nil {
			return nil, fmt.Errorf("failed to send local intersection: %v", err)
		}

		fmt.Printf("   🔄 Receiving intersection from peer...\n")
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

// compareIntersectionResults compares two intersection results and creates a diff if they don't match
func compareIntersectionResults(local, peer *IntersectionResult) (bool, string, error) {
	// Compare basic stats
	if local.Stats.MatchCount != peer.Stats.MatchCount {
		fmt.Printf("   ❌ Match count differs: local=%d, peer=%d\n",
			local.Stats.MatchCount, peer.Stats.MatchCount)
	}

	if local.Stats.TotalComparisons != peer.Stats.TotalComparisons {
		fmt.Printf("   ❌ Total comparisons differ: local=%d, peer=%d\n",
			local.Stats.TotalComparisons, peer.Stats.TotalComparisons)
	}

	// Create sorted match sets for comparison
	localMatches := createMatchSet(local.Matches)
	peerMatches := createMatchSet(peer.Matches)

	// Find differences
	onlyInLocal := make(map[string]*match.MatchResult)
	onlyInPeer := make(map[string]*match.MatchResult)

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
	resultsMatch := len(onlyInLocal) == 0 && len(onlyInPeer) == 0 &&
		local.Stats.MatchCount == peer.Stats.MatchCount

	if resultsMatch {
		return true, "", nil
	}

	// Create diff file
	diffFile := "intersection_diff.json"
	diff := map[string]interface{}{
		"summary": map[string]interface{}{
			"matches":             resultsMatch,
			"local_match_count":   local.Stats.MatchCount,
			"peer_match_count":    peer.Stats.MatchCount,
			"local_comparisons":   local.Stats.TotalComparisons,
			"peer_comparisons":    peer.Stats.TotalComparisons,
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

// createMatchSet creates a map of matches keyed by a canonical string representation
func createMatchSet(matches []*match.MatchResult) map[string]*match.MatchResult {
	matchSet := make(map[string]*match.MatchResult)
	for _, match := range matches {
		if match.IsMatch {
			// Create canonical key (ensure consistent ordering)
			var key string
			if match.ID1 < match.ID2 {
				key = fmt.Sprintf("%s<->%s", match.ID1, match.ID2)
			} else {
				key = fmt.Sprintf("%s<->%s", match.ID2, match.ID1)
			}
			matchSet[key] = match
		}
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
			if value, exists := record[field]; exists && value != "" {
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
		fmt.Printf("🚀 %s (auto-confirmed with force flag)\n", message)
		return true
	}

	options := []string{
		"✅ Yes, continue",
		"❌ Cancel PPRL",
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
	fmt.Println("🔗 CohortBridge PPRL")
	fmt.Println("===================")
	fmt.Println("Peer-to-peer privacy-preserving record linkage")
	fmt.Println()

	fs := flag.NewFlagSet("pprl", flag.ExitOnError)
	var (
		configFile  = fs.String("config", "", "Configuration file")
		interactive = fs.Bool("interactive", false, "Force interactive mode")
		force       = fs.Bool("force", false, "Skip confirmation prompts and run automatically")
		help        = fs.Bool("help", false, "Show help message")
	)
	fs.Parse(args)

	if *help {
		showPPRLHelp()
		return
	}

	// Interactive mode if missing config or requested
	if *configFile == "" || *interactive {
		fmt.Println("🎯 Interactive PPRL Setup")
		fmt.Println("Let's configure your peer-to-peer record linkage...\n")

		if *configFile == "" {
			var err error
			*configFile, err = selectDataFile("Select Configuration File", "config", []string{".yaml"})
			if err != nil {
				fmt.Printf("❌ Error selecting config file: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Println()
	}

	// Show configuration summary
	fmt.Println("📋 PPRL Configuration:")
	fmt.Printf("  📁 Config File: %s\n", *configFile)
	fmt.Println()

	// Confirm before proceeding
	if !*force {
		confirmOptions := []string{
			"✅ Yes, start PPRL",
			"❌ Cancel",
		}

		confirmChoice := promptForChoice("Ready to start peer-to-peer record linkage?", confirmOptions)

		if confirmChoice == 1 {
			fmt.Println("\n👋 PPRL cancelled. Goodbye!")
			os.Exit(0)
		}
	} else {
		fmt.Println("🚀 Starting PPRL automatically (force mode)...")
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Validate config has required fields
	if cfg.Peer.Host == "" || cfg.Peer.Port == 0 {
		log.Fatalf("Configuration missing peer connection details (peer.host and peer.port)")
	}

	if cfg.ListenPort == 0 {
		log.Fatalf("Configuration missing listen_port")
	}

	if cfg.Matching.HammingThreshold == 0 {
		cfg.Matching.HammingThreshold = 20 // Default
	}

	if cfg.Matching.JaccardThreshold == 0 {
		cfg.Matching.JaccardThreshold = 0.5 // Default
	}

	// Run the PPRL workflow
	fmt.Println("🚀 Starting PPRL workflow...\n")
	runUnifiedWorkflow(cfg, *force)
}

func showPPRLHelp() {
	fmt.Println("🔗 CohortBridge PPRL")
	fmt.Println("===================")
	fmt.Println()
	fmt.Println("PPRL STEPS:")
	fmt.Println("  1. 📋 Read configuration file")
	fmt.Println("  2. 🔧 Tokenize dataset (if not pre-tokenized)")
	fmt.Println("  3. 📡 Establish peer connection")
	fmt.Println("  4. 🔄 Exchange tokens with peer")
	fmt.Println("  5. 🔍 Compute intersection using thresholds")
	fmt.Println("  6. 🔄 Exchange intersection results")
	fmt.Println("  7. ⚖️  Compare results and create diff if needed")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge pprl [OPTIONS]")
	fmt.Println("  cohort-bridge pprl                       # Interactive mode")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -config string     Configuration file")
	fmt.Println("  -interactive       Force interactive mode")
	fmt.Println("  -force             Skip confirmation prompts")
	fmt.Println("  -help              Show this help message")
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
	fmt.Println("CONFIGURATION REQUIREMENTS:")
	fmt.Println("  - peer.host and peer.port (peer connection)")
	fmt.Println("  - listen_port (local server port)")
	fmt.Println("  - matching.hamming_threshold (default: 20)")
	fmt.Println("  - matching.jaccard_threshold (default: 0.5)")
}
