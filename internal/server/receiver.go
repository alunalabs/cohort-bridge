package server

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// PatientRecord represents a patient with Bloom filter representation
type PatientRecord struct {
	ID               string
	BloomFilter      *pprl.BloomFilter
	MinHash          *pprl.MinHash
	MinHashSignature []uint32 // Pre-computed MinHash signature
}

// MatchingMessage represents the protocol messages
type MatchingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// BlockingData represents encrypted bucket data for blocking
type BlockingData struct {
	EncryptedBuckets map[string][]string `json:"encrypted_buckets"`
	Signatures       map[string]string   `json:"signatures"`
}

// MatchingData represents Bloom filter data for fuzzy matching
type MatchingData struct {
	Records map[string]string `json:"records"` // ID -> base64 encoded Bloom filter
}

// TimeoutConn wraps a net.Conn with configurable timeouts
type TimeoutConn struct {
	conn         net.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration
}

func NewTimeoutConn(conn net.Conn, cfg *config.Config) *TimeoutConn {
	return &TimeoutConn{
		conn:         conn,
		readTimeout:  cfg.Timeouts.ReadTimeout,
		writeTimeout: cfg.Timeouts.WriteTimeout,
		idleTimeout:  cfg.Timeouts.IdleTimeout,
	}
}

func (tc *TimeoutConn) Read(b []byte) (n int, err error) {
	if tc.readTimeout > 0 {
		tc.conn.SetReadDeadline(time.Now().Add(tc.readTimeout))
	}
	return tc.conn.Read(b)
}

func (tc *TimeoutConn) Write(b []byte) (n int, err error) {
	if tc.writeTimeout > 0 {
		tc.conn.SetWriteDeadline(time.Now().Add(tc.writeTimeout))
	}
	return tc.conn.Write(b)
}

func (tc *TimeoutConn) Close() error {
	return tc.conn.Close()
}

func (tc *TimeoutConn) RemoteAddr() net.Addr {
	return tc.conn.RemoteAddr()
}

func (tc *TimeoutConn) SetDeadline(t time.Time) error {
	return tc.conn.SetDeadline(t)
}

// RunAsReceiver implements the receiver mode with enhanced security and logging
func RunAsReceiver(cfg *config.Config) {
	sessionID := fmt.Sprintf("recv-%d", time.Now().Unix())

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

	Info("Starting receiver mode with session ID: %s", sessionID)

	// Initialize security manager
	securityManager := NewSecurityManager(cfg)
	Info("Security manager initialized with %d allowed IPs", len(cfg.Security.AllowedIPs))

	// Load patient records based on configuration
	var records []PatientRecord
	var csvDB *db.CSVDatabase
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

		// Note: csvDB will be nil for tokenized data, which is fine since we don't need raw PHI
	} else {
		// Load raw PHI data and convert to Bloom filters
		Info("Loading CSV database from: %s", cfg.Database.Filename)
		csvDB, err = db.NewCSVDatabase(cfg.Database.Filename)
		if err != nil {
			Error("Failed to load CSV database: %v", err)
			return
		}

		// Convert CSV records to Bloom filters
		randomBitsPercent := cfg.Database.RandomBitsPercent
		if randomBitsPercent > 0.0 {
			Info("Using %.1f%% random bits in Bloom filters", randomBitsPercent*100)
		}

		Info("Converting CSV records to Bloom filters...")
		records, err = LoadPatientRecordsUtilWithRandomBits(csvDB, cfg.Database.Fields, randomBitsPercent)
		if err != nil {
			Error("Failed to load patient records: %v", err)
			return
		}
		Info("Successfully loaded %d patient records", len(records))
	}

	// Start TCP server with timeouts
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		Error("Failed to start listener: %v", err)
		return
	}
	defer listener.Close()

	Info("Listening for connections on port %d", cfg.ListenPort)
	Info("Security settings: IP check=%v, Rate limit=%d/min, Max connections=%d",
		cfg.Security.RequireIPCheck, cfg.Security.RateLimitPerMin, cfg.Security.MaxConnections)

	// Set listener deadline for clean shutdown capability
	for {
		// Set accept timeout
		if err := listener.(*net.TCPListener).SetDeadline(time.Now().Add(cfg.Timeouts.ConnectionTimeout)); err != nil {
			Error("Failed to set listener deadline: %v", err)
			break
		}

		conn, err := listener.Accept()
		if err != nil {
			// Check if this is a timeout (expected) or real error
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				Debug("Accept timeout, checking for shutdown signal...")
				continue
			}
			Error("Failed to accept connection: %v", err)
			break
		}

		// Handle connection in goroutine for concurrent processing
		go handleSecureConnection(conn, records, csvDB, cfg, securityManager, sessionID)
	}

	Info("Receiver shutting down")
}

// handleSecureConnection handles a single client connection with full security and logging
func handleSecureConnection(conn net.Conn, records []PatientRecord, csvDB *db.CSVDatabase,
	cfg *config.Config, securityManager *SecurityManager, sessionID string) {

	remoteAddr := conn.RemoteAddr().String()
	connID := fmt.Sprintf("%s-%d", sessionID, time.Now().UnixNano())

	Info("New connection attempt from %s (Connection ID: %s)", remoteAddr, connID)

	// Apply security validation
	if err := securityManager.ValidateConnection(remoteAddr); err != nil {
		Error("Connection rejected from %s: %v", remoteAddr, err)
		conn.Close()
		return
	}

	// Record successful connection
	securityManager.RecordConnection(remoteAddr)
	defer securityManager.RecordDisconnection(remoteAddr)

	// Wrap connection with timeouts
	timeoutConn := NewTimeoutConn(conn, cfg)
	defer timeoutConn.Close()

	Info("Connection accepted from %s (ID: %s)", remoteAddr, connID)
	Audit("CONNECTION_ACCEPTED", map[string]interface{}{
		"remote_addr":   remoteAddr,
		"connection_id": connID,
		"session_id":    sessionID,
		"stats":         securityManager.GetConnectionStats(),
	})

	// Set initial handshake timeout
	if err := timeoutConn.SetDeadline(time.Now().Add(cfg.Timeouts.HandshakeTimeout)); err != nil {
		Error("Failed to set handshake deadline: %v", err)
		return
	}

	// Handle the protocol
	if err := handleMatchingProtocol(timeoutConn, records, csvDB, cfg, connID); err != nil {
		Error("Protocol error for connection %s: %v", connID, err)
		Audit("PROTOCOL_ERROR", map[string]interface{}{
			"connection_id": connID,
			"remote_addr":   remoteAddr,
			"error":         err.Error(),
		})
		return
	}

	Info("Successfully completed matching session for connection %s", connID)
	Audit("SESSION_COMPLETED", map[string]interface{}{
		"connection_id": connID,
		"remote_addr":   remoteAddr,
		"session_id":    sessionID,
	})
}

// handleMatchingProtocol implements the secure matching protocol with timeouts
func handleMatchingProtocol(conn *TimeoutConn, records []PatientRecord, csvDB *db.CSVDatabase,
	cfg *config.Config, connID string) error {

	Debug("Starting matching protocol for connection %s", connID)

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	// Convert our records to pprl format
	var pprlRecords []*pprl.Record
	for _, record := range records {
		bloomData, err := pprl.BloomToBase64(record.BloomFilter)
		if err != nil {
			Warn("Failed to encode Bloom filter for record %s: %v", record.ID, err)
			continue
		}

		sig, err := record.MinHash.ComputeSignature(record.BloomFilter)
		if err != nil {
			Warn("Failed to compute MinHash signature for record %s: %v", record.ID, err)
			continue
		}

		pprlRecords = append(pprlRecords, &pprl.Record{
			ID:        record.ID,
			BloomData: bloomData,
			MinHash:   sig,
		})
	}

	Debug("Converted %d records to PPRL format", len(pprlRecords))

	// Protocol handling loop with proper timeouts
	for {
		// Reset deadline for each message
		if err := conn.SetDeadline(time.Now().Add(cfg.Timeouts.ReadTimeout)); err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}

		var msg MatchingMessage
		if err := decoder.Decode(&msg); err != nil {
			if err.Error() == "EOF" {
				Debug("Client closed connection for %s", connID)
				return nil
			}
			return fmt.Errorf("failed to decode message: %w", err)
		}

		Debug("Received message type: %s for connection %s", msg.Type, connID)

		switch msg.Type {
		case "shutdown":
			Info("Received shutdown signal from %s", connID)
			return nil

		case "blocking_request":
			if err := handleBlockingRequest(encoder, records, connID); err != nil {
				return fmt.Errorf("blocking request failed: %w", err)
			}

		case "matching_request":
			if err := handleMatchingRequest(encoder, msg, pprlRecords, csvDB, cfg, connID); err != nil {
				return fmt.Errorf("matching request failed: %w", err)
			}
			// After matching, we can exit the loop
			return nil

		default:
			Warn("Unknown message type '%s' from connection %s", msg.Type, connID)
			return fmt.Errorf("unknown message type: %s", msg.Type)
		}
	}
}

// handleBlockingRequest processes the blocking phase
func handleBlockingRequest(encoder *json.Encoder, records []PatientRecord, connID string) error {
	Info("Processing blocking request for connection %s", connID)

	// Send back our blocking data
	ourBlocking := BlockingData{
		EncryptedBuckets: make(map[string][]string),
		Signatures:       make(map[string]string),
	}

	// Simple bucketing for demo - in production this would use secure LSH
	bucketCount := 0
	for _, record := range records {
		bucket := fmt.Sprintf("bucket_%s", record.ID[:1])
		ourBlocking.EncryptedBuckets[bucket] = append(ourBlocking.EncryptedBuckets[bucket], record.ID)
		bucketCount++
	}

	response := MatchingMessage{
		Type:    "blocking_response",
		Payload: ourBlocking,
	}

	if err := encoder.Encode(response); err != nil {
		return fmt.Errorf("failed to send blocking response: %w", err)
	}

	Info("Sent blocking response with %d buckets for connection %s",
		len(ourBlocking.EncryptedBuckets), connID)
	return nil
}

// handleMatchingRequest processes the matching phase
func handleMatchingRequest(encoder *json.Encoder, msg MatchingMessage, pprlRecords []*pprl.Record,
	csvDB *db.CSVDatabase, cfg *config.Config, connID string) error {

	Info("Processing matching request for connection %s", connID)

	// Parse sender's matching data
	payloadBytes, _ := json.Marshal(msg.Payload)
	var senderMatching MatchingData
	if err := json.Unmarshal(payloadBytes, &senderMatching); err != nil {
		return fmt.Errorf("failed to parse sender matching data: %w", err)
	}

	Info("Comparing %d receiver records with %d sender records for connection %s",
		len(pprlRecords), len(senderMatching.Records), connID)

	// Perform fuzzy matching
	startTime := time.Now()
	matchResults := performFuzzyMatching(pprlRecords, senderMatching, connID)
	matchDuration := time.Since(startTime)

	Info("Fuzzy matching completed in %v, found %d matches for connection %s",
		matchDuration, len(matchResults), connID)

	// Send matching response
	matchResponse := MatchingMessage{
		Type:    "matching_response",
		Payload: map[string]interface{}{"status": "complete"},
	}

	if err := encoder.Encode(matchResponse); err != nil {
		return fmt.Errorf("failed to send matching response: %w", err)
	}

	// Prepare and send results
	results := match.TwoPartyMatchResult{
		Matches:         matchResults,
		TotalMatches:    len(matchResults),
		Party1Records:   len(pprlRecords),
		Party2Records:   len(senderMatching.Records),
		CandidatePairs:  len(pprlRecords) * len(senderMatching.Records),
		MatchingBuckets: 1, // Simplified for demo
	}

	resultsMessage := MatchingMessage{
		Type:    "results",
		Payload: results,
	}

	if err := encoder.Encode(resultsMessage); err != nil {
		return fmt.Errorf("failed to send results: %w", err)
	}

	// Save results to files
	timestamp := time.Now().Format("20060102_150405")
	if err := saveResults(matchResults, csvDB, cfg.Database.Fields, timestamp, connID); err != nil {
		Warn("Failed to save results for connection %s: %v", connID, err)
	}

	Info("Results sent and saved for connection %s", connID)
	return nil
}

// performFuzzyMatching implements the core matching algorithm
func performFuzzyMatching(pprlRecords []*pprl.Record, senderMatching MatchingData, connID string) []*match.MatchResult {
	var matchResults []*match.MatchResult

	fuzzyConfig := &match.FuzzyMatchConfig{
		HammingThreshold:  200, // Allow up to 200 bit differences
		JaccardThreshold:  0.5, // Require at least 50% Jaccard similarity
		UseSecureProtocol: false,
	}

	totalComparisons := 0
	matchesFound := 0

	for _, receiverRecord := range pprlRecords {
		for senderID, senderBloomData := range senderMatching.Records {
			totalComparisons++

			// Decode Bloom filters for comparison
			receiverBF, err := pprl.BloomFromBase64(receiverRecord.BloomData)
			if err != nil {
				Debug("Failed to decode receiver Bloom filter: %v", err)
				continue
			}

			senderBF, err := pprl.BloomFromBase64(senderBloomData)
			if err != nil {
				Debug("Failed to decode sender Bloom filter: %v", err)
				continue
			}

			// Calculate Hamming distance
			hammingDist, err := receiverBF.HammingDistance(senderBF)
			if err != nil {
				Debug("Failed to calculate Hamming distance: %v", err)
				continue
			}

			// Calculate Jaccard similarity for scoring (simple approximation)
			// For Bloom filters, we'll use a simple approximation based on Hamming distance
			maxBits := float64(receiverBF.GetSize())
			jaccard := 1.0 - (float64(hammingDist) / maxBits)

			// Determine if this is a match
			isMatch := hammingDist <= fuzzyConfig.HammingThreshold

			// Create match result
			matchResult := &match.MatchResult{
				ID1:             receiverRecord.ID,
				ID2:             senderID,
				MatchScore:      jaccard,
				HammingDistance: hammingDist,
				IsMatch:         isMatch,
			}

			matchResults = append(matchResults, matchResult)

			if isMatch {
				matchesFound++
				Debug("Match found: %s <-> %s (Hamming: %d, Jaccard: %.3f)",
					receiverRecord.ID, senderID, hammingDist, jaccard)
			}
		}
	}

	Info("Completed %d comparisons, found %d matches for connection %s",
		totalComparisons, matchesFound, connID)

	return matchResults
}

// saveResults saves matching results to CSV files
func saveResults(matches []*match.MatchResult, csvDB *db.CSVDatabase, fields []string, timestamp, connID string) error {
	// Save basic results
	basicFilename := fmt.Sprintf("out/matches_%s_%s.csv", timestamp, connID)
	saveResultsToCSV(matches, basicFilename)

	// Save detailed results with patient data (only if csvDB is available)
	if csvDB != nil {
		detailFilename := fmt.Sprintf("out/match_details_%s_%s.csv", timestamp, connID)
		saveMatchDetailsToCSV(matches, detailFilename, csvDB, fields)
		Info("Results saved to %s and %s", basicFilename, detailFilename)
	} else {
		Info("Results saved to %s (detailed results not available for tokenized data)", basicFilename)
	}

	return nil
}

// saveResultsToCSV saves basic match results
func saveResultsToCSV(matches []*match.MatchResult, filename string) {
	// Ensure output directory exists
	if err := EnsureOutputDirectory(); err != nil {
		Error("Failed to ensure output directory: %v", err)
		return
	}

	file, err := os.Create(filename)
	if err != nil {
		Error("Failed to create results file %s: %v", filename, err)
		return
	}
	defer file.Close()

	// Write CSV header
	file.WriteString("Receiver_ID,Sender_ID,Match_Score,Hamming_Distance,Is_Match\n")

	// Write match results
	for _, match := range matches {
		file.WriteString(fmt.Sprintf("%s,%s,%.3f,%d,%t\n",
			match.ID1, match.ID2, match.MatchScore, match.HammingDistance, match.IsMatch))
	}

	Info("Basic results saved to: %s with %d records", filename, len(matches))
}

// saveMatchDetailsToCSV saves detailed match results with patient demographics
func saveMatchDetailsToCSV(matches []*match.MatchResult, filename string, csvDB *db.CSVDatabase, fields []string) {
	file, err := os.Create(filename)
	if err != nil {
		Error("Failed to create match details file %s: %v", filename, err)
		return
	}
	defer file.Close()

	// Build dynamic header based on configured fields
	header := "Receiver_ID"
	for _, field := range fields {
		header += fmt.Sprintf(",Receiver_%s", strings.Title(field))
	}
	header += ",Sender_ID,Match_Score,Hamming_Distance,Is_Match\n"

	// Write CSV header
	file.WriteString(header)

	// Get all receiver records from CSV
	allReceiverRecords, err := csvDB.List(0, 1000000)
	if err != nil {
		Error("Failed to get receiver records: %v", err)
		return
	}

	// Create a map for quick lookup
	receiverMap := make(map[string]map[string]string)
	for _, record := range allReceiverRecords {
		receiverMap[record["id"]] = record
	}

	// Write match details with patient demographics
	for _, match := range matches {
		receiverRecord, receiverExists := receiverMap[match.ID1]
		if receiverExists {
			// Build row data based on configured fields
			row := match.ID1
			for _, field := range fields {
				if value, exists := receiverRecord[field]; exists {
					row += fmt.Sprintf(",%s", value)
				} else {
					row += ","
				}
			}
			row += fmt.Sprintf(",%s,%.3f,%d,%t\n",
				match.ID2, match.MatchScore, match.HammingDistance, match.IsMatch)

			file.WriteString(row)
		} else {
			// Build empty row for missing receiver record
			row := match.ID1
			for range fields {
				row += ","
			}
			row += fmt.Sprintf(",%s,%.3f,%d,%t\n",
				match.ID2, match.MatchScore, match.HammingDistance, match.IsMatch)

			file.WriteString(row)
		}
	}

	Info("Match details saved to: %s", filename)
}
