package gitai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mayuanyuan/gitai/pkg/token"
)

// DependenciesFile holds the dependencies data filename
const DependenciesFile = "dependencies.json"

// DependencyData stores all dependencies
type DependencyData struct {
	Version       string                 `json:"version"`
	Dependencies  map[string]*Dependency `json:"dependencies"` // prompt_path -> dependency
	mu            sync.RWMutex
}

// Dependency represents a prompt's dependency relationships
type Dependency struct {
	PromptPath string    `json:"prompt_path"`
	Affects    []string  `json:"affects"`     // Files this prompt affects
	UsedBy     []string  `json:"used_by"`     // Other prompts that use this prompt
	LastUsed   time.Time `json:"last_used"`
}

// ImpactResult represents the result of an impact analysis
type ImpactResult struct {
	Prompt         string   `json:"prompt"`
	AffectedFiles  []string `json:"affected_files"`
	DependentFiles []string `json:"dependent_files"` // Files that import/use affected files
	RelatedPrompts []string `json:"related_prompts"` // Other prompts that might be affected
}

// LoadDependencies loads dependencies from disk
func (s *Storage) LoadDependencies() (*DependencyData, error) {
	depsPath := filepath.Join(s.gitaiDir, DependenciesFile)

	// If file doesn't exist, return empty dependencies
	if _, err := os.Stat(depsPath); os.IsNotExist(err) {
		return &DependencyData{
			Version:      "1.0.0",
			Dependencies: make(map[string]*Dependency),
		}, nil
	}

	data, err := os.ReadFile(depsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read dependencies file: %w", err)
	}

	var deps DependencyData
	if err := json.Unmarshal(data, &deps); err != nil {
		return nil, fmt.Errorf("failed to parse dependencies: %w", err)
	}

	if deps.Dependencies == nil {
		deps.Dependencies = make(map[string]*Dependency)
	}

	return &deps, nil
}

// SaveDependencies saves dependencies to disk
func (s *Storage) SaveDependencies(deps *DependencyData) error {
	deps.mu.Lock()
	defer deps.mu.Unlock()

	depsPath := filepath.Join(s.gitaiDir, DependenciesFile)

	data, err := json.MarshalIndent(deps, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal dependencies: %w", err)
	}

	if err := os.WriteFile(depsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write dependencies file: %w", err)
	}

	return nil
}

// SaveDependency saves a single dependency
func (s *Storage) SaveDependency(dep *Dependency) error {
	deps, err := s.LoadDependencies()
	if err != nil {
		return err
	}

	dep.LastUsed = time.Now()
	deps.mu.Lock()
	deps.Dependencies[dep.PromptPath] = dep
	deps.mu.Unlock()

	return s.SaveDependencies(deps)
}

// GetDependency retrieves a dependency by prompt path
func (s *Storage) GetDependency(promptPath string) (*Dependency, error) {
	deps, err := s.LoadDependencies()
	if err != nil {
		return nil, err
	}

	deps.mu.RLock()
	defer deps.mu.RLock()

	dep, exists := deps.Dependencies[promptPath]
	if !exists {
		return nil, fmt.Errorf("dependency not found: %s", promptPath)
	}

	return dep, nil
}

// ListDependencies returns all dependencies
func (s *Storage) ListDependencies() ([]*Dependency, error) {
	deps, err := s.LoadDependencies()
	if err != nil {
		return nil, err
	}

	deps.mu.RLock()
	defer deps.mu.RUnlock()

	result := make([]*Dependency, 0, len(deps.Dependencies))
	for _, dep := range deps.Dependencies {
		result = append(result, dep)
	}

	return result, nil
}

// DeleteDependency removes a dependency
func (s *Storage) DeleteDependency(promptPath string) error {
	deps, err := s.LoadDependencies()
	if err != nil {
		return err
	}

	deps.mu.Lock()
	delete(deps.Dependencies, promptPath)
	deps.mu.Unlock()

	return s.SaveDependencies(deps)
}

// AnalyzeImpact analyzes the impact of changing a prompt
func (s *Storage) AnalyzeImpact(promptPath string) (*ImpactResult, error) {
	deps, err := s.LoadDependencies()
	if err != nil {
		return nil, err
	}

	deps.mu.RLock()
	defer deps.mu.RUnlock()

	dep, exists := deps.Dependencies[promptPath]
	if !exists {
		return &ImpactResult{
			Prompt:         promptPath,
			AffectedFiles:  []string{},
			DependentFiles: []string{},
			RelatedPrompts: []string{},
		}, nil
	}

	result := &ImpactResult{
		Prompt:         promptPath,
		AffectedFiles:  dep.Affects,
		DependentFiles: []string{},
		RelatedPrompts: []string{},
	}

	// Find related prompts (prompts that share affected files)
	for _, otherDep := range deps.Dependencies {
		if otherDep.PromptPath == promptPath {
			continue
		}

		// Check if they share any affected files
		for _, file := range dep.Affects {
			for _, otherFile := range otherDep.Affects {
				if file == otherFile {
					result.RelatedPrompts = append(result.RelatedPrompts, otherDep.PromptPath)
					break
				}
			}
		}
	}

	return result, nil
}

// FindPromptsAffecting finds all prompts that affect a given file
func (s *Storage) FindPromptsAffecting(filePath string) ([]string, error) {
	deps, err := s.LoadDependencies()
	if err != nil {
		return nil, err
	}

	deps.mu.RLock()
	defer deps.mu.RUnlock()

	var prompts []string
	for _, dep := range deps.Dependencies {
		for _, affectedFile := range dep.Affects {
			if affectedFile == filePath {
				prompts = append(prompts, dep.PromptPath)
				break
			}
		}
	}

	return prompts, nil
}

// AutoDetectDependencies automatically detects dependencies from prompt content
func (s *Storage) AutoDetectDependencies(promptPath string, content string) (*Dependency, error) {
	dep := &Dependency{
		PromptPath: promptPath,
		Affects:    []string{},
		UsedBy:     []string{},
		LastUsed:   time.Now(),
	}

	// Parse the prompt to find file references
	parsed := token.Parse(content)

	// Look for file paths in examples and task sections
	// This is a simple heuristic - can be enhanced with more sophisticated parsing
	sections := []string{parsed.Task, parsed.Examples, parsed.Context}

	for _, section := range sections {
		// Look for common file patterns
		filePatterns := []string{
			".go", ".js", ".ts", ".py", ".java", ".cpp", ".c",
			".prompt", ".md", ".json", ".yaml", ".yml",
		}

		for _, pattern := range filePatterns {
			// Simple extraction - in production, use proper parsing
			if strings.Contains(section, pattern) {
				// Extract file paths (simplified)
				words := strings.Fields(section)
				for _, word := range words {
					word = strings.Trim(word, ".,;:()[]{}\"'")
					if strings.Contains(word, pattern) {
						// Avoid duplicates
						found := false
						for _, existing := range dep.Affects {
							if existing == word {
								found = true
								break
							}
						}
						if !found {
							dep.Affects = append(dep.Affects, word)
						}
					}
				}
			}
		}
	}

	return dep, nil
}
