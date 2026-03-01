package gitai

import (
	"testing"
	"time"
)

// TestFeedbackStorage_SaveAndLoad verifies feedback can be saved and retrieved
func TestFeedbackStorage_SaveAndLoad(t *testing.T) {
	storage, _ := intentTestStorage(t)

	feedback := &Feedback{
		ID:         "feedback-1",
		CommitHash: "abc123",
		Rating:     "good",
		Reason:     "Prompt produced excellent results",
		Timestamp:  time.Now(),
	}

	// Save feedback
	err := storage.SaveFeedback(feedback)
	if err != nil {
		t.Fatalf("failed to save feedback: %v", err)
	}

	// Load feedback
	loaded, err := storage.GetFeedback(feedback.ID)
	if err != nil {
		t.Fatalf("failed to load feedback: %v", err)
	}

	// Verify
	if loaded.ID != feedback.ID {
		t.Errorf("expected ID %s, got %s", feedback.ID, loaded.ID)
	}
	if loaded.Rating != feedback.Rating {
		t.Errorf("expected Rating %s, got %s", feedback.Rating, loaded.Rating)
	}
	if loaded.Reason != feedback.Reason {
		t.Errorf("expected Reason %s, got %s", feedback.Reason, loaded.Reason)
	}
}

// TestFeedbackStorage_ListFeedback verifies all feedback can be listed
func TestFeedbackStorage_ListFeedback(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Save multiple feedback entries
	feedbacks := []*Feedback{
		{ID: "fb-1", CommitHash: "abc1", Rating: "good", Reason: "Works well", Timestamp: time.Now()},
		{ID: "fb-2", CommitHash: "abc2", Rating: "bad", Reason: "Has bugs", Timestamp: time.Now()},
		{ID: "fb-3", CommitHash: "abc3", Rating: "neutral", Reason: "Okay", Timestamp: time.Now()},
	}

	for _, fb := range feedbacks {
		if err := storage.SaveFeedback(fb); err != nil {
			t.Fatalf("failed to save feedback: %v", err)
		}
	}

	// List all
	listed, err := storage.ListFeedback()
	if err != nil {
		t.Fatalf("failed to list feedback: %v", err)
	}

	if len(listed) != len(feedbacks) {
		t.Errorf("expected %d feedback entries, got %d", len(feedbacks), len(listed))
	}
}

// TestGetFeedbackByCommit verifies feedback can be found by commit hash
func TestGetFeedbackByCommit(t *testing.T) {
	storage, _ := intentTestStorage(t)

	feedback := &Feedback{
		ID:         "test-fb",
		CommitHash: "commit-xyz",
		Rating:     "good",
		Timestamp:  time.Now(),
	}

	if err := storage.SaveFeedback(feedback); err != nil {
		t.Fatalf("failed to save feedback: %v", err)
	}

	// Find by commit
	found, err := storage.GetFeedbackByCommit("commit-xyz")
	if err != nil {
		t.Fatalf("failed to get feedback by commit: %v", err)
	}

	if found.ID != feedback.ID {
		t.Errorf("expected ID %s, got %s", feedback.ID, found.ID)
	}
}

// TestCalculateFeedbackStats verifies statistics calculation
func TestCalculateFeedbackStats(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Save feedback with different ratings
	feedbacks := []*Feedback{
		{ID: "fb-1", CommitHash: "a", Rating: "good", Timestamp: time.Now()},
		{ID: "fb-2", CommitHash: "b", Rating: "good", Timestamp: time.Now()},
		{ID: "fb-3", CommitHash: "c", Rating: "bad", Timestamp: time.Now()},
		{ID: "fb-4", CommitHash: "d", Rating: "neutral", Timestamp: time.Now()},
	}

	for _, fb := range feedbacks {
		if err := storage.SaveFeedback(fb); err != nil {
			t.Fatalf("failed to save feedback: %v", err)
		}
	}

	// Get stats
	stats, err := storage.GetFeedbackStats()
	if err != nil {
		t.Fatalf("failed to get feedback stats: %v", err)
	}

	if stats.Total != 4 {
		t.Errorf("expected total 4, got %d", stats.Total)
	}
	if stats.Good != 2 {
		t.Errorf("expected 2 good, got %d", stats.Good)
	}
	if stats.Bad != 1 {
		t.Errorf("expected 1 bad, got %d", stats.Bad)
	}
	if stats.Neutral != 1 {
		t.Errorf("expected 1 neutral, got %d", stats.Neutral)
	}
	// (2*100 + 1*0 + 1*50) / 4 = 250/4 = 62.5
	expectedScore := 62.5
	if stats.AverageScore < expectedScore-1 || stats.AverageScore > expectedScore+1 {
		t.Errorf("expected score %.1f, got %.1f", expectedScore, stats.AverageScore)
	}
}
