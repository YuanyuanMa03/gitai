package gitai

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/urfave/cli/v2"
)

// UserStats holds cost statistics per user
type UserStats struct {
	User      string
	Files     int
	Tokens    int
	Commits   int
	USD       float64
	CNY       float64
}

// BranchStats holds cost statistics per branch
type BranchStats struct {
	Branch   string
	Files    int
	Tokens   int
	Commits  int
	USD      float64
	CNY      float64
}

// CostByUserCmd shows cost breakdown by user
func CostByUserCmd(c *cli.Context) error {
	// Get git root
	repo, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	model := c.String("model")
	limit := c.Int("limit")

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files.")
		return nil
	}

	// Get commit history
	commits, err := getCommitHistory(repo, limit)
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	// Calculate stats per user
	userStats, err := calculateUserStats(repo, root, tracked, commits, model)
	if err != nil {
		return fmt.Errorf("failed to calculate user stats: %w", err)
	}

	// Display
	return showUserStats(userStats, model)
}

// CostByBranchCmd shows cost breakdown by branch
func CostByBranchCmd(c *cli.Context) error {
	// Get git root
	repo, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	model := c.String("model")

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files.")
		return nil
	}

	// Get all branches
	branches, err := getLocalBranches(repo)
	if err != nil {
		return fmt.Errorf("failed to get branches: %w", err)
	}

	// Calculate stats per branch
	branchStats, err := calculateBranchStats(repo, root, tracked, branches, model)
	if err != nil {
		return fmt.Errorf("failed to calculate branch stats: %w", err)
	}

	// Display
	return showBranchStats(branchStats, model)
}

// calculateUserStats computes token statistics per user
func calculateUserStats(repo *git.Repository, root string, tracked []TrackedFile, commits []*object.Commit, model string) ([]UserStats, error) {
	costMgr := NewCostManager()

	// Map: user -> UserStats
	statsMap := make(map[string]*UserStats)

	for _, commit := range commits {
		user := commit.Author.Name
		if user == "" {
			user = "Unknown"
		}

		if statsMap[user] == nil {
			statsMap[user] = &UserStats{User: user}
		}
		statsMap[user].Commits++

		// Get file changes in this commit
		changes, err := getPromptChangesInCommit(repo, root, tracked, commit)
		if err != nil {
			continue
		}

		for _, change := range changes {
			statsMap[user].Files++
			statsMap[user].Tokens += change.TokensAfter

			usd, cny, _ := costMgr.CalculateCost(change.TokensAfter, model)
			statsMap[user].USD += usd
			statsMap[user].CNY += cny
		}
	}

	// Convert map to slice
	result := make([]UserStats, 0, len(statsMap))
	for _, stats := range statsMap {
		result = append(result, *stats)
	}

	// Sort by tokens (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Tokens > result[j].Tokens
	})

	return result, nil
}

// calculateBranchStats computes token statistics per branch
func calculateBranchStats(repo *git.Repository, root string, tracked []TrackedFile, branches []plumbing.Reference, model string) ([]BranchStats, error) {
	costMgr := NewCostManager()

	// Map: branch -> BranchStats
	statsMap := make(map[string]*BranchStats)

	for _, branchRef := range branches {
		branchName := branchRef.Name().Short()
		if branchName == "HEAD" {
			continue
		}

		if statsMap[branchName] == nil {
			statsMap[branchName] = &BranchStats{Branch: branchName}
		}

		// Get the commit at branch tip
		hash := branchRef.Hash()

		// Get all commits on this branch
		commitIter, err := repo.Log(&git.LogOptions{
			From:  hash,
			Order: git.LogOrderCommitterTime,
		})
		if err != nil {
			continue
		}

		processedCommits := make(map[plumbing.Hash]bool)

		for {
			c, err := commitIter.Next()
			if err != nil {
				break
			}

			// Avoid processing same commit multiple times
			if processedCommits[c.Hash] {
				continue
			}
			processedCommits[c.Hash] = true

			statsMap[branchName].Commits++

			// Get file changes in this commit
			changes, err := getPromptChangesInCommit(repo, root, tracked, c)
			if err != nil {
				continue
			}

			for _, change := range changes {
				// Only count if this is the first branch to touch this file
				if !isFileInOtherBranch(repo, root, tracked, branchRef, c, change.Path) {
					statsMap[branchName].Files++
					statsMap[branchName].Tokens += change.TokensAfter

					usd, cny, _ := costMgr.CalculateCost(change.TokensAfter, model)
					statsMap[branchName].USD += usd
					statsMap[branchName].CNY += cny
				}
			}
		}

		commitIter.Close()
	}

	// Convert map to slice
	result := make([]BranchStats, 0, len(statsMap))
	for _, stats := range statsMap {
		if stats.Tokens > 0 {
			result = append(result, *stats)
		}
	}

	// Sort by tokens (descending)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Tokens > result[j].Tokens
	})

	return result, nil
}

// isFileInOtherBranch checks if a file exists in other branches at the same commit
func isFileInOtherBranch(repo *git.Repository, root string, tracked []TrackedFile, currentBranch plumbing.Reference, commit *object.Commit, filePath string) bool {
	// For simplicity, always return false
	// A proper implementation would check if the file was introduced in another branch first
	// This is a complex operation that would require traversing the commit DAG
	return false
}

// getLocalBranches returns all local branches
func getLocalBranches(repo *git.Repository) ([]plumbing.Reference, error) {
	branchIter, err := repo.Branches()
	if err != nil {
		return nil, err
	}

	var branches []plumbing.Reference
	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		if !ref.Name().IsRemote() {
			branches = append(branches, *ref)
		}
		return nil
	})

	return branches, err
}

// showUserStats displays user statistics
func showUserStats(stats []UserStats, model string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Cost breakdown by user (model: %s)\n\n", model)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "User", "Commits", "Files", "Tokens", "USD", "CNY")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "----", "-------", "-----", "------", "---", "---")

	totalCommits := 0
	totalFiles := 0
	totalTokens := 0
	totalUSD := 0.0
	totalCNY := 0.0

	for _, s := range stats {
		totalCommits += s.Commits
		totalFiles += s.Files
		totalTokens += s.Tokens
		totalUSD += s.USD
		totalCNY += s.CNY

		shortUser := s.User
		if len(shortUser) > 12 {
			shortUser = shortUser[:9] + ".."
		}

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\t%s\n",
			shortUser,
			s.Commits,
			s.Files,
			s.Tokens,
			FormatCurrency(s.USD, "USD"),
			FormatCurrency(s.CNY, "CNY"),
		)
	}

	w.Flush()

	if len(stats) > 0 {
		fmt.Printf("\n%s\t%d\t%d\t%d\t%s\t%s\n",
			"TOTAL",
			totalCommits,
			totalFiles,
			totalTokens,
			FormatCurrency(totalUSD, "USD"),
			FormatCurrency(totalCNY, "CNY"),
		)
	}

	return nil
}

// showBranchStats displays branch statistics
func showBranchStats(stats []BranchStats, model string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Cost breakdown by branch (model: %s)\n\n", model)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "Branch", "Commits", "Files", "Tokens", "USD", "CNY")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", "------", "-------", "-----", "------", "---", "---")

	totalCommits := 0
	totalFiles := 0
	totalTokens := 0
	totalUSD := 0.0
	totalCNY := 0.0

	for _, s := range stats {
		totalCommits += s.Commits
		totalFiles += s.Files
		totalTokens += s.Tokens
		totalUSD += s.USD
		totalCNY += s.CNY

		shortBranch := s.Branch
		if len(shortBranch) > 15 {
			shortBranch = shortBranch[:12] + ".."
		}

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\t%s\n",
			shortBranch,
			s.Commits,
			s.Files,
			s.Tokens,
			FormatCurrency(s.USD, "USD"),
			FormatCurrency(s.CNY, "CNY"),
		)
	}

	w.Flush()

	if len(stats) > 0 {
		fmt.Printf("\n%s\t%d\t%d\t%d\t%s\t%s\n",
			"TOTAL",
			totalCommits,
			totalFiles,
			totalTokens,
			FormatCurrency(totalUSD, "USD"),
			FormatCurrency(totalCNY, "CNY"),
		)
	}

	return nil
}
