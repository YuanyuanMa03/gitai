package gitai

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

// StatusCmd shows tracked files status
func StatusCmd() error {
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
	files, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No tracked files.")
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "File", "Role", "Tokens", "Hash")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "----", "----", "------", "----")

	totalTokens := 0
	totalTaskTokens := 0
	totalConstraintsTokens := 0
	totalExamplesTokens := 0
	for _, f := range files {
		shortHash := f.CurrentHash
		if len(shortHash) > 12 {
			shortHash = shortHash[:12]
		}
		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", f.Path, f.Role, f.Tokens, shortHash)
		totalTokens += f.Tokens
		totalTaskTokens += f.TaskTokens
		totalConstraintsTokens += f.ConstraintsTokens
		totalExamplesTokens += f.ExamplesTokens
	}

	w.Flush()

	fmt.Printf("\nTotal: %d files, %d tokens\n", len(files), totalTokens)

	// Show section breakdown if available
	if totalTaskTokens > 0 || totalConstraintsTokens > 0 || totalExamplesTokens > 0 {
		fmt.Printf("\nToken breakdown:\n")
		if totalTaskTokens > 0 {
			fmt.Printf("  Task: %d tokens\n", totalTaskTokens)
		}
		if totalConstraintsTokens > 0 {
			fmt.Printf("  Constraints: %d tokens\n", totalConstraintsTokens)
		}
		if totalExamplesTokens > 0 {
			fmt.Printf("  Examples: %d tokens\n", totalExamplesTokens)
		}
	}

	return nil
}

// ShowCmd shows detailed information about a tracked file
func ShowCmd(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("usage: gitai show <file.prompt>")
	}

	filePath := c.Args().Get(0)

	// Get git root
	repo, root := MustGetGitRepo()
	_ = repo

	// Get relative path
	absPath, err := getAbsPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	relPath, err := GetRelativePath(absPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Create storage
	storage := NewStorage(root)

	// Check if .gitai/ is initialized
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Get tracked file
	tracked, err := storage.GetTrackedFileByPath(relPath)
	if err != nil {
		return fmt.Errorf("file not tracked: %s", filePath)
	}

	// Print detailed info
	PrintTrackedFileDetailed(*tracked)

	return nil
}

// PrintTrackedFile prints details of a single tracked file
func PrintTrackedFile(f TrackedFile) {
	fmt.Printf("File: %s\n", f.Path)
	fmt.Printf("  Role: %s\n", f.Role)
	fmt.Printf("  Tokens: %d\n", f.Tokens)
	fmt.Printf("  Added: %s\n", f.AddedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Modified: %s\n", f.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Hash: %s\n", f.CurrentHash)
	if f.Description != "" {
		fmt.Printf("  Description: %s\n", strings.TrimSpace(f.Description))
	}
}

// PrintTrackedFileDetailed prints detailed breakdown of a tracked file
func PrintTrackedFileDetailed(f TrackedFile) {
	fmt.Printf("File: %s\n", f.Path)
	fmt.Printf("  Role: %s\n", f.Role)
	fmt.Printf("  Total Tokens: %d\n", f.Tokens)
	fmt.Printf("  Added: %s\n", f.AddedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Modified: %s\n", f.LastModified.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Hash: %s\n", f.CurrentHash)

	// Print section breakdown
	fmt.Printf("\n  Token Breakdown:\n")
	if f.TaskTokens > 0 {
		fmt.Printf("    Task: %d tokens\n", f.TaskTokens)
	}
	if f.ConstraintsTokens > 0 {
		fmt.Printf("    Constraints: %d tokens\n", f.ConstraintsTokens)
	}
	if f.ExamplesTokens > 0 {
		fmt.Printf("    Examples: %d tokens\n", f.ExamplesTokens)
	}
	if f.InputTokens > 0 {
		fmt.Printf("    Input: %d tokens\n", f.InputTokens)
	}
	if f.OutputTokens > 0 {
		fmt.Printf("    Output: %d tokens\n", f.OutputTokens)
	}
	if f.ContextTokens > 0 {
		fmt.Printf("    Context: %d tokens\n", f.ContextTokens)
	}
	if f.OtherTokens > 0 {
		fmt.Printf("    Other: %d tokens\n", f.OtherTokens)
	}

	// Print content previews
	if f.Task != "" {
		fmt.Printf("\n  Task:\n")
		for _, line := range splitLines(f.Task, 3) {
			fmt.Printf("    %s\n", line)
		}
	}
	if f.Constraints != "" {
		fmt.Printf("\n  Constraints:\n")
		for _, line := range splitLines(f.Constraints, 3) {
			fmt.Printf("    %s\n", line)
		}
	}
}

func getAbsPath(path string) (string, error) {
	if strings.HasPrefix(path, "/") {
		return path, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd + "/" + path, nil
}

func splitLines(s string, maxLines int) []string {
	lines := strings.Split(s, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "...")
	}
	return lines
}
