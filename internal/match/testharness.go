// testharness.go
// Package match provides a comprehensive test harness for validating the secure fuzzy matching system
// with synthetic datasets containing overlapping and noisy patient data.
package match

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/pprl"
)

// TestHarness provides functionality for testing the matching pipeline
type TestHarness struct {
	config        *TestConfig
	storage1      *pprl.Storage
	storage2      *pprl.Storage
	groundTruth   map[string]string // Maps ID1 -> ID2 for known matches
	sharedMinHash *pprl.MinHash     // Shared MinHash instance for consistent signatures
}

// TestConfig defines the configuration for testing
type TestConfig struct {
	NumRecords1       int     `json:"num_records_1"`      // Records in dataset 1
	NumRecords2       int     `json:"num_records_2"`      // Records in dataset 2
	OverlapRate       float64 `json:"overlap_rate"`       // Fraction of records that should match
	NoiseRate         float64 `json:"noise_rate"`         // Rate of character-level noise
	BloomFilterSize   uint32  `json:"bloom_filter_size"`  // Size of Bloom filters
	BloomHashCount    uint32  `json:"bloom_hash_count"`   // Number of hash functions
	MinHashSignatures uint32  `json:"minhash_signatures"` // Length of MinHash signatures
	OutputDir         string  `json:"output_dir"`         // Directory for test outputs
}

// NewTestHarness creates a new test harness
func NewTestHarness(config *TestConfig) (*TestHarness, error) {
	storage1, err := pprl.NewStorage(config.OutputDir + "/dataset1.jsonl")
	if err != nil {
		return nil, fmt.Errorf("failed to create storage1: %w", err)
	}

	storage2, err := pprl.NewStorage(config.OutputDir + "/dataset2.jsonl")
	if err != nil {
		return nil, fmt.Errorf("failed to create storage2: %w", err)
	}

	// Create a shared MinHash instance for consistent signatures
	sharedMinHash, err := pprl.NewMinHashSeeded(config.BloomFilterSize, config.MinHashSignatures, "cohort-bridge-pprl-seed")
	if err != nil {
		return nil, fmt.Errorf("failed to create shared MinHash: %w", err)
	}

	return &TestHarness{
		config:        config,
		storage1:      storage1,
		storage2:      storage2,
		groundTruth:   make(map[string]string),
		sharedMinHash: sharedMinHash,
	}, nil
}

// GenerateTestData creates synthetic datasets with known overlaps and noise
func (th *TestHarness) GenerateTestData() error {
	log.Println("Generating synthetic test datasets...")

	rand.Seed(time.Now().UnixNano())

	// Generate base patient data
	baseRecords := th.generateBaseRecords()

	// Create dataset 1
	dataset1 := th.createDataset1(baseRecords)
	if err := th.saveDataset(dataset1, th.storage1); err != nil {
		return fmt.Errorf("failed to save dataset1: %w", err)
	}

	// Create dataset 2 with overlaps and noise
	dataset2 := th.createDataset2(baseRecords)
	if err := th.saveDataset(dataset2, th.storage2); err != nil {
		return fmt.Errorf("failed to save dataset2: %w", err)
	}

	log.Printf("Generated datasets: %d records in dataset1, %d in dataset2, %d ground truth matches",
		len(dataset1), len(dataset2), len(th.groundTruth))

	return nil
}

// PatientRecord represents a synthetic patient record
type PatientRecord struct {
	ID        string
	FirstName string
	LastName  string
	DOB       string // YYYY-MM-DD format
	SSN       string
	Address   string
	Phone     string
}

// generateBaseRecords creates a set of base patient records
func (th *TestHarness) generateBaseRecords() []*PatientRecord {
	firstNames := []string{"John", "Jane", "Michael", "Sarah", "David", "Emily", "Robert", "Lisa", "William", "Jennifer"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}

	var records []*PatientRecord
	numBase := int(float64(th.config.NumRecords1) * th.config.OverlapRate)

	for i := 0; i < numBase; i++ {
		record := &PatientRecord{
			ID:        fmt.Sprintf("base_%d", i),
			FirstName: firstNames[rand.Intn(len(firstNames))],
			LastName:  lastNames[rand.Intn(len(lastNames))],
			DOB:       th.randomDate(),
			SSN:       th.randomSSN(),
			Address:   th.randomAddress(),
			Phone:     th.randomPhone(),
		}
		records = append(records, record)
	}

	return records
}

// createDataset1 creates the first dataset
func (th *TestHarness) createDataset1(baseRecords []*PatientRecord) []*pprl.Record {
	var dataset []*pprl.Record

	// Add base records (for overlaps)
	for _, base := range baseRecords {
		record, err := th.convertToBloomRecord(base, fmt.Sprintf("d1_%s", base.ID))
		if err != nil {
			log.Printf("Failed to convert record %s: %v", base.ID, err)
			continue
		}
		dataset = append(dataset, record)
	}

	// Add additional unique records
	remaining := th.config.NumRecords1 - len(baseRecords)
	for i := 0; i < remaining; i++ {
		unique := th.generateRandomRecord(fmt.Sprintf("d1_unique_%d", i))
		record, err := th.convertToBloomRecord(unique, unique.ID)
		if err != nil {
			log.Printf("Failed to convert record %s: %v", unique.ID, err)
			continue
		}
		dataset = append(dataset, record)
	}

	return dataset
}

// createDataset2 creates the second dataset with noise and overlaps
func (th *TestHarness) createDataset2(baseRecords []*PatientRecord) []*pprl.Record {
	var dataset []*pprl.Record

	// Add base records with noise (for overlaps)
	for _, base := range baseRecords {
		noisy := th.addNoise(base)
		noisy.ID = fmt.Sprintf("d2_%s", base.ID)

		record, err := th.convertToBloomRecord(noisy, noisy.ID)
		if err != nil {
			log.Printf("Failed to convert record %s: %v", noisy.ID, err)
			continue
		}
		dataset = append(dataset, record)

		// Record ground truth mapping
		th.groundTruth[fmt.Sprintf("d1_%s", base.ID)] = noisy.ID
	}

	// Add additional unique records
	remaining := th.config.NumRecords2 - len(baseRecords)
	for i := 0; i < remaining; i++ {
		unique := th.generateRandomRecord(fmt.Sprintf("d2_unique_%d", i))
		record, err := th.convertToBloomRecord(unique, unique.ID)
		if err != nil {
			log.Printf("Failed to convert record %s: %v", unique.ID, err)
			continue
		}
		dataset = append(dataset, record)
	}

	return dataset
}

// convertToBloomRecord converts a PatientRecord to a pprl.Record with Bloom filter and MinHash
func (th *TestHarness) convertToBloomRecord(patient *PatientRecord, id string) (*pprl.Record, error) {
	// Create Bloom filter
	bf := pprl.NewBloomFilter(th.config.BloomFilterSize, th.config.BloomHashCount)
	if bf == nil {
		return nil, fmt.Errorf("failed to create bloom filter")
	}

	// Add patient data to Bloom filter using n-grams
	th.addPatientDataToBloom(bf, patient)

	// Compute MinHash signature using shared instance
	signature, err := th.sharedMinHash.ComputeSignature(bf)
	if err != nil {
		return nil, fmt.Errorf("failed to compute minhash signature: %w", err)
	}

	// Serialize Bloom filter
	bloomData, err := pprl.BloomToBase64(bf)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize bloom filter: %w", err)
	}

	return &pprl.Record{
		ID:        id,
		BloomData: bloomData,
		MinHash:   signature,
	}, nil
}

// addPatientDataToBloom adds patient data to a Bloom filter using n-grams
func (th *TestHarness) addPatientDataToBloom(bf *pprl.BloomFilter, patient *PatientRecord) {
	// Use 2-grams for names and other text fields
	fields := []string{
		strings.ToLower(patient.FirstName),
		strings.ToLower(patient.LastName),
		patient.DOB,
		patient.SSN,
		strings.ToLower(patient.Address),
		patient.Phone,
	}

	for _, field := range fields {
		// Add whole field
		bf.Add([]byte(field))

		// Add 2-grams
		if len(field) >= 2 {
			for i := 0; i <= len(field)-2; i++ {
				bigram := field[i : i+2]
				bf.Add([]byte(bigram))
			}
		}

		// Add 3-grams for longer fields
		if len(field) >= 3 {
			for i := 0; i <= len(field)-3; i++ {
				trigram := field[i : i+3]
				bf.Add([]byte(trigram))
			}
		}
	}
}

// addNoise introduces noise into patient data
func (th *TestHarness) addNoise(original *PatientRecord) *PatientRecord {
	noisy := &PatientRecord{
		ID:        original.ID,
		FirstName: th.addStringNoise(original.FirstName),
		LastName:  th.addStringNoise(original.LastName),
		DOB:       original.DOB, // Keep DOB exact for better matching
		SSN:       original.SSN, // Keep SSN exact
		Address:   th.addStringNoise(original.Address),
		Phone:     original.Phone, // Keep phone exact
	}
	return noisy
}

// addStringNoise adds character-level noise to a string
func (th *TestHarness) addStringNoise(s string) string {
	if rand.Float64() > th.config.NoiseRate {
		return s // No noise
	}

	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}

	// Choose random position to modify
	pos := rand.Intn(len(runes))

	switch rand.Intn(3) {
	case 0: // Substitute character
		runes[pos] = rune('a' + rand.Intn(26))
	case 1: // Delete character
		if len(runes) > 1 {
			runes = append(runes[:pos], runes[pos+1:]...)
		}
	case 2: // Insert character
		newChar := rune('a' + rand.Intn(26))
		runes = append(runes[:pos], append([]rune{newChar}, runes[pos:]...)...)
	}

	return string(runes)
}

// Helper functions for generating random data
func (th *TestHarness) randomDate() string {
	year := 1950 + rand.Intn(50)
	month := 1 + rand.Intn(12)
	day := 1 + rand.Intn(28)
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func (th *TestHarness) randomSSN() string {
	return fmt.Sprintf("%03d-%02d-%04d",
		100+rand.Intn(899),
		10+rand.Intn(89),
		1000+rand.Intn(8999))
}

func (th *TestHarness) randomAddress() string {
	streets := []string{"Main St", "Oak Ave", "Pine Rd", "Elm Dr", "Cedar Ln"}
	return fmt.Sprintf("%d %s",
		100+rand.Intn(9900),
		streets[rand.Intn(len(streets))])
}

func (th *TestHarness) randomPhone() string {
	return fmt.Sprintf("(%03d) %03d-%04d",
		200+rand.Intn(799),
		200+rand.Intn(799),
		1000+rand.Intn(8999))
}

func (th *TestHarness) generateRandomRecord(id string) *PatientRecord {
	firstNames := []string{"John", "Jane", "Michael", "Sarah", "David"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones"}

	return &PatientRecord{
		ID:        id,
		FirstName: firstNames[rand.Intn(len(firstNames))],
		LastName:  lastNames[rand.Intn(len(lastNames))],
		DOB:       th.randomDate(),
		SSN:       th.randomSSN(),
		Address:   th.randomAddress(),
		Phone:     th.randomPhone(),
	}
}

// saveDataset saves a dataset to storage
func (th *TestHarness) saveDataset(dataset []*pprl.Record, storage *pprl.Storage) error {
	return storage.WriteAll(dataset)
}

// RunTest executes a complete test of the matching pipeline
func (th *TestHarness) RunTest(pipelineConfig *PipelineConfig) (*TestResults, error) {
	log.Println("Starting test harness...")

	// Generate test data
	if err := th.GenerateTestData(); err != nil {
		return nil, fmt.Errorf("failed to generate test data: %w", err)
	}

	// Create two pipelines
	pipeline1, err := NewPipeline(pipelineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline1: %w", err)
	}

	pipeline2, err := NewPipeline(pipelineConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pipeline2: %w", err)
	}

	// Load data into pipelines
	if err := pipeline1.LoadRecords(th.storage1); err != nil {
		return nil, fmt.Errorf("failed to load records into pipeline1: %w", err)
	}

	if err := pipeline2.LoadRecords(th.storage2); err != nil {
		return nil, fmt.Errorf("failed to load records into pipeline2: %w", err)
	}

	// Run two-party matching simulation
	matchResult, err := pipeline1.SimulateTwoPartyMatching(pipeline2)
	if err != nil {
		return nil, fmt.Errorf("failed to run two-party matching: %w", err)
	}

	// Evaluate results against ground truth
	evaluation := th.EvaluateResults(matchResult.PrivateMatches)

	return &TestResults{
		MatchResult:      matchResult,
		Evaluation:       evaluation,
		GroundTruthCount: len(th.groundTruth),
		Pipeline1Stats:   pipeline1.GetStats(),
		Pipeline2Stats:   pipeline2.GetStats(),
	}, nil
}

// EvaluateResults compares match results against ground truth
func (th *TestHarness) EvaluateResults(matches []*PrivateMatchResult) *Evaluation {
	var truePositives, falsePositives, falseNegatives int

	foundMatches := make(map[string]bool)

	// Check each found match against ground truth
	for _, match := range matches {
		key1 := match.LocalID + "->" + match.PeerID
		key2 := match.PeerID + "->" + match.LocalID

		if th.isGroundTruthMatch(match.LocalID, match.PeerID) {
			truePositives++
			foundMatches[key1] = true
			foundMatches[key2] = true
		} else {
			falsePositives++
		}
	}

	// Count false negatives (ground truth matches not found)
	for id1, id2 := range th.groundTruth {
		key := id1 + "->" + id2
		if !foundMatches[key] {
			falseNegatives++
		}
	}

	// Calculate metrics
	precision := float64(truePositives) / float64(truePositives+falsePositives)
	recall := float64(truePositives) / float64(truePositives+falseNegatives)
	f1Score := 2 * (precision * recall) / (precision + recall)

	if truePositives+falsePositives == 0 {
		precision = 0
	}
	if truePositives+falseNegatives == 0 {
		recall = 0
	}
	if precision+recall == 0 {
		f1Score = 0
	}

	return &Evaluation{
		TruePositives:  truePositives,
		FalsePositives: falsePositives,
		FalseNegatives: falseNegatives,
		Precision:      precision,
		Recall:         recall,
		F1Score:        f1Score,
	}
}

// isGroundTruthMatch checks if two IDs represent a ground truth match
func (th *TestHarness) isGroundTruthMatch(id1, id2 string) bool {
	return th.groundTruth[id1] == id2 || th.groundTruth[id2] == id1
}

// TestResults contains the complete results of a test run
type TestResults struct {
	MatchResult      *TwoPartyMatchResult `json:"match_result"`
	Evaluation       *Evaluation          `json:"evaluation"`
	GroundTruthCount int                  `json:"ground_truth_count"`
	Pipeline1Stats   *PipelineStats       `json:"pipeline1_stats"`
	Pipeline2Stats   *PipelineStats       `json:"pipeline2_stats"`
}

// Evaluation contains metrics for evaluating match quality
type Evaluation struct {
	TruePositives  int     `json:"true_positives"`
	FalsePositives int     `json:"false_positives"`
	FalseNegatives int     `json:"false_negatives"`
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	F1Score        float64 `json:"f1_score"`
}
