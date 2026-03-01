package gitai

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/urfave/cli/v2"
)

// Stats represents overall statistics
type Stats struct {
	TotalFiles      int
	TotalTokens     int
	TotalRoles      int
	AvgTokensPerFile float64
	MostCommonRoles []RoleCount
	TokenGrowth     []GrowthData
	CostEstimates   CostEstimate
	Activity        ActivityStats
}

// RoleCount represents a role and its count
type RoleCount struct {
	Role  string
	Count int
	Tokens int
}

// GrowthData represents token growth over time
type GrowthData struct {
	Date   string
	Tokens int
	Files  int
}

// CostEstimate represents cost estimates
type CostEstimate struct {
	USD float64
	CNY float64
}

// ActivityStats represents activity statistics
type ActivityStats struct {
	ActiveFiles    int
	ModifiedToday  int
	ModifiedThisWeek int
	NewThisWeek    int
}

// StatsCmd shows comprehensive statistics
func StatsCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Parse options
	model := c.String("model")
	verbose := c.Bool("verbose")

	// Load config
	config, _ := LoadConfig(root)

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files.")
		return nil
	}

	// Calculate statistics
	stats := calculateStats(tracked, storage, config, model)

	// Display statistics
	return displayStats(stats, verbose)
}

// calculateStats computes all statistics
func calculateStats(tracked []TrackedFile, storage *Storage, config *Config, model string) Stats {
	stats := Stats{
		TotalFiles:  len(tracked),
		TotalTokens: 0,
	}

	// Role statistics
	roleMap := make(map[string]*RoleCount)

	// Token and activity statistics
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	weekAgo := now.AddDate(0, 0, -7)

	for _, f := range tracked {
		stats.TotalTokens += f.Tokens

		// Role statistics
		if roleMap[f.Role] == nil {
			roleMap[f.Role] = &RoleCount{Role: f.Role}
		}
		roleMap[f.Role].Count++
		roleMap[f.Role].Tokens += f.Tokens

		// Activity statistics
		if f.LastModified.After(today) {
			stats.Activity.ModifiedToday++
		}
		if f.LastModified.After(weekAgo) {
			stats.Activity.ModifiedThisWeek++
		}
		if f.AddedAt.After(weekAgo) {
			stats.Activity.NewThisWeek++
		}

		// Active files (modified in last 30 days)
		if f.LastModified.After(now.AddDate(0, 0, -30)) {
			stats.Activity.ActiveFiles++
		}
	}

	// Convert role map to slice
	stats.TotalRoles = len(roleMap)
	for _, rc := range roleMap {
		stats.MostCommonRoles = append(stats.MostCommonRoles, *rc)
	}

	// Sort by count
	sort.Slice(stats.MostCommonRoles, func(i, j int) bool {
		return stats.MostCommonRoles[i].Count > stats.MostCommonRoles[j].Count
	})

	// Limit to top 10
	if len(stats.MostCommonRoles) > 10 {
		stats.MostCommonRoles = stats.MostCommonRoles[:10]
	}

	// Calculate average
	stats.AvgTokensPerFile = float64(stats.TotalTokens) / float64(stats.TotalFiles)

	// Cost estimates
	if config != nil {
		pricing, err := config.GetModelPricing(model)
		if err == nil {
			usd, cny := calculateCostWithPricing(stats.TotalTokens, pricing)
			stats.CostEstimates.USD = usd
			stats.CostEstimates.CNY = cny
		}
	}

	return stats
}

// displayStats shows statistics
func displayStats(stats Stats, verbose bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║        GitAI Statistics Dashboard        ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// Overview
	fmt.Println("📊 Overview")
	fmt.Fprintf(w, "%s\t%d\n", "Total Files:", stats.TotalFiles)
	fmt.Fprintf(w, "%s\t%d\n", "Total Tokens:", stats.TotalTokens)
	fmt.Fprintf(w, "%s\t%s\n", "Total Roles:", fmt.Sprintf("%d", stats.TotalRoles))
	fmt.Fprintf(w, "%s\t%.0f\n", "Avg Tokens/File:", stats.AvgTokensPerFile)
	w.Flush()
	fmt.Println()

	// Cost Estimates
	fmt.Println("💰 Cost Estimates")
	fmt.Fprintf(w, "%s\t%s\n", "USD:", FormatCurrency(stats.CostEstimates.USD, "USD"))
	fmt.Fprintf(w, "%s\t%s\n", "CNY:", FormatCurrency(stats.CostEstimates.CNY, "CNY"))
	w.Flush()
	fmt.Println()

	// Top Roles
	fmt.Println("🎭 Top Roles")
	fmt.Fprintf(w, "%s\t%s\t%s\n", "Role", "Files", "Tokens")
	fmt.Fprintf(w, "%s\t%s\t%s\n", "----", "-----", "------")
	for _, rc := range stats.MostCommonRoles {
		fmt.Fprintf(w, "%s\t%d\t%d\n", rc.Role, rc.Count, rc.Tokens)
	}
	w.Flush()
	fmt.Println()

	// Activity
	fmt.Println("⚡ Activity (Last 7 Days)")
	fmt.Fprintf(w, "%s\t%d\n", "Modified Today:", stats.Activity.ModifiedToday)
	fmt.Fprintf(w, "%s\t%d\n", "Modified This Week:", stats.Activity.ModifiedThisWeek)
	fmt.Fprintf(w, "%s\t%d\n", "New This Week:", stats.Activity.NewThisWeek)
	fmt.Fprintf(w, "%s\t%d\n", "Active Files (30d):", stats.Activity.ActiveFiles)
	w.Flush()
	fmt.Println()

	// Insights
	displayInsights(stats)

	return nil
}

// displayInsights shows derived insights
func displayInsights(stats Stats) {
	fmt.Println("💡 Insights")

	insights := []string{}

	// File count insights
	if stats.TotalFiles < 5 {
		insights = append(insights, "📝 Start tracking more prompt files to get better insights")
	} else if stats.TotalFiles > 50 {
		insights = append(insights, "📚 Large prompt library - consider using tags and templates")
	}

	// Token efficiency
	if stats.AvgTokensPerFile > 500 {
		insights = append(insights, "⚠️  Average prompt is quite long - consider splitting complex prompts")
	} else if stats.AvgTokensPerFile < 100 {
		insights = append(insights, "✅ Prompts are concise and focused")
	}

	// Role diversity
	if stats.TotalRoles == 1 {
		insights = append(insights, "🎯 Consider diversifying roles for different use cases")
	} else if stats.TotalRoles > 10 {
		insights = append(insights, "🌟 Great role diversity - consider standardizing common patterns")
	}

	// Activity
	if stats.Activity.NewThisWeek > 3 {
		insights = append(insights, "🚀 High activity - lots of new prompts this week")
	} else if stats.Activity.ModifiedThisWeek == 0 {
		insights = append(insights, "💤 No activity this week - time to review and update prompts")
	}

	// Cost
	if stats.CostEstimates.USD > 1.0 {
		insights = append(insights, "💵 Significant token investment - consider optimization")
	}

	for _, insight := range insights {
		fmt.Println(insight)
	}
}

// TrendCmd shows token trends over time
func TrendCmd(c *cli.Context) error {
	// Get git root
	repo, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Parse options
	days := c.Int("days")
	if days < 1 {
		days = 30
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

	// Get commit history with limit
	limit := days * 10 // Assume max 10 commits per day
	commits, err := getCommitHistory(repo, limit)
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	// Calculate trend data
	trendData := calculateTrendData(tracked, commits, days)

	// Display trend
	displayTrend(trendData, days)

	return nil
}

// TrendData represents a single data point in the trend
type TrendData struct {
	Date  string
	Files int
	Tokens int
}

// calculateTrendData computes trend data over time
func calculateTrendData(tracked []TrackedFile, commits []*object.Commit, days int) []TrendData {
	now := time.Now()
	dataPoints := make([]TrendData, days)

	// Initialize data points
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		dataPoints[i] = TrendData{
			Date: date.Format("2006-01-02"),
		}
	}

	// Group tracked files by date
	for _, f := range tracked {
		daysSinceMod := int(now.Sub(f.LastModified).Hours() / 24)
		if daysSinceMod >= 0 && daysSinceMod < days {
			dataPoints[daysSinceMod].Files++
			dataPoints[daysSinceMod].Tokens += f.Tokens
		}
	}

	return dataPoints
}

// displayTrend shows trend visualization
func displayTrend(data []TrendData, days int) {
	// Reverse for chronological order
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	fmt.Println("📈 Token Growth Trend")
	fmt.Println()

	// Find max for scaling
	maxTokens := 0
	for _, d := range data {
		if d.Tokens > maxTokens {
			maxTokens = d.Tokens
		}
	}

	// Display chart
	barWidth := 40
	for _, d := range data {
		if d.Tokens == 0 {
			continue
		}

		// Calculate bar length
		barLen := int(float64(d.Tokens) / float64(maxTokens) * float64(barWidth))
		bar := strings.Repeat("█", barLen)

		fmt.Printf("%s │%s %d tokens (%d files)\n",
			d.Date, bar, d.Tokens, d.Files)
	}

	fmt.Println()
	fmt.Printf("Showing last %d days\n", days)
}

// CompareCmd compares statistics between two points
func CompareCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Parse options
	commitA := c.String("a")
	commitB := c.String("b")

	// Get current tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	fmt.Println("📊 Comparison")
	fmt.Println()
	fmt.Println("Current state:")
	fmt.Printf("  Files: %d\n", len(tracked))

	totalTokens := 0
	for _, f := range tracked {
		totalTokens += f.Tokens
	}
	fmt.Printf("  Tokens: %d\n", totalTokens)

	// TODO: Implement historical comparison when commit refs are provided
	if commitA != "" || commitB != "" {
		fmt.Println("\nHistorical comparison not yet implemented")
	}

	return nil
}
