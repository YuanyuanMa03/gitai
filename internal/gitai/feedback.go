package gitai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FeedbackFile holds the feedback data filename
const FeedbackFile = "feedback.json"

// FeedbackData stores all feedback
type FeedbackData struct {
	Version   string                `json:"version"`
	Feedback  map[string]*Feedback  `json:"feedback"` // id -> feedback
	mu        sync.RWMutex
}

// Feedback represents user feedback on a commit/intent
type Feedback struct {
	ID         string    `json:"id"`
	CommitHash string    `json:"commit_hash"`
	IntentID   string    `json:"intent_id,omitempty"`
	Rating     string    `json:"rating"`     // good/bad/neutral
	Reason     string    `json:"reason"`     // User-provided reason
	Tags       []string  `json:"tags"`       // Optional tags for categorization
	Timestamp  time.Time `json:"timestamp"`
}

// FeedbackStats represents aggregated feedback statistics
type FeedbackStats struct {
	Total        int     `json:"total"`
	Good         int     `json:"good"`
	Bad          int     `json:"bad"`
	Neutral      int     `json:"neutral"`
	AverageScore float64 `json:"average_score"` // 0-100 scale
}

// LoadFeedback loads feedback from disk
func (s *Storage) LoadFeedback() (*FeedbackData, error) {
	fbPath := filepath.Join(s.gitaiDir, FeedbackFile)

	// If file doesn't exist, return empty feedback
	if _, err := os.Stat(fbPath); os.IsNotExist(err) {
		return &FeedbackData{
			Version:  "1.0.0",
			Feedback: make(map[string]*Feedback),
		}, nil
	}

	data, err := os.ReadFile(fbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read feedback file: %w", err)
	}

	var fb FeedbackData
	if err := json.Unmarshal(data, &fb); err != nil {
		return nil, fmt.Errorf("failed to parse feedback: %w", err)
	}

	if fb.Feedback == nil {
		fb.Feedback = make(map[string]*Feedback)
	}

	return &fb, nil
}

// SaveFeedback saves feedback to disk
func (s *Storage) SaveFeedbackData(fb *FeedbackData) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	fbPath := filepath.Join(s.gitaiDir, FeedbackFile)

	data, err := json.MarshalIndent(fb, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal feedback: %w", err)
	}

	if err := os.WriteFile(fbPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write feedback file: %w", err)
	}

	return nil
}

// SaveFeedback saves a single feedback entry
func (s *Storage) SaveFeedback(feedback *Feedback) error {
	fb, err := s.LoadFeedback()
	if err != nil {
		return err
	}

	fb.mu.Lock()
	fb.Feedback[feedback.ID] = feedback
	fb.mu.Unlock()

	return s.SaveFeedbackData(fb)
}

// GetFeedback retrieves feedback by ID
func (s *Storage) GetFeedback(id string) (*Feedback, error) {
	fb, err := s.LoadFeedback()
	if err != nil {
		return nil, err
	}

	fb.mu.RLock()
	defer fb.mu.RLock()

	f, exists := fb.Feedback[id]
	if !exists {
		return nil, fmt.Errorf("feedback not found: %s", id)
	}

	return f, nil
}

// ListFeedback returns all feedback
func (s *Storage) ListFeedback() ([]*Feedback, error) {
	fb, err := s.LoadFeedback()
	if err != nil {
		return nil, err
	}

	fb.mu.RLock()
	defer fb.mu.RUnlock()

	result := make([]*Feedback, 0, len(fb.Feedback))
	for _, f := range fb.Feedback {
		result = append(result, f)
	}

	return result, nil
}

// GetFeedbackByCommit finds feedback by commit hash
func (s *Storage) GetFeedbackByCommit(commitHash string) (*Feedback, error) {
	fb, err := s.LoadFeedback()
	if err != nil {
		return nil, err
	}

	fb.mu.RLock()
	defer fb.mu.RUnlock()

	for _, f := range fb.Feedback {
		if f.CommitHash == commitHash {
			return f, nil
		}
	}

	return nil, fmt.Errorf("feedback not found for commit: %s", commitHash)
}

// GetFeedbackStats calculates feedback statistics
func (s *Storage) GetFeedbackStats() (*FeedbackStats, error) {
	fb, err := s.LoadFeedback()
	if err != nil {
		return nil, err
	}

	fb.mu.RLock()
	defer fb.mu.RUnlock()

	stats := &FeedbackStats{
		Total: len(fb.Feedback),
	}

	totalScore := 0.0

	for _, f := range fb.Feedback {
		switch f.Rating {
		case "good":
			stats.Good++
			totalScore += 100.0
		case "bad":
			stats.Bad++
			totalScore += 0.0
		case "neutral":
			stats.Neutral++
			totalScore += 50.0
		}
	}

	if stats.Total > 0 {
		stats.AverageScore = totalScore / float64(stats.Total)
	}

	return stats, nil
}

// DeleteFeedback removes a feedback entry
func (s *Storage) DeleteFeedback(id string) error {
	fb, err := s.LoadFeedback()
	if err != nil {
		return err
	}

	fb.mu.Lock()
	delete(fb.Feedback, id)
	fb.mu.Unlock()

	return s.SaveFeedbackData(fb)
}
