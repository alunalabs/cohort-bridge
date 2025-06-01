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
	var pseudonyms []string
	for _, record := range records {
		var parts []string
		for _, field := range fields {
			val := record[field]
			var norm string
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
