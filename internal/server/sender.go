package server

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// RunAsSender implements the sender mode with enhanced logging and timeout handling
func RunAsSender(cfg *config.Config) {
	sessionID := fmt.Sprintf("send-%d", time.Now().Unix())

	// Ensure required directories exist
	if err := EnsureOutputDirectory(); err != nil {
		fmt.Printf("Warning: Failed to create output directory: %v\n", err)
	}
	if err := EnsureLogsDirectory(); err != nil {
		fmt.Printf("Warning: Failed to create logs directory: %v\n", err)
	}

	// Initialize logging
	if err := InitLogger(cfg, sessionID); err != nil {
		fmt.Printf("Warning: Failed to initialize logger: %v\n", err)
	}
	defer GetLogger().Close()

	Info("Starting sender mode with session ID: %s", sessionID)

	// Load patient records based on configuration
	var records []PatientRecord
	var err error

	if cfg.Database.IsTokenized {
		// Load tokenized data
		Info("Loading tokenized data from: %s", cfg.Database.TokenizedFile)
		records, err = LoadTokenizedRecords(cfg.Database.TokenizedFile)
		if err != nil {
			Error("Failed to load tokenized records: %v", err)
			return
		}
		Info("Successfully loaded %d tokenized records", len(records))
	} else {
		// Load raw PHI data and convert to Bloom filters
		Info("Loading CSV database from: %s", cfg.Database.Filename)
		csvDB, err := db.NewCSVDatabase(cfg.Database.Filename)
		if err != nil {
			Error("Failed to load CSV database: %v", err)
			return
		}

		// Convert CSV records to Bloom filters using the utility function
		randomBitsPercent := cfg.Database.RandomBitsPercent
		if randomBitsPercent > 0.0 {
			Info("Using %.1f%% random bits in Bloom filters", randomBitsPercent*100)
		}

		Info("Converting CSV records to Bloom filters...")
		startTime := time.Now()
		records, err = LoadPatientRecordsUtilWithRandomBits(csvDB, cfg.Database.Fields, randomBitsPercent)
		if err != nil {
			Error("Failed to load patient records: %v", err)
			return
		}
		loadDuration := time.Since(startTime)
		Info("Successfully loaded %d patient records in %v", len(records), loadDuration)
	}

	// Connect to receiver with timeout
	address := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)
	Info("Attempting to connect to receiver at %s...", address)

	conn, err := net.DialTimeout("tcp", address, cfg.Timeouts.ConnectionTimeout)
	if err != nil {
		Error("Failed to connect to receiver at %s: %v", address, err)
		Audit("CONNECTION_FAILED", map[string]interface{}{
			"target_address": address,
			"error":          err.Error(),
			"session_id":     sessionID,
		})
		return
	}
	defer conn.Close()

	Info("Successfully connected to receiver at %s", address)
	Audit("CONNECTION_ESTABLISHED", map[string]interface{}{
		"target_address": address,
		"session_id":     sessionID,
	})

	// Wrap connection with timeouts
	timeoutConn := NewTimeoutConn(conn, cfg)
	defer timeoutConn.Close()

	// Set initial handshake timeout
	if err := timeoutConn.SetDeadline(time.Now().Add(cfg.Timeouts.HandshakeTimeout)); err != nil {
		Error("Failed to set handshake deadline: %v", err)
		return
	}

	// Execute the matching protocol
	if err := executeSenderProtocol(timeoutConn, records, cfg, sessionID); err != nil {
		Error("Protocol execution failed: %v", err)
		Audit("PROTOCOL_ERROR", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	Info("Fuzzy matching session completed successfully")
	Audit("SESSION_COMPLETED", map[string]interface{}{
		"session_id": sessionID,
	})
}

// executeSenderProtocol implements the sender side of the matching protocol
func executeSenderProtocol(conn *TimeoutConn, records []PatientRecord, cfg *config.Config, sessionID string) error {
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	Info("Starting protocol execution for session %s", sessionID)

	// Step 1: Send blocking request
	Debug("Initiating secure blocking phase...")
	if err := sendBlockingRequest(encoder, records, sessionID); err != nil {
		return fmt.Errorf("blocking request failed: %w", err)
	}

	// Receive blocking response with timeout
	if err := conn.SetDeadline(time.Now().Add(cfg.Timeouts.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	var blockingResponse MatchingMessage
	if err := decoder.Decode(&blockingResponse); err != nil {
		return fmt.Errorf("failed to receive blocking response: %w", err)
	}

	if blockingResponse.Type != "blocking_response" {
		return fmt.Errorf("unexpected response type: %s", blockingResponse.Type)
	}

	Info("Blocking phase completed successfully")

	// Step 2: Send matching request
	Debug("Initiating fuzzy matching phase...")
	if err := sendMatchingRequest(encoder, records, sessionID); err != nil {
		return fmt.Errorf("matching request failed: %w", err)
	}

	// Receive matching response with timeout
	if err := conn.SetDeadline(time.Now().Add(cfg.Timeouts.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	var matchingResponse MatchingMessage
	if err := decoder.Decode(&matchingResponse); err != nil {
		return fmt.Errorf("failed to receive matching response: %w", err)
	}

	if matchingResponse.Type != "matching_response" {
		return fmt.Errorf("unexpected response type: %s", matchingResponse.Type)
	}

	Info("Matching phase completed successfully")

	// Step 3: Receive final results
	if err := conn.SetDeadline(time.Now().Add(cfg.Timeouts.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	var resultsMessage MatchingMessage
	if err := decoder.Decode(&resultsMessage); err != nil {
		return fmt.Errorf("failed to receive results: %w", err)
	}

	if resultsMessage.Type != "results" {
		return fmt.Errorf("unexpected response type: %s", resultsMessage.Type)
	}

	// Process and display results
	return processResults(resultsMessage, len(records), sessionID)
}

// sendBlockingRequest sends the blocking phase request
func sendBlockingRequest(encoder *json.Encoder, records []PatientRecord, sessionID string) error {
	Info("Sending blocking request for session %s", sessionID)

	blockingData := BlockingData{
		EncryptedBuckets: make(map[string][]string),
		Signatures:       make(map[string]string),
	}

	// Populate blocking data (simplified bucketing)
	bucketCount := 0
	for _, record := range records {
		bucket := fmt.Sprintf("bucket_%s", record.ID[:1]) // Simple bucketing by first character
		blockingData.EncryptedBuckets[bucket] = append(blockingData.EncryptedBuckets[bucket], record.ID)
		bucketCount++
	}

	blockingRequest := MatchingMessage{
		Type:    "blocking_request",
		Payload: blockingData,
	}

	if err := encoder.Encode(blockingRequest); err != nil {
		return fmt.Errorf("failed to send blocking request: %w", err)
	}

	Info("Blocking request sent with %d buckets for session %s",
		len(blockingData.EncryptedBuckets), sessionID)
	return nil
}

// sendMatchingRequest sends the matching phase request
func sendMatchingRequest(encoder *json.Encoder, records []PatientRecord, sessionID string) error {
	Info("Sending matching request for session %s", sessionID)

	matchingData := MatchingData{
		Records: make(map[string]string),
	}

	// Prepare our Bloom filter data
	successfulEncodes := 0
	for _, record := range records {
		bloomData, err := pprl.BloomToBase64(record.BloomFilter)
		if err != nil {
			Warn("Failed to encode Bloom filter for record %s: %v", record.ID, err)
			continue
		}
		matchingData.Records[record.ID] = bloomData
		successfulEncodes++
	}

	matchingRequest := MatchingMessage{
		Type:    "matching_request",
		Payload: matchingData,
	}

	if err := encoder.Encode(matchingRequest); err != nil {
		return fmt.Errorf("failed to send matching request: %w", err)
	}

	Info("Matching request sent with %d records for session %s", successfulEncodes, sessionID)
	return nil
}

// processResults processes and displays the matching results
func processResults(resultsMessage MatchingMessage, recordCount int, sessionID string) error {
	Info("Processing results for session %s", sessionID)

	// Convert payload to results
	payloadBytes, _ := json.Marshal(resultsMessage.Payload)
	var results match.TwoPartyMatchResult
	if err := json.Unmarshal(payloadBytes, &results); err != nil {
		return fmt.Errorf("failed to parse results: %w", err)
	}

	// Display results
	fmt.Println("\n🎯 Matching Results:")
	fmt.Println("==================")

	fmt.Printf("📈 Statistics:\n")
	fmt.Printf("   Records processed: %d\n", recordCount)
	fmt.Printf("   Matching buckets: %d\n", results.MatchingBuckets)
	fmt.Printf("   Candidate pairs: %d\n", results.CandidatePairs)
	fmt.Printf("   Matches found: %d\n", results.TotalMatches)
	fmt.Printf("   Party 1 records: %d\n", results.Party1Records)
	fmt.Printf("   Party 2 records: %d\n", results.Party2Records)

	matchesFound := 0
	for _, match := range results.Matches {
		if match.IsMatch {
			matchesFound++
		}
	}

	if matchesFound > 0 {
		fmt.Printf("\n📋 Detailed Matches:\n")
		count := 0
		for _, match := range results.Matches {
			if match.IsMatch {
				count++
				fmt.Printf("%3d. %s <-> %s (Score: %.3f, Hamming: %d)\n",
					count, match.ID1, match.ID2, match.MatchScore, match.HammingDistance)
			}
		}
	} else {
		fmt.Println("   No matches found")
	}

	Info("Results processing completed: %d total matches, %d true matches",
		results.TotalMatches, matchesFound)

	Audit("RESULTS_PROCESSED", map[string]interface{}{
		"session_id":      sessionID,
		"total_matches":   results.TotalMatches,
		"true_matches":    matchesFound,
		"records_sent":    recordCount,
		"candidate_pairs": results.CandidatePairs,
	})

	return nil
}

// SendShutdown sends a shutdown signal to the receiver with enhanced logging
func SendShutdown(cfg *config.Config) {
	sessionID := fmt.Sprintf("shutdown-%d", time.Now().Unix())

	// Initialize logging
	if err := InitLogger(cfg, sessionID); err != nil {
		fmt.Printf("Warning: Failed to initialize logger: %v\n", err)
	}
	defer GetLogger().Close()

	Info("Sending shutdown signal to receiver...")

	// Connect to receiver with timeout
	address := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)
	Info("Connecting to receiver at %s...", address)

	conn, err := net.DialTimeout("tcp", address, cfg.Timeouts.ConnectionTimeout)
	if err != nil {
		Error("Failed to connect to receiver: %v", err)
		Audit("SHUTDOWN_FAILED", map[string]interface{}{
			"target_address": address,
			"error":          err.Error(),
		})
		return
	}
	defer conn.Close()

	// Wrap with timeout
	timeoutConn := NewTimeoutConn(conn, cfg)
	defer timeoutConn.Close()

	// Set write timeout
	if err := timeoutConn.SetDeadline(time.Now().Add(cfg.Timeouts.WriteTimeout)); err != nil {
		Error("Failed to set write deadline: %v", err)
		return
	}

	encoder := json.NewEncoder(timeoutConn)

	// Send shutdown message
	shutdownMsg := MatchingMessage{
		Type:    "shutdown",
		Payload: nil,
	}

	if err := encoder.Encode(shutdownMsg); err != nil {
		Error("Failed to send shutdown message: %v", err)
		Audit("SHUTDOWN_FAILED", map[string]interface{}{
			"target_address": address,
			"error":          err.Error(),
		})
		return
	}

	Info("Shutdown signal sent successfully!")
	Audit("SHUTDOWN_SENT", map[string]interface{}{
		"target_address": address,
		"session_id":     sessionID,
	})
}
