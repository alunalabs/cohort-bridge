package crypto

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// NormalizationMethod represents supported normalization methods
type NormalizationMethod string

const (
	NormName   NormalizationMethod = "name"
	NormDate   NormalizationMethod = "date"
	NormGender NormalizationMethod = "gender"
	NormZip    NormalizationMethod = "zip"
)

// FieldNormalization represents a field and its normalization method
type FieldNormalization struct {
	Method NormalizationMethod
	Field  string
}

// ParseNormalizationConfig parses normalization config strings into structured format
// Expects format like "name:FIRST" or "date:DATE_OF_BIRTH"
func ParseNormalizationConfig(normSpecs []string) map[string]NormalizationMethod {
	normMap := make(map[string]NormalizationMethod)

	for _, spec := range normSpecs {
		// Split on colon, if no colon found, leave unnormalized
		parts := strings.Split(spec, ":")
		if len(parts) != 2 {
			continue
		}

		method := strings.ToLower(strings.TrimSpace(parts[0]))
		field := strings.TrimSpace(parts[1])

		// Check if method is supported
		switch method {
		case "name":
			normMap[field] = NormName
		case "date":
			normMap[field] = NormDate
		case "gender":
			normMap[field] = NormGender
		case "zip":
			normMap[field] = NormZip
		default:
			// Unsupported method, skip normalization for this field
			continue
		}
	}

	return normMap
}

// NormalizeName standardizes name fields
func NormalizeName(value string) string {
	if value == "" {
		return ""
	}

	// Convert to lowercase and trim
	normalized := strings.ToLower(strings.TrimSpace(value))

	// Remove common punctuation and extra spaces
	reg := regexp.MustCompile(`[^a-z\s]`)
	normalized = reg.ReplaceAllString(normalized, "")

	// Normalize multiple spaces to single space
	spaceReg := regexp.MustCompile(`\s+`)
	normalized = spaceReg.ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

// NormalizeDate standardizes date fields to YYYY-MM-DD format
func NormalizeDate(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case time.Time:
		return v.Format("2006-01-02")
	case string:
		if v == "" {
			return ""
		}

		// Try to parse common date formats
		dateFormats := []string{
			"2006-01-02",
			"01/02/2006",
			"1/2/2006",
			"01-02-2006",
			"1-2-2006",
			"2006/01/02",
			"2006/1/2",
			"01/02/06",
			"1/2/06",
		}

		for _, format := range dateFormats {
			if t, err := time.Parse(format, v); err == nil {
				return t.Format("2006-01-02")
			}
		}

		// If no format matches, return trimmed lowercase
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return strings.ToLower(strings.TrimSpace(fmt.Sprint(v)))
	}
}

// NormalizeGender standardizes gender fields
func NormalizeGender(value string) string {
	if value == "" {
		return ""
	}

	normalized := strings.ToLower(strings.TrimSpace(value))

	// Standardize common gender representations
	switch normalized {
	case "m", "male", "man", "boy":
		return "m"
	case "f", "female", "woman", "girl":
		return "f"
	case "nb", "nonbinary", "non-binary", "non binary", "enby":
		return "nb"
	case "o", "other":
		return "o"
	case "u", "unknown", "unspecified", "prefer not to say":
		return "u"
	default:
		// Return first character if it's a valid gender initial
		if len(normalized) > 0 {
			first := string(normalized[0])
			if first == "m" || first == "f" || first == "o" || first == "u" {
				return first
			}
		}
		return "u" // Default to unknown
	}
}

// NormalizeZip standardizes ZIP code fields
func NormalizeZip(value string) string {
	if value == "" {
		return ""
	}

	// Remove all non-numeric characters and spaces
	reg := regexp.MustCompile(`[^0-9]`)
	normalized := reg.ReplaceAllString(value, "")

	// Ensure it's a reasonable length (5 or 9 digits for US ZIP codes)
	if len(normalized) >= 5 {
		return normalized[:5] // Take first 5 digits for consistency
	}

	return normalized
}

// NormalizeField applies the appropriate normalization based on the method
func NormalizeField(value interface{}, method NormalizationMethod) string {
	switch method {
	case NormName:
		return NormalizeName(fmt.Sprint(value))
	case NormDate:
		return NormalizeDate(value)
	case NormGender:
		return NormalizeGender(fmt.Sprint(value))
	case NormZip:
		return NormalizeZip(fmt.Sprint(value))
	default:
		// No normalization method specified, apply basic normalization
		if value == nil {
			return ""
		}
		switch v := value.(type) {
		case string:
			return strings.ToLower(strings.TrimSpace(v))
		case time.Time:
			return v.Format("2006-01-02")
		default:
			return strings.ToLower(strings.TrimSpace(fmt.Sprint(v)))
		}
	}
}
