package gitai

import (
	"testing"
)

// TestDependencyStorage_SaveAndLoad verifies dependency can be saved and retrieved
func TestDependencyStorage_SaveAndLoad(t *testing.T) {
	storage, _ := intentTestStorage(t)

	dep := Dependency{
		PromptPath: "test.prompt",
		Affects:    []string{"src/main.go", "src/utils.go"},
		UsedBy:     []string{"other.prompt"},
	}

	// Save dependency
	err := storage.SaveDependency(&dep)
	if err != nil {
		t.Fatalf("failed to save dependency: %v", err)
	}

	// Load dependency
	loaded, err := storage.GetDependency("test.prompt")
	if err != nil {
		t.Fatalf("failed to load dependency: %v", err)
	}

	// Verify
	if loaded.PromptPath != dep.PromptPath {
		t.Errorf("expected PromptPath %s, got %s", dep.PromptPath, loaded.PromptPath)
	}
	if len(loaded.Affects) != len(dep.Affects) {
		t.Errorf("expected %d affects, got %d", len(dep.Affects), len(loaded.Affects))
	}
	if len(loaded.UsedBy) != len(dep.UsedBy) {
		t.Errorf("expected %d used_by, got %d", len(dep.UsedBy), len(loaded.UsedBy))
	}
}

// TestDependencyStorage_ListDependencies verifies all dependencies can be listed
func TestDependencyStorage_ListDependencies(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Save multiple dependencies
	deps := []Dependency{
		{PromptPath: "a.prompt", Affects: []string{"file1.go"}},
		{PromptPath: "b.prompt", Affects: []string{"file2.go"}},
		{PromptPath: "c.prompt", Affects: []string{"file3.go"}},
	}

	for _, dep := range deps {
		if err := storage.SaveDependency(&dep); err != nil {
			t.Fatalf("failed to save dependency: %v", err)
		}
	}

	// List all
	listed, err := storage.ListDependencies()
	if err != nil {
		t.Fatalf("failed to list dependencies: %v", err)
	}

	if len(listed) != len(deps) {
		t.Errorf("expected %d dependencies, got %d", len(deps), len(listed))
	}
}

// TestDependencyStorage_DeleteDependency verifies dependency can be deleted
func TestDependencyStorage_DeleteDependency(t *testing.T) {
	storage, _ := intentTestStorage(t)

	dep := Dependency{
		PromptPath: "to-delete.prompt",
		Affects:    []string{"file.go"},
	}

	if err := storage.SaveDependency(&dep); err != nil {
		t.Fatalf("failed to save dependency: %v", err)
	}

	// Delete
	if err := storage.DeleteDependency("to-delete.prompt"); err != nil {
		t.Fatalf("failed to delete dependency: %v", err)
	}

	// Verify deleted
	_, err := storage.GetDependency("to-delete.prompt")
	if err == nil {
		t.Error("expected error when getting deleted dependency, got nil")
	}
}

// TestAnalyzeImpact verifies impact analysis works
func TestAnalyzeImpact(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Set up dependencies
	storage.SaveDependency(&Dependency{
		PromptPath: "api.prompt",
		Affects:    []string{"api/handlers.go", "api/routes.go"},
	})
	storage.SaveDependency(&Dependency{
		PromptPath: "utils.prompt",
		Affects:    []string{"utils/helpers.go"},
	})

	// Analyze impact of changing api.prompt
	impact, err := storage.AnalyzeImpact("api.prompt")
	if err != nil {
		t.Fatalf("failed to analyze impact: %v", err)
	}

	if len(impact.AffectedFiles) != 2 {
		t.Errorf("expected 2 affected files, got %d", len(impact.AffectedFiles))
	}
}

// TestFindDependentPrompts finds prompts that depend on a file
func TestFindDependentPrompts(t *testing.T) {
	storage, _ := intentTestStorage(t)

	// Set up dependencies
	storage.SaveDependency(&Dependency{
		PromptPath: "a.prompt",
		Affects:    []string{"common.go"},
	})
	storage.SaveDependency(&Dependency{
		PromptPath: "b.prompt",
		Affects:    []string{"common.go"},
	})

	// Find prompts affecting common.go
	prompts, err := storage.FindPromptsAffecting("common.go")
	if err != nil {
		t.Fatalf("failed to find prompts: %v", err)
	}

	if len(prompts) != 2 {
		t.Errorf("expected 2 prompts, got %d", len(prompts))
	}
}
