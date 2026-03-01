package gitai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mayuanyuan/gitai/pkg/token"
)

// PromptQualityResult represents the result of prompt quality check
type PromptQualityResult struct {
	HasIssues bool     `json:"has_issues"`
	MinScore  int      `json:"min_score"`
	Warnings  []string `json:"warnings"`
	Errors    []string `json:"errors"`
}

// Hooks to install
var gitHooks = map[string]string{
	"pre-commit": `#!/bin/sh
# GitAI pre-commit hook
# Runs quality checks on tracked prompt files

gitai pre-commit-check
`,
	"post-commit": `#!/bin/sh
# GitAI post-commit hook
# Records intent after commit

gitai post-commit-record "$@"
`,
}

// InstallHooks installs GitAI hooks to .git/hooks/
func InstallHooks(gitRoot string) error {
	gitDir := filepath.Join(gitRoot, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")

	// Ensure hooks directory exists
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	// Install each hook
	for hookName, hookContent := range gitHooks {
		hookPath := filepath.Join(hooksDir, hookName)

		// Check if hook already exists
		if _, err := os.Stat(hookPath); err == nil {
			// Check if it's a GitAI hook
			content, _ := os.ReadFile(hookPath)
			if strings.Contains(string(content), "GitAI") {
				// Already installed, skip
				continue
			}

			// Back up existing hook
			backupPath := hookPath + ".backup"
			if err := os.Rename(hookPath, backupPath); err != nil {
				return fmt.Errorf("failed to backup existing hook: %w", err)
			}
		}

		// Write hook
		if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
			return fmt.Errorf("failed to write hook: %w", err)
		}
	}

	return nil
}

// UninstallHooks removes GitAI hooks from .git/hooks/
func UninstallHooks(gitRoot string) error {
	hooksDir := filepath.Join(gitRoot, ".git", "hooks")

	for hookName := range gitHooks {
		hookPath := filepath.Join(hooksDir, hookName)

		// Check if hook exists
		if _, err := os.Stat(hookPath); os.IsNotExist(err) {
			continue
		}

		// Check if it's a GitAI hook
		content, _ := os.ReadFile(hookPath)
		if !strings.Contains(string(content), "GitAI") {
			continue
		}

		// Remove hook
		if err := os.Remove(hookPath); err != nil {
			return fmt.Errorf("failed to remove hook: %w", err)
		}

		// Restore backup if exists
		backupPath := hookPath + ".backup"
		if _, err := os.Stat(backupPath); err == nil {
			os.Rename(backupPath, hookPath)
		}
	}

	return nil
}

// CheckPromptQuality checks the quality of a prompt file
func CheckPromptQuality(storage *Storage, filePath string) PromptQualityResult {
	result := PromptQualityResult{
		HasIssues: false,
		MinScore:  100,
		Warnings:  []string{},
		Errors:   []string{},
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		result.HasIssues = true
		result.MinScore = 0
		result.Errors = append(result.Errors, "Failed to read file")
		return result
	}

	// Parse prompt
	parsed := token.Parse(string(content))

	// Calculate quality scores
	clarity := calculateClarity(parsed, &result.Warnings)
	conciseness := calculateConciseness(parsed, len(string(content)), &result.Warnings)
	efficiency := calculateEfficiency(parsed, &result.Warnings)

	// Overall score
	overall := (clarity*0.4 + conciseness*0.3 + efficiency*0.3)
	result.MinScore = int(overall)

	// Check for issues
	if overall < 40 {
		result.HasIssues = true
		result.Errors = append(result.Errors, "Prompt quality is too poor (< 40)")
	} else if overall < 60 {
		result.HasIssues = true
	}

	return result
}

// RunPreCommitCheck runs quality checks on all staged prompt files
func RunPreCommitCheck(storage *Storage) error {
	// Get staged files
	staged, err := getStagedFiles()
	if err != nil {
		return err
	}

	// Filter to .prompt files
	var promptFiles []string
	for _, file := range staged {
		if strings.HasSuffix(file, ".prompt") {
			absPath, err := GetAbsolutePath(file)
			if err == nil {
				promptFiles = append(promptFiles, absPath)
			}
		}
	}

	if len(promptFiles) == 0 {
		return nil
	}

	// Check each prompt
	minScore := 100
	var issues []string

	for _, file := range promptFiles {
		result := CheckPromptQuality(storage, file)
		if result.MinScore < minScore {
			minScore = result.MinScore
		}

		if result.HasIssues {
			issues = append(issues, fmt.Sprintf("%s: score %d", filepath.Base(file), result.MinScore))
		}

		for _, warning := range result.Warnings {
			fmt.Printf("⚠️  %s: %s\n", filepath.Base(file), warning)
		}
	}

	// Fail if quality is too low
	if minScore < 40 {
		return fmt.Errorf("prompt quality too low (min score: %d)\nIssues:\n  %s",
			minScore, strings.Join(issues, "\n  "))
	}

	return nil
}
