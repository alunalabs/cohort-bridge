package server

import (
	"encoding/json"
	"fmt"
	"log"
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
	ID          string
	BloomFilter *pprl.BloomFilter
	MinHash     *pprl.MinHash
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

// RunAsReceiver implements the receiver mode with fuzzy matching
func RunAsReceiver(cfg *config.Config) {
	fmt.Printf("🔄 Starting receiver on port %d...\n", cfg.ListenPort)

	// Load CSV data
	csvDB, err := db.NewCSVDatabase(cfg.Database.Filename)
	if err != nil {
		log.Fatalf("Failed to load CSV database: %v", err)
	}

	// Convert CSV records to Bloom filters
	records, err := LoadPatientRecordsUtil(csvDB, cfg.Database.Fields)
	if err != nil {
		log.Fatalf("Failed to load patient records: %v", err)
	}

	fmt.Printf("📊 Loaded %d patient records\n", len(records))

	// Start TCP server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("📡 Listening for connections on port %d\n", cfg.ListenPort)
	fmt.Println("💡 Receiver will automatically shutdown after processing one matching session")

	// Accept only one connection and process it
	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("Failed to accept connection: %v", err)
	}

	fmt.Printf("📞 Connection accepted from %s\n", conn.RemoteAddr())
	handleConnection(conn, records, csvDB, cfg.Database.Fields)

	fmt.Println("🔄 Matching session complete. Receiver shutting down automatically.")
}

// handleConnection handles a single client connection
func handleConnection(conn net.Conn, records []PatientRecord, csvDB *db.CSVDatabase, configFields []string) {
	defer conn.Close()

	fmt.Printf("📞 Connection accepted from %s\n", conn.RemoteAddr())

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	// Convert our records to pprl format for potential future pipeline use
	var pprlRecords []*pprl.Record
	for _, record := range records {
		bloomData, err := pprl.BloomToBase64(record.BloomFilter)
		if err != nil {
			log.Printf("Failed to encode Bloom filter for record %s: %v", record.ID, err)
			continue
		}

		// Get MinHash signature
		sig, err := record.MinHash.ComputeSignature(record.BloomFilter)
		if err != nil {
			log.Printf("Failed to compute MinHash signature for record %s: %v", record.ID, err)
			continue
		}

		pprlRecords = append(pprlRecords, &pprl.Record{
			ID:        record.ID,
			BloomData: bloomData,
			MinHash:   sig,
		})
	}

	for {
		var msg MatchingMessage
		if err := decoder.Decode(&msg); err != nil {
			if err.Error() != "EOF" {
				log.Printf("Failed to decode message: %v", err)
			}
			break
		}

		switch msg.Type {
		case "blocking_request":
			fmt.Println("🔐 Processing blocking request...")

			// Send back our blocking data
			ourBlocking := BlockingData{
				EncryptedBuckets: make(map[string][]string),
				Signatures:       make(map[string]string),
			}

			// Simple bucketing for demo
			for _, record := range records {
				bucket := fmt.Sprintf("bucket_%s", record.ID[:1])
				ourBlocking.EncryptedBuckets[bucket] = append(ourBlocking.EncryptedBuckets[bucket], record.ID)
			}

			response := MatchingMessage{
				Type:    "blocking_response",
				Payload: ourBlocking,
			}

			if err := encoder.Encode(response); err != nil {
				log.Printf("Failed to send blocking response: %v", err)
				return
			}

		case "matching_request":
			fmt.Println("🔍 Processing matching request...")

			// Parse sender's matching data
			payloadBytes, _ := json.Marshal(msg.Payload)
			var senderMatching MatchingData
			json.Unmarshal(payloadBytes, &senderMatching)

			// Perform real fuzzy matching between sender and receiver records
			matchResults := make([]*match.MatchResult, 0)

			// Create fuzzy matcher with appropriate thresholds
			fuzzyConfig := &match.FuzzyMatchConfig{
				HammingThreshold:  200, // Allow up to 200 bit differences
				JaccardThreshold:  0.5, // Require at least 50% Jaccard similarity
				UseSecureProtocol: false,
			}

			// Compare ALL receiver records with ALL sender records (not just same IDs)
			fmt.Printf("🔍 Comparing %d receiver records with %d sender records\n",
				len(pprlRecords), len(senderMatching.Records))

			for _, receiverRecord := range pprlRecords {
				for senderID, senderBloomData := range senderMatching.Records {
					// Compare EVERY receiver record with EVERY sender record
					// This is the correct approach for finding actual matches

					// Decode Bloom filters for comparison
					receiverBF, err := pprl.BloomFromBase64(receiverRecord.BloomData)
					if err != nil {
						log.Printf("Failed to decode receiver Bloom filter: %v", err)
						continue
					}

					senderBF, err := pprl.BloomFromBase64(senderBloomData)
					if err != nil {
						log.Printf("Failed to decode sender Bloom filter: %v", err)
						continue
					}

					// Calculate Hamming distance
					hammingDist, err := receiverBF.HammingDistance(senderBF)
					if err != nil {
						log.Printf("Failed to calculate Hamming distance: %v", err)
						continue
					}

					// Determine if this is a match based on thresholds
					isMatch := hammingDist <= fuzzyConfig.HammingThreshold

					// Calculate match score
					matchScore := 1.0
					if hammingDist > 0 {
						bfSize := receiverBF.GetSize()
						matchScore = 1.0 - (float64(hammingDist) / float64(bfSize))
					}

					// Only add to results if similarity is high enough (potential match)
					// Use a stricter threshold for reporting to avoid too many false positives
					if matchScore >= 0.95 { // Lowered from 0.98 to capture genuine matches
						result := &match.MatchResult{
							ID1:               receiverRecord.ID,
							ID2:               senderID,
							IsMatch:           isMatch,
							HammingDistance:   hammingDist,
							JaccardSimilarity: matchScore, // Use match score as similarity estimate
							MatchScore:        matchScore,
						}

						matchResults = append(matchResults, result)

						if isMatch {
							fmt.Printf("   ✅ Potential match: Receiver[%s] <-> Sender[%s] (Hamming: %d, Score: %.3f)\n",
								receiverRecord.ID, senderID, hammingDist, matchScore)
						}
					}
				}
			}

			// Filter for actual matches
			actualMatches := make([]*match.MatchResult, 0)
			for _, result := range matchResults {
				if result.IsMatch {
					actualMatches = append(actualMatches, result)
				}
			}

			fmt.Printf("📊 Matching summary:\n")
			fmt.Printf("   Total comparisons: %d\n", len(matchResults))
			fmt.Printf("   Matches found: %d\n", len(actualMatches))

			// Prepare our matching data response
			ourMatching := MatchingData{
				Records: make(map[string]string),
			}

			for _, record := range records {
				// Encode Bloom filter to base64
				bloomData, err := pprl.BloomToBase64(record.BloomFilter)
				if err != nil {
					log.Printf("Failed to encode Bloom filter: %v", err)
					continue
				}
				ourMatching.Records[record.ID] = bloomData
			}

			response := MatchingMessage{
				Type:    "matching_response",
				Payload: ourMatching,
			}

			if err := encoder.Encode(response); err != nil {
				log.Printf("Failed to send matching response: %v", err)
				return
			}

			// Send final results with actual matches
			fmt.Printf("✅ Matching complete! Found %d matches out of %d comparisons\n",
				len(actualMatches), len(matchResults))
			for _, match := range actualMatches {
				fmt.Printf("   Match: %s <-> %s (Score: %.3f, Hamming: %d, Jaccard: %.3f)\n",
					match.ID1, match.ID2, match.MatchScore, match.HammingDistance, match.JaccardSimilarity)
			}

			// Create output directory if it doesn't exist
			err := os.MkdirAll("out", 0755)
			if err != nil {
				log.Printf("Failed to create out directory: %v", err)
			}

			// Save results to CSV files in out/ directory
			timestamp := time.Now().Format("20060102_150405")
			matchesFile := fmt.Sprintf("out/matches_%s.csv", timestamp)
			detailsFile := fmt.Sprintf("out/match_details_%s.csv", timestamp)

			saveResultsToCSV(actualMatches, matchesFile)
			saveMatchDetailsToCSV(actualMatches, detailsFile, csvDB, configFields)

			results := &match.TwoPartyMatchResult{
				MatchingBuckets: len(records), // All records compared
				CandidatePairs:  len(matchResults),
				TotalMatches:    len(actualMatches),
				MatchResults:    matchResults,
				Matches:         actualMatches,
				Party1Records:   len(records),
				Party2Records:   len(senderMatching.Records),
			}

			finalResponse := MatchingMessage{
				Type:    "results",
				Payload: results,
			}

			if err := encoder.Encode(finalResponse); err != nil {
				log.Printf("Failed to send results: %v", err)
				return
			}

			return // End connection after sending results

		case "shutdown":
			fmt.Println("🔴 Received shutdown signal")
			return

		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// saveResultsToCSV saves match results to a CSV file for easy viewing
func saveResultsToCSV(matches []*match.MatchResult, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create results file %s: %v", filename, err)
		return
	}
	defer file.Close()

	// Write CSV header
	file.WriteString("Receiver_ID,Sender_ID,Match_Score,Hamming_Distance,Jaccard_Similarity,Is_Match\n")

	// Write match data
	for _, match := range matches {
		file.WriteString(fmt.Sprintf("%s,%s,%.3f,%d,%.3f,%t\n",
			match.ID1, match.ID2, match.MatchScore, match.HammingDistance, match.JaccardSimilarity, match.IsMatch))
	}

	fmt.Printf("💾 Match results saved to: %s\n", filename)
}

// saveMatchDetailsToCSV saves match details with patient demographics for debugging
func saveMatchDetailsToCSV(matches []*match.MatchResult, filename string, csvDB *db.CSVDatabase, fields []string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Failed to create match details file %s: %v", filename, err)
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
		log.Printf("Failed to get receiver records: %v", err)
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

	fmt.Printf("💾 Match details saved to: %s\n", filename)
}
