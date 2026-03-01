package gitai

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// IntentsFile holds the intents data filename
const IntentsFile = "intents.json"

// IntentData stores all intents
type IntentData struct {
	Version string            `json:"version"`
	Intents map[string]*Intent `json:"intents"` // id -> intent
	mu      sync.RWMutex
}

// Intent represents a commit's intent
type Intent struct {
	ID            string    `json:"id"`
	CommitHash    string    `json:"commit_hash"`
	PromptHashes  []string  `json:"prompt_hashes"`
	Message       string    `json:"message"`
	Category      string    `json:"category"` // fix/feature/refactor/ai-gen/experimental
	EstimatedCost float64   `json:"estimated_cost"`
	ActualCost    float64   `json:"actual_cost"`
	Timestamp     time.Time `json:"timestamp"`
}

// LoadIntents loads intents from disk
func (s *Storage) LoadIntents() (*IntentData, error) {
	intentsPath := filepath.Join(s.gitaiDir, IntentsFile)

	// If file doesn't exist, return empty intents
	if _, err := os.Stat(intentsPath); os.IsNotExist(err) {
		return &IntentData{
			Version: "1.0.0",
			Intents: make(map[string]*Intent),
		}, nil
	}

	data, err := os.ReadFile(intentsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read intents file: %w", err)
	}

	var intents IntentData
	if err := json.Unmarshal(data, &intents); err != nil {
		return nil, fmt.Errorf("failed to parse intents: %w", err)
	}

	if intents.Intents == nil {
		intents.Intents = make(map[string]*Intent)
	}

	return &intents, nil
}

// SaveIntents saves intents to disk
func (s *Storage) SaveIntents(intents *IntentData) error {
	intents.mu.Lock()
	defer intents.mu.Unlock()

	intentsPath := filepath.Join(s.gitaiDir, IntentsFile)

	data, err := json.MarshalIndent(intents, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal intents: %w", err)
	}

	if err := os.WriteFile(intentsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write intents file: %w", err)
	}

	return nil
}

// SaveIntent saves a single intent
func (s *Storage) SaveIntent(intent *Intent) error {
	intents, err := s.LoadIntents()
	if err != nil {
		return err
	}

	intents.mu.Lock()
	intents.Intents[intent.ID] = intent
	intents.mu.Unlock()

	return s.SaveIntents(intents)
}

// GetIntent retrieves an intent by ID
func (s *Storage) GetIntent(id string) (*Intent, error) {
	intents, err := s.LoadIntents()
	if err != nil {
		return nil, err
	}

	intents.mu.RLock()
	defer intents.mu.RUnlock()

	intent, exists := intents.Intents[id]
	if !exists {
		return nil, fmt.Errorf("intent not found: %s", id)
	}

	return intent, nil
}

// ListIntents returns all intents
func (s *Storage) ListIntents() ([]*Intent, error) {
	intents, err := s.LoadIntents()
	if err != nil {
		return nil, err
	}

	intents.mu.RLock()
	defer intents.mu.RUnlock()

	result := make([]*Intent, 0, len(intents.Intents))
	for _, intent := range intents.Intents {
		result = append(result, intent)
	}

	return result, nil
}

// GetIntentByCommit finds an intent by commit hash
func (s *Storage) GetIntentByCommit(commitHash string) (*Intent, error) {
	intents, err := s.LoadIntents()
	if err != nil {
		return nil, err
	}

	intents.mu.RLock()
	defer intents.mu.RUnlock()

	for _, intent := range intents.Intents {
		if intent.CommitHash == commitHash {
			return intent, nil
		}
	}

	return nil, fmt.Errorf("intent not found for commit: %s", commitHash)
}

// DeleteIntent removes an intent
func (s *Storage) DeleteIntent(id string) error {
	intents, err := s.LoadIntents()
	if err != nil {
		return err
	}

	intents.mu.Lock()
	delete(intents.Intents, id)
	intents.mu.Unlock()

	return s.SaveIntents(intents)
}

// GenerateIntentID generates a unique intent ID
func GenerateIntentID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "intent-" + hex.EncodeToString(b)
}

// DetectIntentCategory detects category from commit message
func DetectIntentCategory(message string) string {
	msg := strings.ToLower(strings.TrimSpace(message))

	// Check for prefixes first (order matters - longer prefixes first)
	prefixes := []struct {
		prefix   string
		category string
	}{
		{"experimental:", "experimental"},
		{"experiment:", "experimental"},
		{"ai-gen:", "ai-gen"},
		{"generated:", "ai-gen"},
		{"ai:", "ai-gen"},
		{"refactor:", "refactor"},
		{"refact:", "refactor"},
		{"feature:", "feature"},
		{"feat:", "feature"},
		{"fix:", "fix"},
		{"bug:", "fix"},
	}

	for _, p := range prefixes {
		if strings.HasPrefix(msg, p.prefix) {
			return p.category
		}
	}

	// Check for keywords (only if no prefix matched)
	keywords := []struct {
		keyword  string
		category string
	}{
		{"fix", "fix"},
		{"bug", "fix"},
		{"feat", "feature"},
		{"refactor", "refactor"},
		{"cleanup", "refactor"},
		{"rewrite", "refactor"},
	}

	for _, kw := range keywords {
		if strings.Contains(msg, kw.keyword) {
			return kw.category
		}
	}

	return "other"
}
