package gitai

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/urfave/cli/v2"
)

// CommitCmd creates a commit with intent tracking
func CommitCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Parse options
	message := c.String("message")
	intentOnly := c.Bool("intent-only")
	edit := c.Bool("edit")
	dryRun := c.Bool("dry-run")

	// Get commit message
	var commitMessage string
	if edit {
		// Open editor for commit message
		commitMessage = getCommitMessageFromEditor()
		if commitMessage == "" {
			return fmt.Errorf("empty commit message")
		}
	} else if message != "" {
		commitMessage = message
	} else {
		// Read staged files for context
		stagedMsg := getStagedFilesSummary()
		if stagedMsg == "" {
			return fmt.Errorf("no staged changes. Use 'git add' to stage files first")
		}
		commitMessage = stagedMsg
	}

	// Detect used prompts from staged files
	usedPrompts, err := detectUsedPrompts(storage, root)
	if err != nil {
		return fmt.Errorf("failed to detect prompts: %w", err)
	}

	// Calculate estimated cost
	estimatedCost := 0.0
	config, _ := LoadConfig(root)
	if config != nil {
		for _, prompt := range usedPrompts {
			tf, err := storage.GetTrackedFileByPath(prompt)
			if err == nil {
				pricing, _ := config.GetModelPricing(config.Model)
				if pricing != nil {
					usd, _ := calculateCostWithPricing(tf.Tokens, pricing)
					estimatedCost += usd
				}
			}
		}
	}

	// Create intent
	intent := &Intent{
		ID:            GenerateIntentID(),
		CommitHash:    "", // Will be set after commit
		PromptHashes:  usedPrompts,
		Message:       commitMessage,
		Category:      DetectIntentCategory(commitMessage),
		EstimatedCost: estimatedCost,
		Timestamp:     time.Now(),
	}

	// Display intent summary
	if dryRun || !intentOnly {
		displayIntentSummary(intent, usedPrompts)
	}

	if dryRun {
		fmt.Println("\n[Dry run - no commit created]")
		return nil
	}

	// Note: post-commit hook will reference the intent

	if intentOnly {
		// Only save intent, don't create commit
		if err := storage.SaveIntent(intent); err != nil {
			return fmt.Errorf("failed to save intent: %w", err)
		}
		fmt.Printf("\nIntent saved: %s\n", intent.ID)
		fmt.Println("Link it to a commit with: gitai link-intent <commit>")
		return nil
	}

	// Create the actual commit
	if err := createGitCommit(commitMessage); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Get the commit hash
	commitHash, err := getLatestCommitHash()
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}
	intent.CommitHash = commitHash

	// Save intent
	if err := storage.SaveIntent(intent); err != nil {
		return fmt.Errorf("failed to save intent: %w", err)
	}

	fmt.Printf("\n✓ Commit created: %s\n", commitHash)
	fmt.Printf("✓ Intent saved: %s\n", intent.ID)

	return nil
}

// detectUsedPrompts detects which prompts were likely used from staged files
func detectUsedPrompts(storage *Storage, root string) ([]string, error) {
	// Get staged files
	staged, err := getStagedFiles()
	if err != nil {
		return nil, err
	}

	// Get all tracked prompts
	tracked, _ := storage.ListAllTracked()

	// Simple heuristic: check for .prompt files in staged files
	var used []string
	for _, file := range staged {
		if strings.HasSuffix(file, ".prompt") {
			// Check if tracked
			for _, tf := range tracked {
				if tf.Path == file || strings.HasSuffix(tf.Path, file) {
					used = append(used, tf.Path)
				}
			}
		}
	}

	return used, nil
}

// getStagedFiles gets list of staged files
func getStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	result := []string{}
	for _, f := range files {
		if f != "" {
			result = append(result, f)
		}
	}

	return result, nil
}

// getStagedFilesSummary generates a summary of staged changes
func getStagedFilesSummary() string {
	files, err := getStagedFiles()
	if err != nil || len(files) == 0 {
		return ""
	}

	// Get diff stats for context
	cmd := exec.Command("git", "diff", "--cached", "--stat")
	output, _ := cmd.Output()

	var summary strings.Builder
	summary.WriteString("Changes to be committed:\n")
	summary.WriteString(string(output))

	return summary.String()
}

// getCommitMessageFromEditor opens editor for commit message
func getCommitMessageFromEditor() string {
	editor := os.Getenv("GIT_EDITOR")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vim"
	}

	// Create temp file
	tmpFile := "/tmp/gitai-commit-msg.txt"
	template := fmt.Sprintf("# Enter commit message. Lines starting with # will be ignored.\n#\n# Staged files:\n")
	staged, _ := getStagedFiles()
	for _, f := range staged {
		template += fmt.Sprintf("#   %s\n", f)
	}
	os.WriteFile(tmpFile, []byte(template), 0644)

	// Open editor
	cmd := exec.Command(editor, tmpFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	// Read result
	content, _ := os.ReadFile(tmpFile)
	var lines []string
	for _, line := range strings.Split(string(content), "\n") {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}

	os.Remove(tmpFile)
	return strings.Join(lines, "\n")
}

// createGitCommit creates a git commit
func createGitCommit(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getLatestCommitHash returns the hash of the most recent commit
func getLatestCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// displayIntentSummary displays the intent summary
func displayIntentSummary(intent *Intent, usedPrompts []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Println("Intent Summary:")
	fmt.Fprintf(w, "%s\t%s\n", "ID:", intent.ID)
	fmt.Fprintf(w, "%s\t%s\n", "Category:", intent.Category)
	fmt.Fprintf(w, "%s\t%s\n", "Message:", truncateString(intent.Message, 60))

	if len(usedPrompts) > 0 {
		fmt.Fprintf(w, "%s\t%d\n", "Prompts Used:", len(usedPrompts))
		for _, p := range usedPrompts {
			fmt.Fprintf(w, "  →\t%s\n", p)
		}
	}

	if intent.EstimatedCost > 0 {
		fmt.Fprintf(w, "%s\t%s\n", "Est. Cost:", FormatCurrency(intent.EstimatedCost, "USD"))
	}

	w.Flush()
}

// LogIntentsCmd displays intent history
func LogIntentsCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Parse options
	limit := c.Int("limit")
	category := c.String("category")

	// Get all intents
	intents, err := storage.ListIntents()
	if err != nil {
		return fmt.Errorf("failed to load intents: %w", err)
	}

	if len(intents) == 0 {
		fmt.Println("No intents found. Use 'gitai commit' to track intents.")
		return nil
	}

	// Filter by category
	if category != "" {
		filtered := []*Intent{}
		for _, intent := range intents {
			if intent.Category == category {
				filtered = append(filtered, intent)
			}
		}
		intents = filtered
	}

	// Limit results
	if limit > 0 && len(intents) > limit {
		intents = intents[:limit]
	}

	// Display
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Intent History (%d intents)\n\n", len(intents))

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "ID", "Category", "Commit", "Message", "Time")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "--", "--------", "------", "-------", "----")

	for _, intent := range intents {
		shortID := intent.ID
		if len(shortID) > 12 {
			shortID = shortID[:9] + ".."
		}

		shortMsg := truncateString(intent.Message, 40)
		shortCommit := intent.CommitHash
		if len(shortCommit) > 8 {
			shortCommit = shortCommit[:7]
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			shortID,
			intent.Category,
			shortCommit,
			shortMsg,
			intent.Timestamp.Format("2006-01-02 15:04"),
		)
	}

	w.Flush()

	return nil
}
