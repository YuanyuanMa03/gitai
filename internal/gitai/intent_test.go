package gitai

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// intentTestStorage creates a temporary storage for testing
func intentTestStorage(t *testing.T) (*Storage, string) {
	t.Helper()

	tmpDir := t.TempDir()
	gitaiDir := filepath.Join(tmpDir, ".gitai")

	// Create .gitai/ directory
	if err := os.MkdirAll(gitaiDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	// Create minimal tracked.json to satisfy IsInitialized
	trackedPath := filepath.Join(gitaiDir, "tracked.json")
	trackedData := []byte(`{"version":"0.1.0","files":[]}`)
	if err := os.WriteFile(trackedPath, trackedData, 0644); err != nil {
		t.Fatalf("failed to create tracked.json: %v", err)
	}

	storage := NewStorage(tmpDir)
	return storage, tmpDir
}

// TestIntentStorage_SaveAndLoad verifies intent can be saved and retrieved
func TestIntentStorage_SaveAndLoad(t *testing.T) {
	storage, _ := intentTestStorage(t)

	intent := Intent{
		ID:         "test-intent-1",
		CommitHash: "abc123",
		PromptHashes: []string{"prompt-1", "prompt-2"},
		Message:    "Fix user authentication bug",
		Category:   "fix",
		Timestamp:  time.Now(),
	}

	// Save intent
	err := storage.SaveIntent(&intent)
	if err != nil {
		t.Fatalf("failed to save intent: %v", err)
	}

	// Load intent
	loaded, err := storage.GetIntent(intent.ID)
	if err != nil {
		t.Fatalf("failed to load intent: %v", err)
	}

	// Verify
	if loaded.ID != intent.ID {
		t.Errorf("expected ID %s, got %s", intent.ID, loaded.ID)
	}
	if loaded.CommitHash != intent.CommitHash {
		t.Errorf("expected CommitHash %s, got %s", intent.CommitHash, loaded.CommitHash)
	}
	if loaded.Message != intent.Message {
		t.Errorf("expected Message %s, got %s", intent.Message, loaded.Message)
	}
	if loaded.Category != intent.Category {
		t.Errorf("expected Category %s, got %s", intent.Category, loaded.Category)
	}
	if len(loaded.PromptHashes) != len(intent.PromptHashes) {
		t.Errorf("expected %d prompt hashes, got %d", len(intent.PromptHashes), len(loaded.PromptHashes))
	}
}

// TestIntentStorage_ListIntents verifies multiple intents can be listed
func TestIntentStorage_ListIntents(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Save multiple intents
	intents := []Intent{
		{ID: "intent-1", CommitHash: "abc1", Message: "First", Category: "feature", Timestamp: time.Now()},
		{ID: "intent-2", CommitHash: "abc2", Message: "Second", Category: "fix", Timestamp: time.Now()},
		{ID: "intent-3", CommitHash: "abc3", Message: "Third", Category: "refactor", Timestamp: time.Now()},
	}

	for _, intent := range intents {
		if err := storage.SaveIntent(&intent); err != nil {
			t.Fatalf("failed to save intent: %v", err)
		}
	}

	// List all
	listed, err := storage.ListIntents()
	if err != nil {
		t.Fatalf("failed to list intents: %v", err)
	}

	if len(listed) != len(intents) {
		t.Errorf("expected %d intents, got %d", len(intents), len(listed))
	}
}

// TestIntentStorage_GetIntentByCommit verifies intent can be found by commit hash
func TestIntentStorage_GetIntentByCommit(t *testing.T) {
	storage, _ := intentTestStorage(t)

	intent := Intent{
		ID:         "test-intent",
		CommitHash: "commit-xyz",
		Message:    "Test message",
		Category:   "feature",
		Timestamp:  time.Now(),
	}

	if err := storage.SaveIntent(&intent); err != nil {
		t.Fatalf("failed to save intent: %v", err)
	}

	// Find by commit
	found, err := storage.GetIntentByCommit("commit-xyz")
	if err != nil {
		t.Fatalf("failed to get intent by commit: %v", err)
	}

	if found.ID != intent.ID {
		t.Errorf("expected ID %s, got %s", intent.ID, found.ID)
	}
}

// TestIntentStorage_DeleteIntent verifies intent can be deleted
func TestIntentStorage_DeleteIntent(t *testing.T) {
	storage, _ := intentTestStorage(t)

	intent := Intent{
		ID:         "to-delete",
		CommitHash: "abc",
		Message:    "Will be deleted",
		Category:   "test",
		Timestamp:  time.Now(),
	}

	if err := storage.SaveIntent(&intent); err != nil {
		t.Fatalf("failed to save intent: %v", err)
	}

	// Delete
	if err := storage.DeleteIntent(intent.ID); err != nil {
		t.Fatalf("failed to delete intent: %v", err)
	}

	// Verify deleted
	_, err := storage.GetIntent(intent.ID)
	if err == nil {
		t.Error("expected error when getting deleted intent, got nil")
	}
}
