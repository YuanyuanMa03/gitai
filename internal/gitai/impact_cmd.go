package gitai

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

// ImpactCmd analyzes the impact of a prompt change
func ImpactCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	args := c.Args()
	if args.Len() == 0 {
		return fmt.Errorf("prompt path required. Usage: gitai impact <prompt-file>")
	}

	promptPath := args.First()
	verbose := c.Bool("verbose")

	// Analyze impact
	impact, err := storage.AnalyzeImpact(promptPath)
	if err != nil {
		return fmt.Errorf("failed to analyze impact: %v", err)
	}

	// Display results
	displayImpactResult(impact, verbose)

	return nil
}

// displayImpactResult displays the impact analysis result
func displayImpactResult(impact *ImpactResult, verbose bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("📊 Impact Analysis: %s\n\n", impact.Prompt)

	if len(impact.AffectedFiles) == 0 {
		fmt.Println("No impact data found for this prompt.")
		fmt.Println("\nTip: Use 'gitai track <file>' to start tracking prompt dependencies.")
		return
	}

	fmt.Fprintf(w, "%s\t%s\n", "Affected Files:", fmt.Sprintf("%d files", len(impact.AffectedFiles)))
	for _, file := range impact.AffectedFiles {
		fmt.Fprintf(w, "  →\t%s\n", file)
	}
	w.Flush()

	if len(impact.RelatedPrompts) > 0 {
		fmt.Printf("\n⚠️  Related Prompts (%d)\n", len(impact.RelatedPrompts))
		fmt.Println("These prompts share affected files and may need review:")
		for _, prompt := range impact.RelatedPrompts {
			fmt.Printf("  • %s\n", prompt)
		}
	}

	if verbose {
		// Show detailed analysis
		fmt.Printf("\n📈 Detailed Analysis:\n")
		fmt.Printf("  Direct impact: %d files\n", len(impact.AffectedFiles))
		fmt.Printf("  Related prompts: %d\n", len(impact.RelatedPrompts))
		fmt.Printf("  Total potential impact: %d items\n", len(impact.AffectedFiles)+len(impact.RelatedPrompts))
	}
}

// ScanDependenciesCmd scans tracked prompts and auto-detects dependencies
func ScanDependenciesCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Get tracked prompts
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked prompts found.")
		return nil
	}

	fmt.Printf("Scanning %d prompts for dependencies...\n\n", len(tracked))

	updated := 0
	for _, tf := range tracked {
		// Read prompt content
		absPath, err := GetAbsolutePath(tf.Path)
		if err != nil {
			continue
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		// Auto-detect dependencies
		dep, err := storage.AutoDetectDependencies(tf.Path, string(content))
		if err != nil {
			continue
		}

		// Only save if we found something
		if len(dep.Affects) > 0 {
			if err := storage.SaveDependency(dep); err == nil {
				updated++
				fmt.Printf("✓ %s: %d files\n", tf.Path, len(dep.Affects))
			}
		}
	}

	fmt.Printf("\nDependency scan complete: %d prompts updated\n", updated)
	return nil
}

// ShowDependenciesCmd shows all tracked dependencies
func ShowDependenciesCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Get all dependencies
	deps, err := storage.ListDependencies()
	if err != nil {
		return fmt.Errorf("failed to list dependencies: %w", err)
	}

	if len(deps) == 0 {
		fmt.Println("No dependencies tracked.")
		fmt.Println("\nTip: Run 'gitai scan-deps' to auto-detect dependencies.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Dependencies (%d prompts)\n\n", len(deps))

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Prompt", "Affects", "Used By", "Last Used")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "------", "-------", "-------", "---------")

	for _, dep := range deps {
		shortPrompt := dep.PromptPath
		if len(shortPrompt) > 25 {
			shortPrompt = shortPrompt[:22] + ".."
		}

		affects := fmt.Sprintf("%d files", len(dep.Affects))
		usedBy := fmt.Sprintf("%d prompts", len(dep.UsedBy))
		lastUsed := dep.LastUsed.Format("2006-01-02")

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", shortPrompt, affects, usedBy, lastUsed)
	}

	w.Flush()

	return nil
}
