package gitai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	GitaiDirName = ".gitai"
	MetadataFile = "tracked.json"
	ConfigFile   = "config.toml"
)

// TrackedFile represents a tracked .prompt file
type TrackedFile struct {
	Path         string    `json:"path"`
	AddedAt      time.Time `json:"added_at"`
	LastModified time.Time `json:"last_modified"`
	CurrentHash  string    `json:"current_hash"`
	Role         string    `json:"role,omitempty"`
	Description  string    `json:"description,omitempty"`
	Tokens       int       `json:"tokens"`
	TokenCostUSD float64   `json:"token_cost_usd"`

	// Parsed sections from Day2
	Task        string `json:"task,omitempty"`
	Constraints string `json:"constraints,omitempty"`
	Examples    string `json:"examples,omitempty"`
	Input       string `json:"input,omitempty"`
	Output      string `json:"output,omitempty"`
	Context     string `json:"context,omitempty"`

	// Token breakdown by section
	TaskTokens        int `json:"task_tokens,omitempty"`
	ConstraintsTokens int `json:"constraints_tokens,omitempty"`
	ExamplesTokens    int `json:"examples_tokens,omitempty"`
	InputTokens       int `json:"input_tokens,omitempty"`
	OutputTokens      int `json:"output_tokens,omitempty"`
	ContextTokens     int `json:"context_tokens,omitempty"`
	OtherTokens       int `json:"other_tokens,omitempty"`
}

// TrackedFiles manages all tracked files
type TrackedFiles struct {
	Version string        `json:"version"`
	Files   []TrackedFile `json:"files"`
	mu      sync.RWMutex
}

// Storage handles .gitai/ directory operations
type Storage struct {
	gitRoot  string
	gitaiDir string
}

// NewStorage creates a new storage instance
func NewStorage(gitRoot string) *Storage {
	gitaiDir := filepath.Join(gitRoot, GitaiDirName)
	return &Storage{
		gitRoot:  gitRoot,
		gitaiDir: gitaiDir,
	}
}

// Initialize creates .gitai/ directory and initial metadata
func (s *Storage) Initialize() error {
	if _, err := os.Stat(s.gitaiDir); err == nil {
		return fmt.Errorf(".gitai/ already exists")
	}

	// Create .gitai/ directory
	if err := os.MkdirAll(s.gitaiDir, 0755); err != nil {
		return fmt.Errorf("failed to create .gitai/: %w", err)
	}

	// Create initial tracked.json
	tracked := &TrackedFiles{
		Version: "0.1.0",
		Files:   []TrackedFile{},
	}

	if err := s.SaveTracked(tracked); err != nil {
		return fmt.Errorf("failed to save tracked.json: %w", err)
	}

	// Create empty config.toml
	configPath := filepath.Join(s.gitaiDir, ConfigFile)
	if err := os.WriteFile(configPath, []byte("# GitAI Configuration\n"), 0644); err != nil {
		return fmt.Errorf("failed to create config.toml: %w", err)
	}

	// Create .gitignore for .gitai/ itself
	gitignorePath := filepath.Join(s.gitaiDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*\n"), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	return nil
}

// IsInitialized checks if .gitai/ exists
func (s *Storage) IsInitialized() bool {
	_, err := os.Stat(s.gitaiDir)
	return err == nil
}

// GetTrackedFiles loads tracked files from disk
func (s *Storage) GetTrackedFiles() (*TrackedFiles, error) {
	data, err := os.ReadFile(filepath.Join(s.gitaiDir, MetadataFile))
	if err != nil {
		if os.IsNotExist(err) {
			return &TrackedFiles{
				Version: "0.1.0",
				Files:   []TrackedFile{},
			}, nil
		}
		return nil, err
	}

	var tracked TrackedFiles
	if err := json.Unmarshal(data, &tracked); err != nil {
		return nil, err
	}

	return &tracked, nil
}

// SaveTracked saves tracked files to disk
func (s *Storage) SaveTracked(tracked *TrackedFiles) error {
	data, err := json.MarshalIndent(tracked, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(s.gitaiDir, MetadataFile), data, 0644)
}

// AddTrackedFile adds a new tracked file
func (s *Storage) AddTrackedFile(file TrackedFile) error {
	tracked, err := s.GetTrackedFiles()
	if err != nil {
		return err
	}

	// Check if already tracked
	for i, f := range tracked.Files {
		if f.Path == file.Path {
			// Update existing
			tracked.Files[i] = file
			return s.SaveTracked(tracked)
		}
	}

	// Add new
	tracked.Files = append(tracked.Files, file)
	return s.SaveTracked(tracked)
}

// RemoveTrackedFile removes a tracked file
func (s *Storage) RemoveTrackedFile(path string) error {
	tracked, err := s.GetTrackedFiles()
	if err != nil {
		return err
	}

	newFiles := []TrackedFile{}
	for _, f := range tracked.Files {
		if f.Path != path {
			newFiles = append(newFiles, f)
		}
	}

	tracked.Files = newFiles
	return s.SaveTracked(tracked)
}

// GetTrackedFileByPath gets a tracked file by path
func (s *Storage) GetTrackedFileByPath(path string) (*TrackedFile, error) {
	tracked, err := s.GetTrackedFiles()
	if err != nil {
		return nil, err
	}

	for _, f := range tracked.Files {
		if f.Path == path {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("file not tracked: %s", path)
}

// ListAllTracked returns all tracked files
func (s *Storage) ListAllTracked() ([]TrackedFile, error) {
	tracked, err := s.GetTrackedFiles()
	if err != nil {
		return nil, err
	}
	return tracked.Files, nil
}
