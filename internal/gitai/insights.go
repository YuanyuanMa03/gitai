package gitai

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

// InsightData represents aggregated insights
type InsightData struct {
	TotalCommits      int     `json:"total_commits"`
	TotalIntents      int     `json:"total_intents"`
	TotalFeedback     int     `json:"total_feedback"`
	AverageScore      float64 `json:"average_score"`
	CategoryBreakdown map[string]int `json:"category_breakdown"`
	TopPrompts        []PromptUsage `json:"top_prompts"`
	Learnings         []string `json:"learnings"`
}

// PromptUsage represents prompt usage statistics
type PromptUsage struct {
	PromptPath string `json:"prompt_path"`
	UsageCount int    `json:"usage_count"`
	AvgRating  float64 `json:"avg_rating"`
}

// TraceResult represents intent trace result
type TraceResult struct {
	IntentID        string   `json:"intent_id"`
	CommitHash      string   `json:"commit_hash"`
	Message         string   `json:"message"`
	Prompts         []string `json:"prompts"`
	AffectedFiles   []string `json:"affected_files"`
	RelatedIntents  []string `json:"related_intents"`
}

// GenerateInsights generates insights from all tracked data
func GenerateInsights(storage *Storage) (*InsightData, error) {
	insights := &InsightData{
		CategoryBreakdown: make(map[string]int),
		TopPrompts:        []PromptUsage{},
		Learnings:         []string{},
	}

	// Get all intents
	intents, err := storage.ListIntents()
	if err != nil {
		return nil, err
	}
	insights.TotalIntents = len(intents)

	// Get all feedback
	feedbackList, _ := storage.ListFeedback()
	insights.TotalFeedback = len(feedbackList)

	// Get feedback stats
	stats, _ := storage.GetFeedbackStats()
	if stats != nil {
		insights.AverageScore = stats.AverageScore
	}

	// Category breakdown
	for _, intent := range intents {
		insights.CategoryBreakdown[intent.Category]++
	}

	// Track commits (from intents)
	commits := make(map[string]bool)
	for _, intent := range intents {
		commits[intent.CommitHash] = true
	}
	insights.TotalCommits = len(commits)

	// Count prompt usage
	promptUsage := make(map[string]int)
	for _, intent := range intents {
		for _, prompt := range intent.PromptHashes {
			promptUsage[prompt]++
		}
	}

	// Convert to top prompts
	for prompt, count := range promptUsage {
		insights.TopPrompts = append(insights.TopPrompts, PromptUsage{
			PromptPath: prompt,
			UsageCount: count,
		})
	}

	// Generate learnings
	insights.Learnings = generateLearnings(insights, intents, feedbackList)

	return insights, nil
}

// generateLearnings generates actionable insights
func generateLearnings(insights *InsightData, intents []*Intent, feedbackList []*Feedback) []string {
	learnings := []string{}

	// Category insights
	maxCategory := ""
	maxCount := 0
	for cat, count := range insights.CategoryBreakdown {
		if count > maxCount {
			maxCount = count
			maxCategory = cat
		}
	}
	if maxCategory != "" {
		learnings = append(learnings, fmt.Sprintf("Most active category: %s (%d intents)", maxCategory, maxCount))
	}

	// Score insights
	if insights.AverageScore >= 75 {
		learnings = append(learnings, "High satisfaction rate - prompts are working well")
	} else if insights.AverageScore < 50 {
		learnings = append(learnings, "Low satisfaction - consider prompt optimization")
	}

	// Volume insights
	if insights.TotalIntents > 50 {
		learnings = append(learnings, "High activity - consider standardizing common patterns")
	}

	return learnings
}

// TraceIntent traces an intent to its effects
func TraceIntent(storage *Storage, intentID string) (*TraceResult, error) {
	// Get intent
	intent, err := storage.GetIntent(intentID)
	if err != nil {
		return nil, err
	}

	result := &TraceResult{
		IntentID:       intentID,
		CommitHash:     intent.CommitHash,
		Message:        intent.Message,
		Prompts:        intent.PromptHashes,
		AffectedFiles:  []string{},
		RelatedIntents: []string{},
	}

	// Get affected files from dependencies
	for _, prompt := range intent.PromptHashes {
		dep, err := storage.GetDependency(prompt)
		if err == nil {
			result.AffectedFiles = append(result.AffectedFiles, dep.Affects...)
		}
	}

	// Find related intents (ints that share prompts)
	allIntents, _ := storage.ListIntents()
	for _, otherIntent := range allIntents {
		if otherIntent.ID == intentID {
			continue
		}

		// Check if they share any prompts
		for _, p1 := range intent.PromptHashes {
			for _, p2 := range otherIntent.PromptHashes {
				if p1 == p2 {
					result.RelatedIntents = append(result.RelatedIntents, otherIntent.ID)
					break
				}
			}
		}
	}

	return result, nil
}

// TraceCodeToIntent traces a file back to intents that affected it
func TraceCodeToIntent(storage *Storage, filePath string) ([]string, error) {
	// Find prompts that affect this file
	prompts, err := storage.FindPromptsAffecting(filePath)
	if err != nil {
		return nil, err
	}

	if len(prompts) == 0 {
		return []string{}, nil
	}

	// Find intents that use these prompts
	intents, _ := storage.ListIntents()
	intentSet := make(map[string]bool)

	for _, prompt := range prompts {
		for _, intent := range intents {
			for _, promptHash := range intent.PromptHashes {
				if promptHash == prompt {
					intentSet[intent.ID] = true
				}
			}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(intentSet))
	for intentID := range intentSet {
		result = append(result, intentID)
	}

	return result, nil
}

// InsightsCmd shows insights from all tracked data
func InsightsCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Generate insights
	insights, err := GenerateInsights(storage)
	if err != nil {
		return fmt.Errorf("failed to generate insights: %w", err)
	}

	// Display
	displayAIInsights(insights)

	return nil
}

// displayAIInsights displays AI-era insights
func displayAIInsights(insights *InsightData) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Println("💡 GitAI Insights")
	fmt.Println()

	fmt.Fprintf(w, "%s\t%d\n", "Total Commits Tracked:", insights.TotalCommits)
	fmt.Fprintf(w, "%s\t%d\n", "Total Intents:", insights.TotalIntents)
	fmt.Fprintf(w, "%s\t%d\n", "Total Feedback:", insights.TotalFeedback)
	if insights.AverageScore > 0 {
		fmt.Fprintf(w, "%s\t%.1f/100\n", "Average Score:", insights.AverageScore)
	}
	w.Flush()

	// Category breakdown
	if len(insights.CategoryBreakdown) > 0 {
		fmt.Println("\n📊 Activity by Category:")
		for cat, count := range insights.CategoryBreakdown {
			fmt.Printf("  %s: %d\n", cat, count)
		}
	}

	// Top prompts
	if len(insights.TopPrompts) > 0 {
		fmt.Println("\n🔝 Most Used Prompts:")
		for i, p := range insights.TopPrompts {
			if i >= 5 {
				break
			}
			fmt.Printf("  %d. %s (%d uses)\n", i+1, p.PromptPath, p.UsageCount)
		}
	}

	// Learnings
	if len(insights.Learnings) > 0 {
		fmt.Println("\n✨ Key Learnings:")
		for _, learning := range insights.Learnings {
			fmt.Printf("  • %s\n", learning)
		}
	}
}

// TraceCmd traces intents or code
func TraceCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	args := c.Args()
	if args.Len() == 0 {
		return fmt.Errorf("usage: gitai trace <intent-id|file-path>")
	}

	target := args.First()

	// Try to trace as intent first
	_, err := storage.GetIntent(target)
	if err == nil {
		return traceIntent(storage, target)
	}

	// Try to trace as file
	return traceFile(storage, target)
}

// traceIntent traces an intent
func traceIntent(storage *Storage, intentID string) error {
	trace, err := TraceIntent(storage, intentID)
	if err != nil {
		return fmt.Errorf("failed to trace intent: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("🔍 Trace: %s\n\n", intentID)

	fmt.Fprintf(w, "%s\t%s\n", "Commit:", trace.CommitHash)
	fmt.Fprintf(w, "%s\t%s\n", "Message:", truncateString(trace.Message, 60))
	w.Flush()

	fmt.Printf("\n📝 Prompts Used (%d):\n", len(trace.Prompts))
	for _, prompt := range trace.Prompts {
		fmt.Printf("  → %s\n", prompt)
	}

	fmt.Printf("\n📁 Affected Files (%d):\n", len(trace.AffectedFiles))
	for _, file := range trace.AffectedFiles {
		fmt.Printf("  → %s\n", file)
	}

	if len(trace.RelatedIntents) > 0 {
		fmt.Printf("\n🔗 Related Intents (%d):\n", len(trace.RelatedIntents))
		for _, ri := range trace.RelatedIntents {
			fmt.Printf("  → %s\n", ri)
		}
	}

	return nil
}

// traceFile traces a file back to intents
func traceFile(storage *Storage, filePath string) error {
	intents, err := TraceCodeToIntent(storage, filePath)
	if err != nil {
		return fmt.Errorf("failed to trace file: %w", err)
	}

	if len(intents) == 0 {
		fmt.Printf("No intents found affecting: %s\n", filePath)
		return nil
	}

	fmt.Printf("🔍 Trace: %s\n\n", filePath)
	fmt.Printf("Found %d intent(s) that affected this file:\n", len(intents))

	for i, intentID := range intents {
		intent, err := storage.GetIntent(intentID)
		if err == nil {
			fmt.Printf("\n%d. %s\n", i+1, intentID)
			fmt.Printf("   Commit: %s\n", intent.CommitHash)
			fmt.Printf("   Message: %s\n", truncateString(intent.Message, 60))
		}
	}

	return nil
}
