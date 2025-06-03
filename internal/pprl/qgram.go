// qgram.go
// Package pprl provides q-gram functionality for improved string matching.
package pprl

import (
	"strings"
)

// QGram represents a q-gram (substring of length q) with its frequency
type QGram struct {
	Value     string
	Frequency int
}

// QGramSet represents a set of q-grams extracted from a string
type QGramSet struct {
	Q       int            // q-gram length
	Grams   map[string]int // map of q-gram to frequency
	Padding string         // padding character for start/end
}

// NewQGramSet creates a new QGramSet with the specified q-gram length
func NewQGramSet(q int, padding string) *QGramSet {
	if q < 1 {
		q = 2 // default to bigrams if invalid q
	}
	if padding == "" {
		padding = "#" // default padding character
	}
	return &QGramSet{
		Q:       q,
		Grams:   make(map[string]int),
		Padding: padding,
	}
}

// ExtractQGrams extracts all q-grams from a string, including padding
func (qs *QGramSet) ExtractQGrams(s string) {
	// Clear existing grams
	qs.Grams = make(map[string]int)

	// Add padding to start and end
	padded := strings.Repeat(qs.Padding, qs.Q-1) + s + strings.Repeat(qs.Padding, qs.Q-1)

	// Extract q-grams
	for i := 0; i <= len(padded)-qs.Q; i++ {
		gram := padded[i : i+qs.Q]
		qs.Grams[gram]++
	}
}

// GetQGramFrequency returns the frequency of a specific q-gram
func (qs *QGramSet) GetQGramFrequency(gram string) int {
	return qs.Grams[gram]
}

// GetQGramSimilarity calculates the Jaccard similarity between two q-gram sets
func (qs *QGramSet) GetQGramSimilarity(other *QGramSet) float64 {
	if qs.Q != other.Q {
		return 0.0 // different q-gram lengths are incomparable
	}

	// Calculate intersection and union sizes
	intersection := 0
	union := 0

	// Count intersection and union using the first set as base
	for gram, freq1 := range qs.Grams {
		freq2 := other.Grams[gram]
		intersection += min(freq1, freq2)
		union += max(freq1, freq2)
	}

	// Add remaining grams from second set to union
	for gram, freq2 := range other.Grams {
		if _, exists := qs.Grams[gram]; !exists {
			union += freq2
		}
	}

	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

// GetQGramDistance calculates the edit distance between two q-gram sets
func (qs *QGramSet) GetQGramDistance(other *QGramSet) int {
	if qs.Q != other.Q {
		return -1 // different q-gram lengths are incomparable
	}

	// Calculate symmetric difference
	distance := 0

	// Count differences using first set as base
	for gram, freq1 := range qs.Grams {
		freq2 := other.Grams[gram]
		distance += abs(freq1 - freq2)
	}

	// Add remaining grams from second set
	for gram, freq2 := range other.Grams {
		if _, exists := qs.Grams[gram]; !exists {
			distance += freq2
		}
	}

	return distance
}

// NormalizeString prepares a string for q-gram extraction by:
// 1. Converting to lowercase
// 2. Removing non-alphanumeric characters
// 3. Normalizing whitespace
func NormalizeString(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove non-alphanumeric characters and normalize whitespace
	var result strings.Builder
	lastWasSpace := true

	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			result.WriteRune(r)
			lastWasSpace = false
		} else if !lastWasSpace {
			result.WriteRune(' ')
			lastWasSpace = true
		}
	}

	return strings.TrimSpace(result.String())
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
