package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// TokenizeRecords normalizes and tokenizes patient records using HMAC-SHA256.
func TokenizeRecords(records []map[string]any, fields []string, salt string) ([]string, error) {
	return TokenizeRecordsWithNormalization(records, fields, salt, nil)
}

// TokenizeRecordsWithNormalization normalizes and tokenizes patient records with custom normalization.
func TokenizeRecordsWithNormalization(records []map[string]any, fields []string, salt string, normalizationConfig map[string]NormalizationMethod) ([]string, error) {
	var pseudonyms []string
	for _, record := range records {
		var parts []string
		for _, field := range fields {
			val := record[field]
			var norm string

			// Check if we have a specific normalization method for this field
			if normalizationConfig != nil {
				if method, exists := normalizationConfig[field]; exists {
					norm = NormalizeField(val, method)
				} else {
					// No specific normalization configured, apply basic normalization
					norm = NormalizeField(val, "")
				}
			} else {
				// Legacy normalization for backward compatibility
				switch v := val.(type) {
				case string:
					norm = strings.ToLower(strings.TrimSpace(v))
					if field == "zip" {
						norm = strings.ReplaceAll(norm, " ", "")
					}
				case time.Time:
					norm = v.Format("2006-01-02")
				case []byte:
					norm = string(v)
				default:
					if v != nil {
						norm = strings.ToLower(strings.TrimSpace(fmt.Sprint(v)))
					}
				}
			}
			parts = append(parts, norm)
		}
		concat := strings.Join(parts, "|")
		h := hmac.New(sha256.New, []byte(salt))
		h.Write([]byte(concat))
		token := hex.EncodeToString(h.Sum(nil))
		pseudonyms = append(pseudonyms, token)
	}
	return pseudonyms, nil
}
