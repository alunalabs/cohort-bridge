package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// RunAsSender implements the sender mode with fuzzy matching
func RunAsSender(cfg *config.Config) {
	fmt.Printf("🚀 Starting sender mode...\n")

	// Load CSV data
	csvDB, err := db.NewCSVDatabase(cfg.Database.Filename)
	if err != nil {
		log.Fatalf("Failed to load CSV database: %v", err)
	}

	// Convert CSV records to Bloom filters using the utility function
	records, err := LoadPatientRecordsUtil(csvDB, cfg.Database.Fields)
	if err != nil {
		log.Fatalf("Failed to load patient records: %v", err)
	}

	fmt.Printf("📊 Loaded %d patient records\n", len(records))

	// Connect to receiver
	address := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)
	fmt.Printf("🔗 Connecting to receiver at %s...\n", address)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to receiver: %v", err)
	}
	defer conn.Close()

	fmt.Println("✅ Connected to receiver")

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	// Step 1: Send blocking request
	fmt.Println("🔐 Initiating secure blocking...")

	blockingData := BlockingData{
		EncryptedBuckets: make(map[string][]string),
		Signatures:       make(map[string]string),
	}

	// Populate blocking data (simplified bucketing)
	for _, record := range records {
		bucket := fmt.Sprintf("bucket_%s", record.ID[:1]) // Simple bucketing by first character
		blockingData.EncryptedBuckets[bucket] = append(blockingData.EncryptedBuckets[bucket], record.ID)
	}

	blockingRequest := MatchingMessage{
		Type:    "blocking_request",
		Payload: blockingData,
	}

	if err := encoder.Encode(blockingRequest); err != nil {
		log.Fatalf("Failed to send blocking request: %v", err)
	}

	// Receive blocking response
	var blockingResponse MatchingMessage
	if err := decoder.Decode(&blockingResponse); err != nil {
		log.Fatalf("Failed to receive blocking response: %v", err)
	}

	if blockingResponse.Type != "blocking_response" {
		log.Fatalf("Unexpected response type: %s", blockingResponse.Type)
	}

	fmt.Println("✅ Blocking phase complete")

	// Step 2: Send matching request
	fmt.Println("🔍 Initiating fuzzy matching...")

	matchingData := MatchingData{
		Records: make(map[string]string),
	}

	// Prepare our Bloom filter data
	for _, record := range records {
		bloomData, err := pprl.BloomToBase64(record.BloomFilter)
		if err != nil {
			log.Printf("Failed to encode Bloom filter for record %s: %v", record.ID, err)
			continue
		}
		matchingData.Records[record.ID] = bloomData
	}

	matchingRequest := MatchingMessage{
		Type:    "matching_request",
		Payload: matchingData,
	}

	if err := encoder.Encode(matchingRequest); err != nil {
		log.Fatalf("Failed to send matching request: %v", err)
	}

	// Receive matching response
	var matchingResponse MatchingMessage
	if err := decoder.Decode(&matchingResponse); err != nil {
		log.Fatalf("Failed to receive matching response: %v", err)
	}

	if matchingResponse.Type != "matching_response" {
		log.Fatalf("Unexpected response type: %s", matchingResponse.Type)
	}

	fmt.Println("✅ Matching phase complete")

	// Receive final results
	var resultsMessage MatchingMessage
	if err := decoder.Decode(&resultsMessage); err != nil {
		log.Fatalf("Failed to receive results: %v", err)
	}

	if resultsMessage.Type != "results" {
		log.Fatalf("Unexpected response type: %s", resultsMessage.Type)
	}

	// Process and display results
	fmt.Println("\n🎯 Matching Results:")
	fmt.Println("==================")

	// Convert payload to results
	payloadBytes, _ := json.Marshal(resultsMessage.Payload)
	var results match.TwoPartyMatchResult
	if err := json.Unmarshal(payloadBytes, &results); err != nil {
		log.Printf("Failed to parse results: %v", err)
		return
	}

	fmt.Printf("📈 Statistics:\n")
	fmt.Printf("   Records processed: %d\n", len(records))
	fmt.Printf("   Matching buckets: %d\n", results.MatchingBuckets)
	fmt.Printf("   Candidate pairs: %d\n", results.CandidatePairs)
	fmt.Printf("   Matches found: %d\n", results.TotalMatches)
	fmt.Printf("   Party 1 records: %d\n", results.Party1Records)
	fmt.Printf("   Party 2 records: %d\n", results.Party2Records)

	if len(results.Matches) > 0 {
		fmt.Printf("\n📋 Detailed Matches:\n")
		for i, match := range results.Matches {
			fmt.Printf("%3d. %s <-> %s (Score: %.3f)\n",
				i+1, match.ID1, match.ID2, match.MatchScore)
		}
	} else {
		fmt.Println("   No matches found")
	}

	fmt.Println("\n✅ Fuzzy matching session complete!")
}

// SendShutdown sends a shutdown signal to the receiver
func SendShutdown(cfg *config.Config) {
	fmt.Printf("🔴 Sending shutdown signal to receiver...\n")

	// Connect to receiver
	address := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)
	fmt.Printf("🔗 Connecting to receiver at %s...\n", address)

	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to receiver: %v", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)

	// Send shutdown message
	shutdownMsg := MatchingMessage{
		Type:    "shutdown",
		Payload: nil,
	}

	if err := encoder.Encode(shutdownMsg); err != nil {
		log.Fatalf("Failed to send shutdown message: %v", err)
	}

	fmt.Println("✅ Shutdown signal sent successfully!")
}
