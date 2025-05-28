package server

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
)

// GetPatientRecords reads patient data from the CSV file specified in the config.
// It handles multi-column CSVs and maps columns based on CSV header.
func GetPatientRecords(cfg *config.Config) []map[string]interface{} {
	if cfg.Database.Type != "csv" {
		fmt.Printf("Unsupported database type: %s\n", cfg.Database.Type)
		return nil
	}
	// Use Database.Filename, as it's more specific for CSV files in the config structure
	csvPath := cfg.Database.Filename
	if csvPath == "" {
		fmt.Println("CSV filename not specified in config.Database.Filename")
		return nil
	}

	file, err := os.Open(csvPath)
	if err != nil {
		fmt.Printf("Error opening CSV file %s: %v\n", csvPath, err)
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	allRows, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV data from %s: %v\n", csvPath, err)
		return nil
	}

	if len(allRows) < 1 { // Must have at least a header
		fmt.Printf("CSV file %s is empty or has no header.\n", csvPath)
		return nil
	}

	header := allRows[0]
	var dataRows [][]string
	if len(allRows) > 1 {
		dataRows = allRows[1:]
	}

	fieldIndexMap := make(map[string]int)
	for i, fieldName := range header {
		fieldIndexMap[strings.TrimSpace(fieldName)] = i
	}

	var records []map[string]interface{}
	for _, row := range dataRows {
		rec := make(map[string]interface{})
		// Populate record with all fields from CSV header
		for fieldNameFromHeader, idx := range fieldIndexMap {
			if idx < len(row) {
				rec[fieldNameFromHeader] = row[idx]
			} else {
				rec[fieldNameFromHeader] = "" // Or some other placeholder for missing data
			}
		}
		records = append(records, rec)
	}
	return records
}

// GetOriginalIDsFromDB extracts the 'id' field from records.
// It relies on GetPatientRecords to load the data.
func GetOriginalIDsFromDB(cfg *config.Config) []string {
	var originalIDs []string
	records := GetPatientRecords(cfg)
	if records == nil {
		fmt.Println("Warning: GetPatientRecords returned nil in GetOriginalIDsFromDB")
		return originalIDs
	}

	for _, rec := range records {
		idVal, ok := rec["id"]
		if !ok || idVal == nil {
			fmt.Printf("Warning: record missing 'id' field or id is nil in GetOriginalIDsFromDB. Record: %v\n", rec)
			continue
		}
		id, ok := idVal.(string)
		if !ok {
			id = fmt.Sprintf("%v", idVal)
		}
		originalIDs = append(originalIDs, id)
	}
	return originalIDs
}

// BuildTokenMapForPSI creates a map of H(originalID)_hex -> originalID.
// This is used by the PSI sender.
func BuildTokenMapForPSI(cfg *config.Config) map[string]string {
	tokenMap := make(map[string]string)
	records := GetPatientRecords(cfg)
	if records == nil {
		fmt.Println("Warning: no records for PSI token building")
		return tokenMap
	}
	for _, rec := range records {
		val, ok := rec["id"]
		if !ok || val == nil {
			continue
		}
		id := fmt.Sprint(val)
		P := crypto.HashToCurve(id)
		tokenHex := hex.EncodeToString(P.Bytes())
		tokenMap[tokenHex] = tokenHex
	}
	return tokenMap
}
