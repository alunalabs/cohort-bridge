// main.go
// End-to-end validation script for the HIPAA-compliant fuzzy matching system
// This script runs the complete matching pipeline and validates against ground truth
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/match"
	"github.com/auroradata-ai/cohort-bridge/internal/server"
)

func main() {
	fmt.Println("🎯 End-to-End Validation Script")
	fmt.Println("===============================")
	fmt.Println("This script runs the complete matching pipeline and validates results against ground truth.")

	// Get validation parameters
	params, err := getValidationParameters()
	if err != nil {
		log.Fatalf("Failed to get validation parameters: %v", err)
	}

	// Load configurations
	cfg1, err := config.Load(params.Config1Path)
	if err != nil {
		log.Fatalf("Failed to load config 1: %v", err)
	}

	cfg2, err := config.Load(params.Config2Path)
	if err != nil {
		log.Fatalf("Failed to load config 2: %v", err)
	}

	// Load ground truth
	groundTruth, err := loadGroundTruth(params.GroundTruthPath)
	if err != nil {
		log.Fatalf("Failed to load ground truth: %v", err)
	}

	fmt.Printf("📊 Loaded %d ground truth matches\n", len(groundTruth))

	// Load datasets
	fmt.Println("📋 Loading datasets...")

	csvDB1, err := db.NewCSVDatabase(cfg1.Database.Filename)
	if err != nil {
		log.Fatalf("Failed to load dataset 1: %v", err)
	}

	csvDB2, err := db.NewCSVDatabase(cfg2.Database.Filename)
	if err != nil {
		log.Fatalf("Failed to load dataset 2: %v", err)
	}

	// Convert to patient records
	records1, err := server.LoadPatientRecordsUtil(csvDB1, cfg1.Database.Fields)
	if err != nil {
		log.Fatalf("Failed to convert dataset 1: %v", err)
	}

	records2, err := server.LoadPatientRecordsUtil(csvDB2, cfg2.Database.Fields)
	if err != nil {
		log.Fatalf("Failed to convert dataset 2: %v", err)
	}

	fmt.Printf("✅ Dataset 1: %d records\n", len(records1))
	fmt.Printf("✅ Dataset 2: %d records\n", len(records2))

	// Run matching pipeline
	fmt.Println("\n🔄 Running matching pipeline...")
	matches, allComparisons, err := runMatchingPipeline(records1, records2, params)
	if err != nil {
		log.Fatalf("Matching pipeline failed: %v", err)
	}

	// Validate results
	fmt.Println("\n📈 Validating results...")
	validation := validateResults(matches, allComparisons, groundTruth)

	// Display results
	displayValidationResults(validation, len(groundTruth), len(matches))

	// Save detailed results
	if err := saveValidationResults(validation, params.OutputPath); err != nil {
		log.Printf("Failed to save validation results: %v", err)
	} else {
		fmt.Printf("💾 Validation results saved to: %s\n", params.OutputPath)
	}
}

type ValidationParameters struct {
	Config1Path        string
	Config2Path        string
	GroundTruthPath    string
	OutputPath         string
	HammingThreshold   uint32
	JaccardThreshold   float64
	CandidateThreshold float64 // Minimum similarity score to be considered a candidate
	MatchThreshold     uint32  // Hamming distance threshold for matches
}

func getValidationParameters() (*ValidationParameters, error) {
	params := &ValidationParameters{}

	// Check if command line arguments are provided
	args := os.Args[1:]

	if len(args) >= 3 {
		// Use command line arguments
		params.Config1Path = args[0]
		params.Config2Path = args[1]
		params.GroundTruthPath = args[2]

		// Optional output path (default if not provided)
		if len(args) >= 4 {
			params.OutputPath = args[3]
		} else {
			params.OutputPath = "validation_results.csv"
		}

		// Optional candidate threshold (default if not provided)
		if len(args) >= 5 {
			if threshold, err := strconv.ParseFloat(args[4], 64); err == nil {
				params.CandidateThreshold = threshold
			} else {
				return nil, fmt.Errorf("invalid candidate threshold: %s", args[4])
			}
		} else {
			params.CandidateThreshold = 0.95 // Default candidate threshold
		}

		// Optional Hamming threshold (default if not provided)
		if len(args) >= 6 {
			if threshold, err := strconv.ParseUint(args[5], 10, 32); err == nil {
				params.MatchThreshold = uint32(threshold)
			} else {
				return nil, fmt.Errorf("invalid Hamming threshold: %s", args[5])
			}
		} else {
			params.MatchThreshold = 100 // Default Hamming threshold
		}

		fmt.Printf("📝 Using arguments:\n")
		fmt.Printf("  Config 1: %s\n", params.Config1Path)
		fmt.Printf("  Config 2: %s\n", params.Config2Path)
		fmt.Printf("  Ground Truth: %s\n", params.GroundTruthPath)
		fmt.Printf("  Output: %s\n", params.OutputPath)
		fmt.Printf("  Candidate Threshold: %.3f\n", params.CandidateThreshold)
		fmt.Printf("  Hamming Threshold: %d\n", params.MatchThreshold)

	} else {
		fmt.Printf("❌ Usage: %s <config1> <config2> <ground_truth> [output_file] [candidate_threshold] [hamming_threshold]\n", os.Args[0])
		fmt.Printf("Example: %s config_a.yaml config_b.yaml data/expected_matches.csv validation_results.csv 0.95 100\n", os.Args[0])
		fmt.Printf("  candidate_threshold: Minimum similarity score to be considered (default: 0.95)\n")
		fmt.Printf("  hamming_threshold: Maximum Hamming distance for match (default: 100)\n")
		return nil, fmt.Errorf("insufficient arguments provided")
	}

	// Validate that files exist
	if _, err := os.Stat(params.Config1Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file 1 does not exist: %s", params.Config1Path)
	}

	if _, err := os.Stat(params.Config2Path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file 2 does not exist: %s", params.Config2Path)
	}

	if _, err := os.Stat(params.GroundTruthPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("ground truth file does not exist: %s", params.GroundTruthPath)
	}

	// Set remaining thresholds based on the match threshold
	params.HammingThreshold = params.MatchThreshold
	params.JaccardThreshold = 0.5 // Not used for match determination, just for reference

	return params, nil
}

func loadGroundTruth(path string) (map[string]string, error) {
	groundTruthDB, err := db.NewCSVDatabase(path)
	if err != nil {
		return nil, err
	}

	records, err := groundTruthDB.List(0, 1000000)
	if err != nil {
		return nil, err
	}

	groundTruth := make(map[string]string)
	for _, record := range records {
		patientA := record["patient_a_id"]
		patientB := record["patient_b_id"]
		if patientA != "" && patientB != "" {
			groundTruth[patientA] = patientB
		}
	}

	return groundTruth, nil
}

func runMatchingPipeline(records1, records2 []server.PatientRecord, params *ValidationParameters) ([]*match.MatchResult, []*match.MatchResult, error) {
	var allComparisons []*match.MatchResult
	var matches []*match.MatchResult

	// Create fuzzy matcher with specified thresholds (same as receiver)
	fuzzyConfig := &match.FuzzyMatchConfig{
		HammingThreshold:  params.HammingThreshold, // Default: 100
		JaccardThreshold:  params.JaccardThreshold, // Not used for match determination
		UseSecureProtocol: false,
	}

	fmt.Printf("🔍 Comparing %d records from Party 1 with %d records from Party 2\n",
		len(records1), len(records2))
	fmt.Printf("🔧 Using thresholds: Hamming <= %d, Candidate Score >= %.3f\n",
		params.HammingThreshold, params.CandidateThreshold)

	totalComparisons := 0
	candidatesFound := 0

	// Compare all records from Party 1 with all records from Party 2
	for _, record1 := range records1 {
		for _, record2 := range records2 {
			totalComparisons++

			// Calculate Hamming distance
			hammingDist, err := record1.BloomFilter.HammingDistance(record2.BloomFilter)
			if err != nil {
				continue // Skip on error
			}

			// Calculate match score
			bfSize := record1.BloomFilter.GetSize()
			matchScore := 1.0
			if hammingDist > 0 {
				matchScore = 1.0 - (float64(hammingDist) / float64(bfSize))
			}

			// Determine if this is a match based on Hamming threshold ONLY (same as receiver)
			isMatch := hammingDist <= fuzzyConfig.HammingThreshold

			// Create match result for all comparisons
			result := &match.MatchResult{
				ID1:               record1.ID,
				ID2:               record2.ID,
				IsMatch:           isMatch,
				HammingDistance:   hammingDist,
				JaccardSimilarity: matchScore, // Use match score as similarity estimate
				MatchScore:        matchScore,
			}

			// Store all comparisons
			allComparisons = append(allComparisons, result)

			// Only add to matches if similarity is high enough (configurable candidate threshold)
			if matchScore >= params.CandidateThreshold {
				candidatesFound++
				matches = append(matches, result)

				if isMatch {
					fmt.Printf("   ✅ Potential match: %s <-> %s (Hamming: %d, Score: %.3f)\n",
						record1.ID, record2.ID, hammingDist, matchScore)
				}
			}
		}
	}

	// Filter for actual matches only from the candidates
	actualMatches := make([]*match.MatchResult, 0)
	for _, result := range matches {
		if result.IsMatch {
			actualMatches = append(actualMatches, result)
		}
	}

	fmt.Printf("📊 Pipeline results: %d total comparisons, %d candidates (score >= %.3f), %d matches found\n",
		totalComparisons, candidatesFound, params.CandidateThreshold, len(actualMatches))

	return actualMatches, allComparisons, nil
}

type ValidationResult struct {
	TruePositives              int
	FalsePositives             int
	FalseNegatives             int
	Precision                  float64
	Recall                     float64
	F1Score                    float64
	MatchedPairs               []MatchPair
	MissedMatches              []string
	MissedMatchPairs           []MatchPair // False negatives with scores
	FalseMatches               []MatchPair
	LowestGroundTruthScore     float64 // Lowest score among all ground truth pairs (TP + FN)
	HighestNonGroundTruthScore float64 // Highest score among all non-ground truth pairs (FP + TN)
}

type MatchPair struct {
	ID1   string
	ID2   string
	Score float64
}

func validateResults(matches []*match.MatchResult, allComparisons []*match.MatchResult, groundTruth map[string]string) *ValidationResult {
	result := &ValidationResult{
		MatchedPairs:     make([]MatchPair, 0),
		MissedMatches:    make([]string, 0),
		MissedMatchPairs: make([]MatchPair, 0),
		FalseMatches:     make([]MatchPair, 0),
	}

	// Create a set of predicted matches
	predictedMatches := make(map[string]string)
	for _, match := range matches {
		if match.IsMatch {
			predictedMatches[match.ID1] = match.ID2
			result.MatchedPairs = append(result.MatchedPairs, MatchPair{
				ID1:   match.ID1,
				ID2:   match.ID2,
				Score: match.MatchScore,
			})
		}
	}

	// Create a lookup map for all comparison results (including non-matches)
	allComparisonsMap := make(map[string]map[string]*match.MatchResult)
	for _, comparison := range allComparisons {
		if allComparisonsMap[comparison.ID1] == nil {
			allComparisonsMap[comparison.ID1] = make(map[string]*match.MatchResult)
		}
		allComparisonsMap[comparison.ID1][comparison.ID2] = comparison
	}

	// Calculate True Positives and False Negatives
	for groundID1, groundID2 := range groundTruth {
		if predictedID2, exists := predictedMatches[groundID1]; exists {
			if predictedID2 == groundID2 {
				result.TruePositives++
			} else {
				result.FalseNegatives++
				result.MissedMatches = append(result.MissedMatches, fmt.Sprintf("%s -> %s", groundID1, groundID2))

				// Find the actual comparison score for this false negative
				if compResults, exists := allComparisonsMap[groundID1]; exists {
					if compResult, exists := compResults[groundID2]; exists {
						result.MissedMatchPairs = append(result.MissedMatchPairs, MatchPair{
							ID1:   groundID1,
							ID2:   groundID2,
							Score: compResult.MatchScore,
						})
					} else {
						// No comparison found, score would be very low
						result.MissedMatchPairs = append(result.MissedMatchPairs, MatchPair{
							ID1:   groundID1,
							ID2:   groundID2,
							Score: 0.0,
						})
					}
				} else {
					// No comparison found, score would be very low
					result.MissedMatchPairs = append(result.MissedMatchPairs, MatchPair{
						ID1:   groundID1,
						ID2:   groundID2,
						Score: 0.0,
					})
				}
			}
		} else {
			result.FalseNegatives++
			result.MissedMatches = append(result.MissedMatches, fmt.Sprintf("%s -> %s", groundID1, groundID2))

			// Find the actual comparison score for this false negative
			if compResults, exists := allComparisonsMap[groundID1]; exists {
				if compResult, exists := compResults[groundID2]; exists {
					result.MissedMatchPairs = append(result.MissedMatchPairs, MatchPair{
						ID1:   groundID1,
						ID2:   groundID2,
						Score: compResult.MatchScore,
					})
				} else {
					// No comparison found, score would be very low
					result.MissedMatchPairs = append(result.MissedMatchPairs, MatchPair{
						ID1:   groundID1,
						ID2:   groundID2,
						Score: 0.0,
					})
				}
			} else {
				// No comparison found, score would be very low
				result.MissedMatchPairs = append(result.MissedMatchPairs, MatchPair{
					ID1:   groundID1,
					ID2:   groundID2,
					Score: 0.0,
				})
			}
		}
	}

	// Calculate False Positives
	for predID1, predID2 := range predictedMatches {
		if groundID2, exists := groundTruth[predID1]; !exists || groundID2 != predID2 {
			result.FalsePositives++
			// Find the score for this false match
			for _, match := range allComparisons {
				if match.ID1 == predID1 && match.ID2 == predID2 {
					result.FalseMatches = append(result.FalseMatches, MatchPair{
						ID1:   predID1,
						ID2:   predID2,
						Score: match.MatchScore,
					})
					break
				}
			}
		}
	}

	// Calculate metrics
	if result.TruePositives+result.FalsePositives > 0 {
		result.Precision = float64(result.TruePositives) / float64(result.TruePositives+result.FalsePositives)
	}

	if result.TruePositives+result.FalseNegatives > 0 {
		result.Recall = float64(result.TruePositives) / float64(result.TruePositives+result.FalseNegatives)
	}

	if result.Precision+result.Recall > 0 {
		result.F1Score = 2 * (result.Precision * result.Recall) / (result.Precision + result.Recall)
	}

	// Calculate lowest score for actual match and highest score for missed match
	lowestGroundTruthScore := 1.0     // Lowest score among all ground truth pairs (TP + FN)
	highestNonGroundTruthScore := 0.0 // Highest score among all non-ground truth pairs (FP + TN)

	// Create a set of ground truth pairs for quick lookup
	groundTruthPairs := make(map[string]bool)
	for id1, id2 := range groundTruth {
		groundTruthPairs[id1+"->"+id2] = true
	}

	// Go through all comparisons and categorize them
	for _, comparison := range allComparisons {
		pairKey := comparison.ID1 + "->" + comparison.ID2
		isGroundTruthPair := groundTruthPairs[pairKey]

		if isGroundTruthPair {
			// This is a ground truth pair (TP or FN)
			if comparison.MatchScore < lowestGroundTruthScore {
				lowestGroundTruthScore = comparison.MatchScore
			}
		} else {
			// This is NOT a ground truth pair (FP or TN)
			if comparison.MatchScore > highestNonGroundTruthScore {
				highestNonGroundTruthScore = comparison.MatchScore
			}
		}
	}

	result.LowestGroundTruthScore = lowestGroundTruthScore
	result.HighestNonGroundTruthScore = highestNonGroundTruthScore

	return result
}

func displayValidationResults(validation *ValidationResult, groundTruthCount, matchesFound int) {
	fmt.Println("\n🏆 Validation Results")
	fmt.Println("====================")

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("  Ground truth matches: %d\n", groundTruthCount)
	fmt.Printf("  System found matches: %d\n", matchesFound)
	fmt.Printf("  True Positives: %d\n", validation.TruePositives)
	fmt.Printf("  False Positives: %d\n", validation.FalsePositives)
	fmt.Printf("  False Negatives: %d\n", validation.FalseNegatives)

	fmt.Printf("\n🎯 Evaluation Metrics:\n")
	fmt.Printf("  Precision: %.3f\n", validation.Precision)
	fmt.Printf("  Recall: %.3f\n", validation.Recall)
	fmt.Printf("  F1-Score: %.3f\n", validation.F1Score)

	fmt.Printf("\n📈 Score Analysis:\n")
	fmt.Printf("  Lowest score for ground truth pairs (TP+FN): %.3f\n", validation.LowestGroundTruthScore)
	fmt.Printf("  Highest score for non-ground truth pairs (FP+TN): %.3f\n", validation.HighestNonGroundTruthScore)
	if validation.HighestNonGroundTruthScore > validation.LowestGroundTruthScore {
		fmt.Printf("  ⚠️  Score overlap detected: Some non-matches have higher scores than true matches!\n")
	} else {
		fmt.Printf("  ✅ Clear score separation between ground truth and non-ground truth pairs\n")
	}

	// Quality assessment
	var precision, recall, f1 string

	if validation.Precision >= 0.9 {
		precision = "Excellent"
	} else if validation.Precision >= 0.8 {
		precision = "Good"
	} else if validation.Precision >= 0.7 {
		precision = "Fair"
	} else {
		precision = "Poor"
	}

	if validation.Recall >= 0.9 {
		recall = "Excellent"
	} else if validation.Recall >= 0.8 {
		recall = "Good"
	} else if validation.Recall >= 0.7 {
		recall = "Fair"
	} else {
		recall = "Poor"
	}

	if validation.F1Score >= 0.9 {
		f1 = "Excellent"
	} else if validation.F1Score >= 0.8 {
		f1 = "Good"
	} else if validation.F1Score >= 0.7 {
		f1 = "Fair"
	} else {
		f1 = "Poor"
	}

	fmt.Printf("\n🔍 Quality Assessment:\n")
	fmt.Printf("  Precision: %s (%.3f)\n", precision, validation.Precision)
	fmt.Printf("  Recall: %s (%.3f)\n", recall, validation.Recall)
	fmt.Printf("  Overall: %s (F1: %.3f)\n", f1, validation.F1Score)

	// Show some examples if there are errors
	if validation.FalsePositives > 0 {
		fmt.Printf("\n❌ False Positives (first 5):\n")
		for i, fp := range validation.FalseMatches {
			if i >= 5 {
				break
			}
			fmt.Printf("  %s <-> %s (Score: %.3f)\n", fp.ID1, fp.ID2, fp.Score)
		}
	}

	if validation.FalseNegatives > 0 {
		fmt.Printf("\n⚠️  Missed Matches (first 5):\n")
		for i, missed := range validation.MissedMatches {
			if i >= 5 {
				break
			}
			fmt.Printf("  %s\n", missed)
		}
	}
}

func saveValidationResults(validation *ValidationResult, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write summary header
	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"True_Positives", fmt.Sprintf("%d", validation.TruePositives)})
	writer.Write([]string{"False_Positives", fmt.Sprintf("%d", validation.FalsePositives)})
	writer.Write([]string{"False_Negatives", fmt.Sprintf("%d", validation.FalseNegatives)})
	writer.Write([]string{"Precision", fmt.Sprintf("%.6f", validation.Precision)})
	writer.Write([]string{"Recall", fmt.Sprintf("%.6f", validation.Recall)})
	writer.Write([]string{"F1_Score", fmt.Sprintf("%.6f", validation.F1Score)})
	writer.Write([]string{"Lowest_Ground_Truth_Score", fmt.Sprintf("%.6f", validation.LowestGroundTruthScore)})
	writer.Write([]string{"Highest_Non_Ground_Truth_Score", fmt.Sprintf("%.6f", validation.HighestNonGroundTruthScore)})

	// Write detailed results
	writer.Write([]string{""}) // Empty row
	writer.Write([]string{"Type", "ID1", "ID2", "Score"})

	for _, match := range validation.MatchedPairs {
		writer.Write([]string{"True_Positive", match.ID1, match.ID2, fmt.Sprintf("%.6f", match.Score)})
	}

	for _, match := range validation.FalseMatches {
		writer.Write([]string{"False_Positive", match.ID1, match.ID2, fmt.Sprintf("%.6f", match.Score)})
	}

	for _, match := range validation.MissedMatchPairs {
		writer.Write([]string{"False_Negative", match.ID1, match.ID2, fmt.Sprintf("%.6f", match.Score)})
	}

	return nil
}
