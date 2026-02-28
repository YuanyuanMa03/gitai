package gitai

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mayuanyuan/gitai/pkg/token"
	"github.com/urfave/cli/v2"
)

// TrackCmd tracks a .prompt file
func TrackCmd(c *cli.Context) error {
	if c.Args().Len() == 0 {
		return fmt.Errorf("usage: gitai track <file.prompt>")
	}

	filePath := c.Args().Get(0)

	// Get git root
	repo, root := MustGetGitRepo()
	_ = repo

	// Get relative path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if file is within git repo
	if !strings.HasPrefix(absPath, root) {
		return fmt.Errorf("file is outside git repository")
	}

	relPath, err := GetRelativePath(absPath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Check if it's a .prompt file
	if !strings.HasSuffix(filePath, ".prompt") {
		return fmt.Errorf("file must have .prompt extension")
	}

	// Create storage
	storage := NewStorage(root)

	// Check if .gitai/ is initialized
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Read file content
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// Calculate hash
	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	// Parse the prompt file using Day2 parser
	parsed := token.Parse(contentStr)

	// Get file info
	fileInfo, _ := os.Stat(absPath)

	// Build description from task (first 100 chars)
	description := parsed.Task
	if len(description) > 100 {
		description = description[:100] + "..."
	}

	// Calculate token breakdown
	taskTokens := 0
	constraintsTokens := 0
	examplesTokens := 0
	inputTokens := 0
	outputTokens := 0
	contextTokens := 0
	otherTokens := 0

	for _, section := range parsed.Sections {
		switch section.Type {
		case token.SectionTask:
			taskTokens += section.Tokens
		case token.SectionConstraints:
			constraintsTokens += section.Tokens
		case token.SectionExamples:
			examplesTokens += section.Tokens
		case token.SectionInput:
			inputTokens += section.Tokens
		case token.SectionOutput:
			outputTokens += section.Tokens
		case token.SectionContext:
			contextTokens += section.Tokens
		case token.SectionOther:
			otherTokens += section.Tokens
		}
	}

	// Use parsed role as fallback
	role := parsed.Role
	if role == "" {
		role = "default"
	}

	// Create tracked file entry
	trackedFile := TrackedFile{
		Path:              relPath,
		AddedAt:           time.Now(),
		LastModified:      fileInfo.ModTime(),
		CurrentHash:       hashStr,
		Role:              role,
		Description:       description,
		Tokens:            parsed.TotalTokens,
		TokenCostUSD:      0,
		Task:              parsed.Task,
		Constraints:       parsed.Constraints,
		Examples:          parsed.Examples,
		Input:             parsed.Input,
		Output:            parsed.Output,
		Context:           parsed.Context,
		TaskTokens:        taskTokens,
		ConstraintsTokens: constraintsTokens,
		ExamplesTokens:    examplesTokens,
		InputTokens:       inputTokens,
		OutputTokens:      outputTokens,
		ContextTokens:     contextTokens,
		OtherTokens:       otherTokens,
	}

	// Add to storage
	if err := storage.AddTrackedFile(trackedFile); err != nil {
		return fmt.Errorf("failed to track file: %w", err)
	}

	// Print success with detailed breakdown
	fmt.Printf("✓ Tracked: %s\n", relPath)
	fmt.Printf("  Role: %s\n", role)
	fmt.Printf("  Total Tokens: %d\n", parsed.TotalTokens)
	fmt.Printf("  Hash: %s\n", hashStr[:16])

	// Print section breakdown if we have sections
	if len(parsed.Sections) > 0 {
		fmt.Printf("  Sections:\n")
		for _, section := range parsed.Sections {
			if section.Tokens > 0 {
				fmt.Printf("    %s: %d tokens\n", section.Type, section.Tokens)
			}
		}
	}

	return nil
}

// parsePromptContent extracts role and description from .prompt file
// DEPRECATED: Use token.Parse instead (kept for compatibility)
func parsePromptContent(content string) (role, description string) {
	lines := strings.Split(content, "\n")
	role = "default"
	description = ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# role:") || strings.HasPrefix(line, "# Role:") {
			role = strings.TrimSpace(strings.TrimPrefix(line, "# role:"))
			role = strings.TrimSpace(strings.TrimPrefix(role, "# Role:"))
		}
		if strings.HasPrefix(line, "# description:") || strings.HasPrefix(line, "# Description:") {
			description = strings.TrimSpace(strings.TrimPrefix(line, "# description:"))
			description = strings.TrimSpace(strings.TrimPrefix(description, "# Description:"))
		}
	}

	if role == "" {
		role = "default"
	}

	return role, description
}

// EstimateTokens estimates token count
// DEPRECATED: Use token.Estimate instead (kept for compatibility)
func EstimateTokens(text string) int {
	return token.Estimate(text)
}
