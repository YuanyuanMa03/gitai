package gitai

import (
	"testing"
)

// TestGenerateIntentID verifies intent IDs are generated correctly
func TestGenerateIntentID(t *testing.T) {
	id1 := GenerateIntentID()
	id2 := GenerateIntentID()

	if id1 == "" {
		t.Error("generated ID is empty")
	}
	if id2 == "" {
		t.Error("generated ID is empty")
	}
	if id1 == id2 {
		t.Error("generated IDs should be unique")
	}
}

// TestDetectIntentCategory verifies category detection from commit message
func TestDetectIntentCategory(t *testing.T) {
	tests := []struct {
		message  string
		expected string
	}{
		{"fix: user login bug", "fix"},
		{"feat: add dark mode", "feature"},
		{"refactor: cleanup utils", "refactor"},
		{"ai: generate component", "ai-gen"},
		{"experimental: try new approach", "experimental"},
		{"random message", "other"},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			category := DetectIntentCategory(tt.message)
			if category != tt.expected {
				t.Errorf("expected category %s, got %s", tt.expected, category)
			}
		})
	}
}
