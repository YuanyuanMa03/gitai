package gitai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHooksInstall verifies hooks can be installed
func TestHooksInstall(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	gitaiDir := filepath.Join(tmpDir, ".gitai")
	hooksDir := filepath.Join(gitDir, "hooks")

	// Create git directory structure
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("failed to create test git dir: %v", err)
	}
	if err := os.MkdirAll(gitaiDir, 0755); err != nil {
		t.Fatalf("failed to create test gitai dir: %v", err)
	}

	// Install hooks
	err := InstallHooks(tmpDir)
	if err != nil {
		t.Fatalf("failed to install hooks: %v", err)
	}

	// Verify pre-commit hook exists
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	if _, err := os.Stat(preCommitPath); os.IsNotExist(err) {
		t.Error("pre-commit hook was not installed")
	}

	// Verify pre-commit hook is executable
	info, _ := os.Stat(preCommitPath)
	if info.Mode().Perm()&0111 == 0 {
		t.Error("pre-commit hook is not executable")
	}

	// Verify hook contains gitai command
	content, _ := os.ReadFile(preCommitPath)
	contentStr := string(content)
	if !strings.Contains(contentStr, "gitai") {
		t.Error("pre-commit hook does not contain gitai command")
	}
}

// TestHooksUninstall verifies hooks can be removed
func TestHooksUninstall(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	hooksDir := filepath.Join(gitDir, "hooks")

	// Create git directory structure
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("failed to create test git dir: %v", err)
	}

	// Create a fake gitai hook
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	hookContent := []byte("#!/bin/sh\n# GitAI hook\n")
	if err := os.WriteFile(preCommitPath, hookContent, 0755); err != nil {
		t.Fatalf("failed to create test hook: %v", err)
	}

	// Uninstall hooks
	err := UninstallHooks(tmpDir)
	if err != nil {
		t.Fatalf("failed to uninstall hooks: %v", err)
	}

	// Verify hook was removed
	if _, err := os.Stat(preCommitPath); !os.IsNotExist(err) {
		t.Error("pre-commit hook was not removed")
	}
}

// TestPreCommitCheck verifies pre-commit quality check
func TestPreCommitCheck(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Add a tracked file with good quality
	goodPrompt := TrackedFile{
		Path: "good.prompt",
		Role: "code-reviewer",
		Tokens: 250,
		TaskTokens: 100,
		ConstraintsTokens: 50,
		ExamplesTokens: 50,
		InputTokens: 25,
		OutputTokens: 25,
	}
	storage.AddTrackedFile(goodPrompt)

	// Create a temp file with good prompt content
	tmpDir := t.TempDir()
	goodFile := filepath.Join(tmpDir, "good.prompt")
	goodContent := `# role: code-reviewer

## Task:
Review the code for bugs and improvements.

## Constraints:
- Follow best practices
- Check for security issues

## Examples:
Input: function with bug
Output: fixed function

## Input:
Code to review

## Output:
Review comments
`
	if err := os.WriteFile(goodFile, []byte(goodContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Check should pass for good prompt
	result := CheckPromptQuality(storage, goodFile)
	if result.HasIssues {
		t.Error("good prompt should not have issues")
	}
	if result.MinScore < 60 {
		t.Errorf("good prompt should have score >= 60, got %d", result.MinScore)
	}
}

// TestPromptQualityResult verifies quality check result structure
func TestPromptQualityResult(t *testing.T) {
	result := PromptQualityResult{
		HasIssues: true,
		MinScore:  45,
		Warnings:  []string{"Missing examples"},
		Errors:    []string{"No task defined"},
	}

	if !result.HasIssues {
		t.Error("expected HasIssues to be true")
	}
	if result.MinScore != 45 {
		t.Errorf("expected MinScore 45, got %d", result.MinScore)
	}
	if len(result.Warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(result.Warnings))
	}
	if len(result.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(result.Errors))
	}
}
