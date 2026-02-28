package token

import (
	"strings"
	"unicode"
)

// Estimator handles token estimation for mixed Chinese/English text
type Estimator struct {
	// Approximate token ratios based on common tokenizers
	chineseRatio    float64 // ~0.7 tokens per Chinese character
	englishRatio    float64 // ~0.25 tokens per English character
	numberRatio     float64 // ~0.3 tokens per digit
	punctuationRatio float64 // ~0.5 tokens per punctuation
}

// NewEstimator creates a new token estimator
func NewEstimator() *Estimator {
	return &Estimator{
		chineseRatio:    0.7,
		englishRatio:    0.25,
		numberRatio:     0.3,
		punctuationRatio: 0.5,
	}
}

// Estimate estimates token count for the given text
func (e *Estimator) Estimate(text string) int {
	if text == "" {
		return 0
	}

	chinese := 0
	english := 0
	numbers := 0
	punctuation := 0
	whitespace := 0

	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			// CJK Unified Ideographs
			chinese++
		} else if r >= 0x3400 && r <= 0x4dbf {
			// CJK Extension A
			chinese++
		} else if r >= 0x20000 && r <= 0x2a6df {
			// CJK Extension B
			chinese++
		} else if r >= 0xff00 && r <= 0xffef {
			// Halfwidth and Fullwidth Forms
			chinese++
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			english++
		} else if r >= '0' && r <= '9' {
			numbers++
		} else if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			punctuation++
		} else if unicode.IsSpace(r) {
			whitespace++
		}
	}

	// Calculate tokens
	tokens := float64(chinese)*e.chineseRatio +
		float64(english)*e.englishRatio +
		float64(numbers)*e.numberRatio +
		float64(punctuation)*e.punctuationRatio

	// Add some tokens for word boundaries (spaces separate words in English)
	// Rough approximation: every ~4 characters of text + whitespace = 1 word boundary
	if english > 0 || whitespace > 0 {
		wordBoundaries := (english + whitespace) / 4
		tokens += float64(wordBoundaries) * 0.1
	}

	result := int(tokens)
	if result == 0 && len(text) > 0 {
		// Fallback: rough character count
		result = len(text) / 3
	}

	return result
}

// EstimateLines estimates tokens for multiple lines
func (e *Estimator) EstimateLines(lines []string) int {
	total := 0
	for _, line := range lines {
		total += e.Estimate(line)
	}
	return total
}

// EstimateWords estimates tokens counting by words (more accurate for pure English)
func (e *Estimator) EstimateWords(text string) int {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}

	// Count Chinese characters separately
	chineseChars := 0
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			chineseChars++
		}
	}

	// English: ~0.75 tokens per word
	// Chinese: ~0.7 tokens per character
	tokens := int(float64(len(words))*0.75 + float64(chineseChars)*0.7)

	return tokens
}

// EstimateBySection estimates tokens for structured sections
func (e *Estimator) EstimateBySection(sections map[string]string) map[string]int {
	results := make(map[string]int)
	for name, content := range sections {
		results[name] = e.Estimate(content)
	}
	return results
}

// Default estimator instance
var defaultEstimator = NewEstimator()

// Estimate is a convenience function using the default estimator
func Estimate(text string) int {
	return defaultEstimator.Estimate(text)
}

// EstimateLines is a convenience function for multiple lines
func EstimateLines(lines []string) int {
	return defaultEstimator.EstimateLines(lines)
}

// EstimateBySection is a convenience function for sections
func EstimateBySection(sections map[string]string) map[string]int {
	return defaultEstimator.EstimateBySection(sections)
}
