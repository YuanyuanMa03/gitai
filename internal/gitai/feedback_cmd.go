package gitai

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v2"
)

// FeedbackCmd handles feedback commands
func FeedbackCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	args := c.Args()

	if args.Len() == 0 {
		// Show feedback summary
		return ShowFeedbackSummaryCmd(c)
	}

	action := args.Get(0)
	switch action {
	case "good", "bad", "neutral":
		return AddFeedbackCmd(c, storage, action)
	case "list", "ls":
		return ListFeedbackCmd(c, storage)
	case "stats":
		return ShowFeedbackStatsCmd(c, storage)
	default:
		return fmt.Errorf("unknown feedback action: %s\nUsage: gitai feedback [good|bad|neutral|list|stats] [commit]", action)
	}
}

// AddFeedbackCmd adds feedback for a commit
func AddFeedbackCmd(c *cli.Context, storage *Storage, rating string) error {
	args := c.Args()

	// Get commit hash
	var commitHash string
	if args.Len() > 1 {
		commitHash = args.Get(1)
	} else {
		// Default to HEAD
		var err error
		commitHash, err = getLatestCommitHash()
		if err != nil {
			return fmt.Errorf("failed to get HEAD commit: %w", err)
		}
	}

	// Get reason
	reason := c.String("reason")
	if reason == "" {
		// Default reasons based on rating
		defaultReasons := map[string]string{
			"good":    "Produced excellent results",
			"bad":     "Had issues or errors",
			"neutral": "Acceptable but could be improved",
		}
		reason = defaultReasons[rating]
	}

	// Get associated intent if exists
	intent, _ := storage.GetIntentByCommit(commitHash)

	// Create feedback
	feedback := &Feedback{
		ID:         fmt.Sprintf("fb-%d", time.Now().Unix()),
		CommitHash: commitHash,
		Rating:     rating,
		Reason:     reason,
		Timestamp:  time.Now(),
	}

	if intent != nil {
		feedback.IntentID = intent.ID
	}

	// Save feedback
	if err := storage.SaveFeedback(feedback); err != nil {
		return fmt.Errorf("failed to save feedback: %w", err)
	}

	fmt.Printf("✓ Feedback recorded: %s for commit %s\n", rating, commitHash[:7])
	if reason != "" {
		fmt.Printf("  Reason: %s\n", reason)
	}

	return nil
}

// ListFeedbackCmd lists all feedback
func ListFeedbackCmd(c *cli.Context, storage *Storage) error {
	feedbackList, err := storage.ListFeedback()
	if err != nil {
		return fmt.Errorf("failed to load feedback: %w", err)
	}

	if len(feedbackList) == 0 {
		fmt.Println("No feedback recorded yet.")
		fmt.Println("\nUsage: gitai feedback [good|bad|neutral] [commit]")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Feedback History (%d entries)\n\n", len(feedbackList))

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "Rating", "Commit", "Reason", "Intent", "Time")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "------", "------", "-------", "-------", "----")

	for _, fb := range feedbackList {
		shortCommit := fb.CommitHash
		if len(shortCommit) > 7 {
			shortCommit = shortCommit[:7]
		}

		shortReason := truncateString(fb.Reason, 30)
		shortIntent := fb.IntentID
		if shortIntent == "" {
			shortIntent = "-"
		} else if len(shortIntent) > 10 {
			shortIntent = shortIntent[:7] + ".."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			fb.Rating, shortCommit, shortReason, shortIntent, fb.Timestamp.Format("2006-01-02"))
	}

	w.Flush()

	return nil
}

// ShowFeedbackStatsCmd shows feedback statistics
func ShowFeedbackStatsCmd(c *cli.Context, storage *Storage) error {
	stats, err := storage.GetFeedbackStats()
	if err != nil {
		return fmt.Errorf("failed to get feedback stats: %w", err)
	}

	if stats.Total == 0 {
		fmt.Println("No feedback data yet.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Println("Feedback Statistics")
	fmt.Println()

	fmt.Fprintf(w, "%s\t%d\n", "Total Feedback:", stats.Total)
	fmt.Fprintf(w, "%s\t%d (%.1f%%)\n", "Good:", stats.Good, float64(stats.Good)/float64(stats.Total)*100)
	fmt.Fprintf(w, "%s\t%d (%.1f%%)\n", "Bad:", stats.Bad, float64(stats.Bad)/float64(stats.Total)*100)
	fmt.Fprintf(w, "%s\t%d (%.1f%%)\n", "Neutral:", stats.Neutral, float64(stats.Neutral)/float64(stats.Total)*100)
	w.Flush()

	fmt.Println()

	// Visual score indicator
	indicator := getFeedbackIndicator(stats.AverageScore)
	fmt.Printf("Average Score: %.1f/100 %s\n", stats.AverageScore, indicator)

	return nil
}

// ShowFeedbackSummaryCmd shows a quick summary
func ShowFeedbackSummaryCmd(c *cli.Context) error {
	_, root := MustGetGitRepo()
	storage := NewStorage(root)

	stats, err := storage.GetFeedbackStats()
	if err != nil {
		return err
	}

	if stats.Total == 0 {
		fmt.Println("No feedback recorded yet.")
		fmt.Println("\nRecord feedback with: gitai feedback [good|bad|neutral] [commit]")
		return nil
	}

	fmt.Printf("Feedback: %d total | ", stats.Total)
	fmt.Printf("Good: %d | Bad: %d | Neutral: %d\n", stats.Good, stats.Bad, stats.Neutral)
	fmt.Printf("Score: %.1f/100\n", stats.AverageScore)

	return nil
}

// getFeedbackIndicator returns a visual indicator for feedback score
func getFeedbackIndicator(score float64) string {
	if score >= 75 {
		return "✓ Excellent"
	} else if score >= 50 {
		return "↑ Good"
	} else if score >= 25 {
		return "~ Fair"
	} else {
		return "↓ Poor"
	}
}
