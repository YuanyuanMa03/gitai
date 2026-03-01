package gitai

import (
	"testing"
	"time"
)

// TestGenerateInsights verifies insights generation
func TestGenerateInsights(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Add some sample data
	storage.SaveIntent(&Intent{
		ID:         "intent-1",
		CommitHash: "abc1",
		Message:    "fix: user login bug",
		Category:   "fix",
		Timestamp:  time.Now(),
	})

	storage.SaveFeedback(&Feedback{
		ID:         "fb-1",
		CommitHash: "abc1",
		Rating:     "good",
		Timestamp:  time.Now(),
	})

	// Generate insights
	insights, err := GenerateInsights(storage)
	if err != nil {
		t.Fatalf("failed to generate insights: %v", err)
	}

	if insights == nil {
		t.Error("insights should not be nil")
	}

	// Verify insights have required fields
	if insights.TotalCommits != 1 {
		t.Errorf("expected 1 commit, got %d", insights.TotalCommits)
	}
	if insights.TotalFeedback != 1 {
		t.Errorf("expected 1 feedback, got %d", insights.TotalFeedback)
	}
}

// TestTraceIntentToCode verifies intent tracing
func TestTraceIntentToCode(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Create intent with associated data
	intent := &Intent{
		ID:           "test-intent",
		CommitHash:   "commit-xyz",
		PromptHashes: []string{"api.prompt", "utils.prompt"},
		Message:      "feat: add authentication",
		Category:     "feature",
		Timestamp:    time.Now(),
	}
	storage.SaveIntent(intent)

	// Create dependencies
	storage.SaveDependency(&Dependency{
		PromptPath: "api.prompt",
		Affects:    []string{"api/auth.go", "api/middleware.go"},
	})

	storage.SaveDependency(&Dependency{
		PromptPath: "utils.prompt",
		Affects:    []string{"utils/helpers.go"},
	})

	// Trace intent
	trace, err := TraceIntent(storage, "test-intent")
	if err != nil {
		t.Fatalf("failed to trace intent: %v", err)
	}

	if trace.IntentID != "test-intent" {
		t.Errorf("expected intent ID test-intent, got %s", trace.IntentID)
	}
	if len(trace.Prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(trace.Prompts))
	}
	if len(trace.AffectedFiles) != 3 {
		t.Errorf("expected 3 affected files, got %d", len(trace.AffectedFiles))
	}
}

// TestTraceCodeToIntent verifies reverse tracing
func TestTraceCodeToIntent(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Set up: prompt affects file, intent uses prompt
	storage.SaveIntent(&Intent{
		ID:           "intent-1",
		PromptHashes: []string{"api.prompt"},
		Message:      "fix api bug",
		Category:     "fix",
	})

	storage.SaveDependency(&Dependency{
		PromptPath: "api.prompt",
		Affects:    []string{"api/handlers.go"},
	})

	// Trace file to intents
	intents, err := TraceCodeToIntent(storage, "api/handlers.go")
	if err != nil {
		t.Fatalf("failed to trace code: %v", err)
	}

	if len(intents) != 1 {
		t.Errorf("expected 1 intent, got %d", len(intents))
	}
	if intents[0] != "intent-1" {
		t.Errorf("expected intent-1, got %s", intents[0])
	}
}
