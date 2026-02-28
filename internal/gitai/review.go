package gitai

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/mayuanyuan/gitai/pkg/token"
	"github.com/urfave/cli/v2"
)

// ReviewScore holds review metrics for a prompt
type ReviewScore struct {
	File            string
	Role            string
	TotalTokens     int
	ClarityScore   float64 // 0-100
	ConcisenessScore float64 // 0-100
	EfficiencyScore  float64 // 0-100
	OverallScore    float64 // 0-100
	Improvements    []string
}

// ReviewCmd reviews tracked prompt files for quality
func ReviewCmd(c *cli.Context) error {
	// Get git root
	repo, root := MustGetGitRepo()
	_ = repo

	// Create storage
	storage := NewStorage(root)

	// Check if .gitai/ is initialized
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files to review.")
		return nil
	}

	// Parse options
	detailed := c.Bool("detailed")
	minScore := c.Float64("min-score")

	// Analyze each file
	scores := make([]ReviewScore, 0, len(tracked))
	for _, f := range tracked {
		score := analyzePrompt(f)
		if score.OverallScore >= minScore {
			scores = append(scores, score)
		}
	}

	// Sort by overall score (ascending - worst first)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].OverallScore < scores[j].OverallScore
	})

	// Display results
	return displayReviewScores(scores, detailed)
}

// analyzePrompt analyzes a prompt file for quality metrics
func analyzePrompt(f TrackedFile) ReviewScore {
	score := ReviewScore{
		File:         f.Path,
		Role:         f.Role,
		TotalTokens:  f.Tokens,
		Improvements: []string{},
	}

	// Read file content
	absPath, err := GetAbsolutePath(f.Path)
	if err != nil {
		score.OverallScore = 50
		return score
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		score.OverallScore = 50
		return score
	}

	contentStr := string(content)

	// Parse the prompt
	parsed := token.Parse(contentStr)

	// Calculate clarity score
	score.ClarityScore = calculateClarity(parsed, &score.Improvements)

	// Calculate conciseness score
	score.ConcisenessScore = calculateConciseness(parsed, f.Tokens, &score.Improvements)

	// Calculate efficiency score
	score.EfficiencyScore = calculateEfficiency(parsed, &score.Improvements)

	// Overall score (weighted average)
	score.OverallScore = (score.ClarityScore*0.4 +
		score.ConcisenessScore*0.3 +
		score.EfficiencyScore*0.3)

	return score
}

// calculateClarity measures how clear and specific the prompt is
func calculateClarity(parsed *token.ParsedPrompt, improvements *[]string) float64 {
	score := 100.0 // Start with perfect score

	// Check for role
	if parsed.Role == "" || parsed.Role == "default" {
		score -= 20
		*improvements = append(*improvements, "Add a specific role to clarify AI's persona")
	}

	// Check for task section
	if parsed.Task == "" {
		score -= 30
		*improvements = append(*improvements, "Add a ## Task: section to define what the AI should do")
	} else if len(parsed.Task) < 50 {
		score -= 10
		*improvements = append(*improvements, "Task description is too brief (consider adding more details)")
	}

	// Check for constraints
	if parsed.Constraints == "" {
		score -= 15
		*improvements = append(*improvements, "Add ## Constraints: to define rules and limitations")
	}

	// Check for examples
	if parsed.Examples == "" {
		score -= 10
		*improvements = append(*improvements, "Add ## Examples: to show expected input/output")
	}

	// Check for vague language
	vaguePhrases := []string{
		"do something", "do some stuff", "help me",
		"whatever", "things", "stuff", "etc",
	}

	taskLower := strings.ToLower(parsed.Task)
	for _, phrase := range vaguePhrases {
		if strings.Contains(taskLower, phrase) {
			score -= 5
			*improvements = append(*improvements, "Avoid vague language like '"+phrase+"'")
		}
	}

	// Bonus: structured sections
	if parsed.Input != "" {
		score += 5
	}
	if parsed.Output != "" {
		score += 5
	}

	// Ensure score is in range [0, 100]
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateConciseness measures token efficiency
func calculateConciseness(parsed *token.ParsedPrompt, totalTokens int, improvements *[]string) float64 {
	score := 100.0

	// Check for redundancy
	redundancyCount := 0
	lines := strings.Split(parsed.Task, "\n")
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if seen[lower] {
			redundancyCount++
		}
		seen[lower] = true
	}

	if redundancyCount > 0 {
		penalty := redundancyCount * 5
		if penalty > 30 {
			penalty = 30
		}
		score -= float64(penalty)
		*improvements = append(*improvements, fmt.Sprintf("Remove %d redundant line(s)", redundancyCount))
	}

	// Check token efficiency
	if totalTokens > 500 {
		score -= 10
		*improvements = append(*improvements, "Consider splitting into smaller, focused prompts")
	} else if totalTokens > 300 {
		score -= 5
	}

	// Check for over-explaining
	if len(parsed.Task) > 500 {
		score -= 10
		*improvements = append(*improvements, "Task section is very long, consider summarizing")
	}

	// Bonus: brief but complete
	if totalTokens > 50 && totalTokens < 200 && parsed.Constraints != "" && parsed.Examples != "" {
		score += 10
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// calculateEfficiency measures structure and completeness
func calculateEfficiency(parsed *token.ParsedPrompt, improvements *[]string) float64 {
	score := 100.0

	// Check for all key sections
	hasRole := parsed.Role != "" && parsed.Role != "default"
	hasTask := parsed.Task != ""
	hasConstraints := parsed.Constraints != ""
	hasExamples := parsed.Examples != ""

	sectionCount := 0
	if hasRole {
		sectionCount++
	}
	if hasTask {
		sectionCount++
	}
	if hasConstraints {
		sectionCount++
	}
	if hasExamples {
		sectionCount++
	}

	// Penalize missing sections
	if !hasRole {
		score -= 15
	}
	if !hasTask {
		score -= 30
	}
	if !hasConstraints {
		score -= 10
	}
	if !hasExamples {
		score -= 10
	}

	// Bonus for having input/output specification
	if parsed.Input != "" {
		score += 5
	}
	if parsed.Output != "" {
		score += 5
	}

	// Check for proper formatting
	hasProperStructure := strings.Contains(parsed.Task, "-") ||
		strings.Contains(parsed.Constraints, "-")

	if hasProperStructure {
		score += 10
	} else {
		*improvements = append(*improvements, "Use bullet points (-) for better readability")
	}

	// Check for consistent tense
	consistency := checkConsistency(parsed)
	if !consistency {
		score -= 5
		*improvements = append(*improvements, "Maintain consistent tense throughout")
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// checkConsistency checks for consistent tense and style
func checkConsistency(parsed *token.ParsedPrompt) bool {
	// Simple heuristic: check for tense consistency
	// Count imperatives vs statements
	task := parsed.Task
	if task == "" {
		return true
	}

	words := strings.Fields(task)
	imperativeCount := 0
	statementCount := 0

	imperativeVerbs := []string{"write", "create", "make", "add", "use", "follow", "implement", "generate", "build"}

	for _, word := range words {
		lower := strings.ToLower(word)
		for _, verb := range imperativeVerbs {
			if strings.HasPrefix(lower, verb) {
				imperativeCount++
				break
			}
		}
		if strings.HasSuffix(word, "ing") || strings.HasSuffix(word, "tion") {
			statementCount++
		}
	}

	// Good if mostly one style
	total := imperativeCount + statementCount
	if total == 0 {
		return true
	}

	dominant := imperativeCount
	if statementCount > imperativeCount {
		dominant = statementCount
	}

	ratio := float64(dominant) / float64(total)
	return ratio > 0.7 // 70% consistency is good
}

// displayReviewScores shows the review results
func displayReviewScores(scores []ReviewScore, detailed bool) error {
	if len(scores) == 0 {
		fmt.Println("All files passed the review threshold!")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Prompt Review Results (%d files)\n\n", len(scores))

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "File", "Role", "Tokens", "Clarity", "Concise", "Efficient")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "----", "----", "------", "-------", "-------", "---------")

	for i := range scores {
		s := scores[i]
		shortFile := s.File
		if len(shortFile) > 15 {
			shortFile = shortFile[:12] + ".."
		}

		shortRole := s.Role
		if len(shortRole) > 12 {
			shortRole = shortRole[:9] + ".."
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%.0f\t%.0f\t%.0f\n",
			shortFile,
			shortRole,
			s.TotalTokens,
			s.ClarityScore,
			s.ConcisenessScore,
			s.EfficiencyScore,
		)

		// Print overall score indicator
		indicator := getScoreIndicator(s.OverallScore)
		fmt.Fprintf(w, "  %s: %.0f/100\n", indicator, s.OverallScore)

		// Print improvements if detailed
		if detailed && len(s.Improvements) > 0 {
			fmt.Fprintf(w, "  Suggestions:\n")
			for i, imp := range s.Improvements {
				if i >= 3 {
					fmt.Fprintf(w, "    ... %d more suggestions\n", len(s.Improvements)-i)
					break
				}
				fmt.Fprintf(w, "    - %s\n", imp)
			}
		}
	}

	w.Flush()

	// Overall statistics
	avgClarity := 0.0
	avgConciseness := 0.0
	avgEfficiency := 0.0
	avgOverall := 0.0

	for _, s := range scores {
		avgClarity += s.ClarityScore
		avgConciseness += s.ConcisenessScore
		avgEfficiency += s.EfficiencyScore
		avgOverall += s.OverallScore
	}

	count := float64(len(scores))
	avgClarity /= count
	avgConciseness /= count
	avgEfficiency /= count
	avgOverall /= count

	fmt.Printf("\nAverage Scores:\n")
	fmt.Printf("  Clarity:    %.1f/100\n", avgClarity)
	fmt.Printf("  Conciseness: %.1f/100\n", avgConciseness)
	fmt.Printf("  Efficiency: %.1f/100\n", avgEfficiency)
	fmt.Printf("  Overall:    %.1f/100\n", avgOverall)

	return nil
}

// getScoreIndicator returns a visual indicator for the score
func getScoreIndicator(score float64) string {
	if score >= 80 {
		return "✓ Excellent"
	} else if score >= 60 {
		return "↑ Good"
	} else if score >= 40 {
		return "~ Fair"
	} else {
		return "↓ Poor"
	}
}

// QualityCmd checks overall quality of all prompts
func QualityCmd(c *cli.Context) error {
	// Get git root
	repo, root := MustGetGitRepo()
	_ = repo

	// Create storage
	storage := NewStorage(root)

	// Check if .gitai/ is initialized
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files.")
		return nil
	}

	// Quick quality check
	excellent := 0
	good := 0
	fair := 0
	poor := 0

	for _, f := range tracked {
		score := analyzePrompt(f)
		if score.OverallScore >= 80 {
			excellent++
		} else if score.OverallScore >= 60 {
			good++
		} else if score.OverallScore >= 40 {
			fair++
		} else {
			poor++
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Quality Overview (%d files)\n\n", len(tracked))

	fmt.Fprintf(w, "%s\t%s\n", "Rating", "Count")
	fmt.Fprintf(w, "%s\t%s\n", "------", "-----")

	if excellent > 0 {
		fmt.Fprintf(w, "%s\t%d\n", "✓ Excellent (80+)", excellent)
	}
	if good > 0 {
		fmt.Fprintf(w, "%s\t%d\n", "↑ Good (60+)", good)
	}
	if fair > 0 {
		fmt.Fprintf(w, "%s\t%d\n", "~ Fair (40+)", fair)
	}
	if poor > 0 {
		fmt.Fprintf(w, "%s\t%d\n", "↓ Poor (<40)", poor)
	}

	w.Flush()

	// Recommendations
	if poor > 0 || fair > 0 {
		fmt.Printf("\nRecommendations:\n")
		if poor > 0 {
			fmt.Printf("  - Review %d file(s) marked as 'Poor'\n", poor)
		}
		if fair > 0 {
			fmt.Printf("  - Consider improving %d file(s) marked as 'Fair'\n", fair)
		}
		fmt.Printf("  - Run 'gitai review --detailed' for specific suggestions\n")
	}

	return nil
}

// EstimateTokensQuick provides a quick token estimate without parsing
func EstimateTokensQuick(text string) int {
	// Count Chinese characters and English words
	chineseCount := 0
	englishCount := 0
	whiteCount := 0

	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff {
			chineseCount++
		} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			englishCount++
		} else if unicode.IsSpace(r) {
			whiteCount++
		}
	}

	// Estimate tokens
	chineseTokens := int(float64(chineseCount) * 0.7)
	englishTokens := int(float64(englishCount) * 0.25)
	totalTokens := chineseTokens + englishTokens

	if totalTokens == 0 && len(text) > 0 {
		totalTokens = len(text) / 3
	}

	return totalTokens
}

// GetTokenDensity returns tokens per section
func GetTokenDensity(parsed *token.ParsedPrompt) map[string]int {
	density := make(map[string]int)

	if parsed.Task != "" {
		density["Task"] = EstimateTokensQuick(parsed.Task)
	}
	if parsed.Constraints != "" {
		density["Constraints"] = EstimateTokensQuick(parsed.Constraints)
	}
	if parsed.Examples != "" {
		density["Examples"] = EstimateTokensQuick(parsed.Examples)
	}
	if parsed.Input != "" {
		density["Input"] = EstimateTokensQuick(parsed.Input)
	}
	if parsed.Output != "" {
		density["Output"] = EstimateTokensQuick(parsed.Output)
	}

	return density
}
