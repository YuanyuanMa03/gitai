package gitai

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

// InitCmd initializes .gitai/ in current git repository
func InitCmd(c *cli.Context) error {
	// Check if we're in a git repo
	repoInfo, err := GetGitRepoInfo()
	if err != nil {
		return fmt.Errorf("failed to get git info: %w", err)
	}

	if !repoInfo.IsRepo {
		return fmt.Errorf("not in a git repository")
	}

	// Create storage
	storage := NewStorage(repoInfo.Path)

	// Check if already initialized
	if storage.IsInitialized() {
		return fmt.Errorf(".gitai/ already exists in %s", repoInfo.Path)
	}

	// Initialize
	if err := storage.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize .gitai/: %w", err)
	}

	// Print success message
	fmt.Printf("✓ Initialized .gitai/ in %s\n", repoInfo.Path)
	fmt.Printf("  Repository: %s\n", repoInfo.Path)
	fmt.Printf("  Branch: %s\n", repoInfo.Branch)
	fmt.Printf("  Head: %s\n", repoInfo.Commit[:12])

	return nil
}
