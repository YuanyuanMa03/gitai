package gitai

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/mayuanyuan/gitai/pkg/token"
	"github.com/urfave/cli/v2"
)

// DiffCmd shows differences between two versions of a prompt file
func DiffCmd(c *cli.Context) error {
	args := c.Args()

	// Parse arguments
	// Usage: gitai diff <file.prompt> [commit] [commit]
	// - gitai diff file        : compare working dir to HEAD
	// - gitai diff file A      : compare working dir to commit A
	// - gitai diff file A B    : compare commit A to commit B

	filePath := ""
	commitA := ""      // empty = working directory
	commitB := "HEAD"  // default compare target

	switch args.Len() {
	case 0:
		return fmt.Errorf("usage: gitai diff <file.prompt> [commit] [commit]")
	case 1:
		filePath = args.Get(0)
		// commitA = "" (working dir), commitB = "HEAD"
	case 2:
		filePath = args.Get(0)
		commitA = ""           // working dir
		commitB = args.Get(1)  // compare to this commit
	case 3:
		filePath = args.Get(0)
		commitA = args.Get(1)
		commitB = args.Get(2)
	default:
		return fmt.Errorf("too many arguments")
	}

	// Validate file extension
	if !strings.HasSuffix(filePath, ".prompt") {
		return fmt.Errorf("file must have .prompt extension")
	}

	// Get git root
	repo, _ := MustGetGitRepo()

	// Get relative path
	absPath, err := getAbsPath(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	relPath, err := GetRelativePath(absPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Get file content for comparison
	var contentA, contentB *string
	var displayA, displayB string
	var workingDir bool

	if commitA == "" {
		// Compare working directory to commit
		workingDir = true
		displayA = "working"

		// Read from working directory
		data, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		content := string(data)
		contentA = &content

		// Get content from commit B
		hashB, err := resolveCommitHash(repo, commitB)
		if err != nil {
			return fmt.Errorf("failed to resolve commit %s: %w", commitB, err)
		}
		displayB = hashB.String()[:12]

		contentB, err = getFileContentAtCommit(repo, relPath, hashB)
		if err != nil && err != ErrFileNotFound {
			return fmt.Errorf("failed to get file at %s: %w", commitB, err)
		}
	} else {
		// Compare two commits
		workingDir = false

		hashA, err := resolveCommitHash(repo, commitA)
		if err != nil {
			return fmt.Errorf("failed to resolve commit %s: %w", commitA, err)
		}
		displayA = hashA.String()[:12]

		hashB, err := resolveCommitHash(repo, commitB)
		if err != nil {
			return fmt.Errorf("failed to resolve commit %s: %w", commitB, err)
		}
		displayB = hashB.String()[:12]

		contentA, err = getFileContentAtCommit(repo, relPath, hashA)
		if err != nil && err != ErrFileNotFound {
			return fmt.Errorf("failed to get file at %s: %w", commitA, err)
		}

		contentB, err = getFileContentAtCommit(repo, relPath, hashB)
		if err != nil && err != ErrFileNotFound {
			return fmt.Errorf("failed to get file at %s: %w", commitB, err)
		}
	}

	// Handle file not found cases
	if contentA == nil && contentB == nil {
		return fmt.Errorf("file not found in either location")
	}

	// Parse and compare
	var parsedA, parsedB *token.ParsedPrompt
	if contentA != nil {
		parsedA = token.Parse(*contentA)
	}
	if contentB != nil {
		parsedB = token.Parse(*contentB)
	}

	// Display diff
	return displayDiff(filePath, displayA, displayB, workingDir, parsedA, parsedB)
}

// displayDiff shows the diff between two parsed prompts
func displayDiff(filePath, displayA, displayB string, workingDir bool, parsedA, parsedB *token.ParsedPrompt) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Display shows: displayA → displayB (user's perspective)
	fmt.Printf("Comparing: %s\n", filePath)
	if workingDir {
		fmt.Printf("  (working directory) → %s\n\n", displayB)
	} else {
		fmt.Printf("  %s → %s\n\n", displayA, displayB)
	}

	// For the diff display:
	// - commitA content is "before" (removed, -)
	// - commitB content is "after" (added, +)

	// Show role change
	roleA := ""
	roleB := ""
	if parsedA != nil {
		roleA = parsedA.Role
	}
	if parsedB != nil {
		roleB = parsedB.Role
	}

	if roleA != roleB {
		fmt.Fprintf(w, "Role:\n")
		fmt.Fprintf(w, "  - %s\n", roleA)  // From commitA (before)
		fmt.Fprintf(w, "  + %s\n\n", roleB) // From commitB (after)
	} else if roleA != "" {
		fmt.Fprintf(w, "Role: %s\n\n", roleA)
	}

	// Show token change by section
	// Before = commitA, After = commitB
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Section", "Before", "After", "Delta")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "-------", "------", "-----", "-----")

	totalBefore := 0
	totalAfter := 0

	sections := []struct {
		name        string
		tokenType   token.SectionType
		tokensA     int
		tokensB     int
		hasContentA bool
		hasContentB bool
		contentA    string
		contentB    string
	}{
		{"Task", token.SectionTask, 0, 0, false, false, "", ""},
		{"Constraints", token.SectionConstraints, 0, 0, false, false, "", ""},
		{"Examples", token.SectionExamples, 0, 0, false, false, "", ""},
		{"Input", token.SectionInput, 0, 0, false, false, "", ""},
		{"Output", token.SectionOutput, 0, 0, false, false, "", ""},
		{"Context", token.SectionContext, 0, 0, false, false, "", ""},
		{"Other", token.SectionOther, 0, 0, false, false, "", ""},
	}

	// Fill in data from parsed prompts
	if parsedA != nil {
		for _, s := range parsedA.Sections {
			for i := range sections {
				if sections[i].tokenType == s.Type {
					sections[i].tokensA = s.Tokens
					sections[i].hasContentA = true
					sections[i].contentA = s.Content
					break
				}
			}
		}
		totalBefore = parsedA.TotalTokens
	}

	if parsedB != nil {
		for _, s := range parsedB.Sections {
			for i := range sections {
				if sections[i].tokenType == s.Type {
					sections[i].tokensB = s.Tokens
					sections[i].hasContentB = true
					sections[i].contentB = s.Content
					break
				}
			}
		}
		totalAfter = parsedB.TotalTokens
	}

	// Print section table: Before=commitA, After=commitB
	for _, s := range sections {
		if s.hasContentA || s.hasContentB {
			delta := s.tokensB - s.tokensA  // After - Before
			deltaStr := fmt.Sprintf("%+d", delta)
			fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", s.name, s.tokensA, s.tokensB, deltaStr)
		}
	}

	// Total: Before=commitA, After=commitB
	totalDelta := totalAfter - totalBefore
	totalDeltaStr := fmt.Sprintf("%+d", totalDelta)
	fmt.Fprintf(w, "%s\t%d\t%d\t%s\n\n", "TOTAL", totalBefore, totalAfter, totalDeltaStr)

	w.Flush()

	// Show content preview for changed sections
	fmt.Println("Content Changes:")
	fmt.Println("---------------")

	for _, s := range sections {
		if s.hasContentA || s.hasContentB {
			if s.contentA != s.contentB {
				fmt.Printf("\n## %s\n", s.name)

				if s.contentA != "" && s.contentB != "" {
					// Both exist, show diff
					// before=contentA (commitA), after=contentB (commitB)
					linesA := strings.Split(s.contentA, "\n")
					linesB := strings.Split(s.contentB, "\n")
					showContentDiff(linesA, linesB)
				} else if s.contentA != "" {
					// Only in before (commitA)
					fmt.Printf("  - Removed (%d tokens)\n", s.tokensA)
					for _, line := range truncateLines(s.contentA, 3) {
						fmt.Printf("    - %s\n", line)
					}
				} else if s.contentB != "" {
					// Only in after (commitB)
					fmt.Printf("  + Added (%d tokens)\n", s.tokensB)
					for _, line := range truncateLines(s.contentB, 3) {
						fmt.Printf("    + %s\n", line)
					}
				}
			}
		}
	}

	return nil
}

// showContentDiff shows line-by-line content diff
func showContentDiff(before, after []string) {
	maxLines := len(before)
	if len(after) > maxLines {
		maxLines = len(after)
	}

	for i := 0; i < maxLines && i < 6; i++ {
		lineBefore := ""
		lineAfter := ""

		if i < len(before) {
			lineBefore = before[i]
		}
		if i < len(after) {
			lineAfter = after[i]
		}

		if lineBefore == lineAfter {
			if lineBefore != "" {
				fmt.Printf("    %s\n", lineBefore)
			}
		} else {
			if lineBefore != "" {
				fmt.Printf("  - %s\n", lineBefore)
			}
			if lineAfter != "" {
				fmt.Printf("  + %s\n", lineAfter)
			}
		}
	}

	if len(after) > 6 {
		fmt.Printf("  + ... (%d more lines)\n", len(after)-6)
	}
	if len(before) > 6 {
		fmt.Printf("  - ... (%d more lines)\n", len(before)-6)
	}
}

// truncateLines returns first n lines with ... if truncated
func truncateLines(content string, max int) []string {
	lines := strings.Split(content, "\n")
	if len(lines) > max {
		lines = append(lines[:max], "...")
	}
	return lines
}

// LogCmd shows commit history with token deltas
func LogCmd(c *cli.Context) error {
	// Get git root
	repo, root := MustGetGitRepo()

	// Get storage
	storage := NewStorage(root)
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

	// Parse limit
	limit := 10
	if c.IsSet("limit") {
		limit = c.Int("limit")
	}

	// Get commit log
	commits, err := getCommitHistory(repo, limit)
	if err != nil {
		return fmt.Errorf("failed to get commit history: %w", err)
	}

	if len(commits) == 0 {
		fmt.Println("No commits found.")
		return nil
	}

	// Show log with token info
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "Commit", "Author", "Files", "Tokens", "Message")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "-------", "------", "-----", "------", "-------")

	for _, commit := range commits {
		// Get file changes in this commit
		fileChanges, _ := getPromptChangesInCommit(repo, root, tracked, commit)

		totalTokens := 0
		for _, fc := range fileChanges {
			totalTokens += fc.TokensAfter
		}

		author := commit.Author.Name
		if len(author) > 12 {
			author = author[:12] + ".."
		}

		message := commit.Message
		lines := strings.Split(message, "\n")
		if len(lines) > 0 {
			message = lines[0]
		}
		if len(message) > 40 {
			message = message[:40] + ".."
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\n",
			commit.Hash.String()[:8],
			author,
			len(fileChanges),
			totalTokens,
			message,
		)
	}

	w.Flush()

	return nil
}

// FileChange represents a prompt file change in a commit
type FileChange struct {
	Path        string
	TokensBefore int
	TokensAfter  int
	Delta       int
}

// resolveCommitHash resolves a commit reference to a hash
func resolveCommitHash(repo *git.Repository, ref string) (plumbing.Hash, error) {
	if ref == "" {
		ref = "HEAD"
	}

	// Try as hash first
	hash := plumbing.NewHash(ref)
	if !hash.IsZero() {
		return hash, nil
	}

	// Try as reference
	h, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return plumbing.ZeroHash, err
	}

	return *h, nil
}

// getFileContentAtCommit gets file content at a specific commit
func getFileContentAtCommit(repo *git.Repository, path string, hash plumbing.Hash) (*string, error) {
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	file, err := tree.File(path)
	if err != nil {
		return nil, ErrFileNotFound
	}

	contents, err := file.Contents()
	if err != nil {
		return nil, err
	}

	return &contents, nil
}

// getCommitHistory gets recent commits
func getCommitHistory(repo *git.Repository, limit int) ([]*object.Commit, error) {
	ref, err := repo.Head()
	if err != nil {
		return nil, err
	}

	commitIter, err := repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}
	defer commitIter.Close()

	commits := make([]*object.Commit, 0, limit)
	i := 0
	for {
		commit, err := commitIter.Next()
		if err != nil {
			break
		}

		commits = append(commits, commit)
		i++

		if i >= limit {
			break
		}
	}

	return commits, nil
}

// getPromptChangesInCommit gets prompt file changes in a commit
func getPromptChangesInCommit(repo *git.Repository, root string, tracked []TrackedFile, commit *object.Commit) ([]FileChange, error) {
	changes := make([]FileChange, 0)

	// Get parent commit
	parents := commit.Parents()
	var parent *object.Commit

	if p, err := parents.Next(); err == nil {
		parent = p
	}

	// Check each tracked file
	for _, file := range tracked {
		var contentBefore, contentAfter *string

		// Get content from parent (before)
		if parent != nil {
			contentBefore, _ = getFileContentAtCommit(repo, file.Path, parent.Hash)
		}

		// Get content from current commit (after)
		contentAfter, _ = getFileContentAtCommit(repo, file.Path, commit.Hash)

		// Skip if file doesn't exist in either
		if contentBefore == nil && contentAfter == nil {
			continue
		}

		tokensBefore := 0
		tokensAfter := 0

		if contentBefore != nil {
			tokensBefore = token.Estimate(*contentBefore)
		}
		if contentAfter != nil {
			tokensAfter = token.Estimate(*contentAfter)
		}

		// Only add if there's actual content
		if tokensBefore > 0 || tokensAfter > 0 {
			changes = append(changes, FileChange{
				Path:         file.Path,
				TokensBefore: tokensBefore,
				TokensAfter:  tokensAfter,
				Delta:        tokensAfter - tokensBefore,
			})
		}
	}

	return changes, nil
}

// ErrFileNotFound is returned when file is not found in commit
var ErrFileNotFound = fmt.Errorf("file not found")
