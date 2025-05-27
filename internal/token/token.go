package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// GenerateToken normalizes fields and returns the HMAC pseudonym.
func GenerateToken(fields []string, record map[string]interface{}, salt string) string {
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
	return hex.EncodeToString(h.Sum(nil))
}
