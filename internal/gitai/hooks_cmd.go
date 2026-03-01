package gitai

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

// HooksInstallCmd installs GitAI hooks
func HooksInstallCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Install hooks
	if err := InstallHooks(root); err != nil {
		return fmt.Errorf("failed to install hooks: %w", err)
	}

	fmt.Println("✓ GitAI hooks installed:")
	fmt.Println("  • pre-commit - Quality check on prompt files")
	fmt.Println("  • post-commit - Record intent after commit")
	fmt.Println("\nHooks installed to .git/hooks/")
	fmt.Println("Uninstall with: gitai hooks uninstall")

	return nil
}

// HooksUninstallCmd removes GitAI hooks
func HooksUninstallCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Uninstall hooks
	if err := UninstallHooks(root); err != nil {
		return fmt.Errorf("failed to uninstall hooks: %w", err)
	}

	fmt.Println("✓ GitAI hooks uninstalled")
	fmt.Println("\nExisting hooks were backed up with .backup extension")

	return nil
}

// PreCommitCheckCmd runs pre-commit quality checks
func PreCommitCheckCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		// Not initialized, skip checks
		return nil
	}

	// Run quality checks
	if err := RunPreCommitCheck(storage); err != nil {
		fmt.Fprintf(os.Stderr, "GitAI pre-commit check failed:\n%v\n", err)
		os.Exit(1)
	}

	return nil
}

// PostCommitRecordCmd records intent after commit
func PostCommitRecordCmd(c *cli.Context) error {
	// Get git root
	repo, root := MustGetGitRepo()
	_ = repo

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return nil
	}

	// Get latest commit hash
	commitHash, err := getLatestCommitHash()
	if err != nil {
		return nil // Silently fail if we can't get commit
	}

	// Check if intent already exists for this commit
	if _, err := storage.GetIntentByCommit(commitHash); err == nil {
		// Intent already recorded, skip
		return nil
	}

	// Get commit message
	commitMsg, err := getCommitMessage(commitHash)
	if err != nil {
		return nil
	}

	// Detect used prompts
	usedPrompts, _ := detectUsedPrompts(storage, root)

	// Create intent
	intent := &Intent{
		ID:           GenerateIntentID(),
		CommitHash:   commitHash,
		PromptHashes: usedPrompts,
		Message:      commitMsg,
		Category:     DetectIntentCategory(commitMsg),
		Timestamp:    getCurrentTimestamp(),
	}

	// Save intent silently
	_ = storage.SaveIntent(intent)

	return nil
}

// getCommitMessage gets the message for a commit
func getCommitMessage(commitHash string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%s", commitHash)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getCurrentTimestamp returns current time
func getCurrentTimestamp() time.Time {
	return time.Now()
}
