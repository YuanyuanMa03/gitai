package gitai

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

// ExportFormat defines the export format type
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// ExportData represents the full export structure
type ExportData struct {
	Version  string        `json:"version"`
	RepoRoot string        `json:"repo_root"`
	Files    []TrackedFile `json:"files"`
	Config   *Config       `json:"config,omitempty"`
}

// ExportCmd exports tracked files to JSON or CSV
func ExportCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Parse options
	format := ExportFormat(c.String("format"))
	output := c.String("output")
	includeConfig := c.Bool("include-config")

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files to export.")
		return nil
	}

	// Generate output filename if not provided
	if output == "" {
		timestamp := getTimestamp()
		switch format {
		case FormatCSV:
			output = fmt.Sprintf("gitai-export-%s.csv", timestamp)
		default:
			output = fmt.Sprintf("gitai-export-%s.json", timestamp)
		}
	}

	// Export based on format
	switch format {
	case FormatCSV:
		return exportToCSV(tracked, output)
	case FormatJSON:
		return exportToJSON(tracked, root, output, includeConfig)
	default:
		return fmt.Errorf("unsupported format: %s (use json or csv)", format)
	}
}

// exportToJSON exports tracked files to JSON
func exportToJSON(tracked []TrackedFile, root, output string, includeConfig bool) error {
	data := ExportData{
		Version:  "1.0.0",
		RepoRoot: root,
		Files:    tracked,
	}

	// Include config if requested
	if includeConfig {
		config, err := LoadConfig(root)
		if err == nil {
			data.Config = config
		}
	}

	// Marshal to JSON with pretty print
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(output, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	fmt.Printf("Exported %d files to %s\n", len(tracked), output)
	return nil
}

// exportToCSV exports tracked files to CSV
func exportToCSV(tracked []TrackedFile, output string) error {
	// Create CSV file
	file, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	csvHeader := []string{
		"Path", "Role", "Tokens",
		"TaskTokens", "ConstraintsTokens", "ExamplesTokens",
		"InputTokens", "OutputTokens", "ContextTokens", "OtherTokens",
		"AddedAt", "LastModified",
	}
	if err := writer.Write(csvHeader); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, f := range tracked {
		row := []string{
			f.Path,
			f.Role,
			fmt.Sprintf("%d", f.Tokens),
			fmt.Sprintf("%d", f.TaskTokens),
			fmt.Sprintf("%d", f.ConstraintsTokens),
			fmt.Sprintf("%d", f.ExamplesTokens),
			fmt.Sprintf("%d", f.InputTokens),
			fmt.Sprintf("%d", f.OutputTokens),
			fmt.Sprintf("%d", f.ContextTokens),
			fmt.Sprintf("%d", f.OtherTokens),
			f.AddedAt.Format("2006-01-02 15:04:05"),
			f.LastModified.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	fmt.Printf("Exported %d files to %s\n", len(tracked), output)
	return nil
}

// ImportCmd imports tracked files from export file
func ImportCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Parse options
	input := c.String("input")
	merge := c.Bool("merge")
	force := c.Bool("force")

	if input == "" {
		return fmt.Errorf("input file required. Use --input <file>")
	}

	// Check file exists
	if _, err := os.Stat(input); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", input)
	}

	// Detect format by extension
	ext := filepath.Ext(input)
	var format ExportFormat
	switch ext {
	case ".csv":
		format = FormatCSV
	case ".json":
		format = FormatJSON
	default:
		return fmt.Errorf("unsupported file format: %s (use .json or .csv)", ext)
	}

	// Import based on format
	switch format {
	case FormatCSV:
		return importFromCSV(input, storage, merge, force)
	case FormatJSON:
		return importFromJSON(input, storage, root, merge, force)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// importFromJSON imports tracked files from JSON
func importFromJSON(input string, storage *Storage, root string, merge, force bool) error {
	// Read file
	data, err := os.ReadFile(input)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	// Parse JSON
	var exportData ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Get existing tracked files
	existing, _ := storage.ListAllTracked()
	existingMap := make(map[string]TrackedFile)
	for _, f := range existing {
		existingMap[f.Path] = f
	}

	added := 0
	updated := 0
	skipped := 0

	// Import files
	for _, f := range exportData.Files {
		// Check if file exists
		absPath, err := GetAbsolutePath(f.Path)
		if err != nil {
			fmt.Printf("Skipping invalid path: %s\n", f.Path)
			skipped++
			continue
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			fmt.Printf("Skipping non-existent file: %s\n", f.Path)
			skipped++
			continue
		}

		// Check if already tracked
		if _, exists := existingMap[f.Path]; exists {
			if !merge && !force {
				fmt.Printf("Skipping existing file (use --merge or --force): %s\n", f.Path)
				skipped++
				continue
			}

			// Update with force - remove and re-add
			if force {
				f.AddedAt = f.AddedAt // Keep original add time
				if err := storage.RemoveTrackedFile(f.Path); err != nil {
					// Ignore error if file doesn't exist
				}
				if err := storage.AddTrackedFile(f); err != nil {
					fmt.Printf("Failed to update %s: %v\n", f.Path, err)
					skipped++
					continue
				}
				updated++
			} else {
				skipped++
				continue
			}
		} else {
			// Add new
			if err := storage.AddTrackedFile(f); err != nil {
				fmt.Printf("Failed to track %s: %v\n", f.Path, err)
				skipped++
				continue
			}
			added++
		}
	}

	// Import config if present and requested
	if exportData.Config != nil && force {
		if err := SaveConfig(root, exportData.Config); err == nil {
			fmt.Printf("Imported configuration\n")
		}
	}

	fmt.Printf("\nImport complete: %d added, %d updated, %d skipped\n", added, updated, skipped)
	return nil
}

// importFromCSV imports tracked files from CSV
func importFromCSV(input string, storage *Storage, merge, force bool) error {
	// Open CSV file
	file, err := os.Open(input)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Get existing tracked files
	existing, _ := storage.ListAllTracked()
	existingMap := make(map[string]TrackedFile)
	for _, f := range existing {
		existingMap[f.Path] = f
	}

	added := 0
	updated := 0
	skipped := 0

	// Read data rows
	for {
		row, err := reader.Read()
		if err != nil {
			break
		}

		// Parse row (simple implementation)
		if len(row) < 3 {
			skipped++
			continue
		}

		path := row[0]
		role := row[1]

		// Check if file exists
		absPath, err := GetAbsolutePath(path)
		if err != nil {
			skipped++
			continue
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			skipped++
			continue
		}

		// Check if already tracked
		if _, exists := existingMap[path]; exists {
			if !merge && !force {
				skipped++
				continue
			}
			skipped++
			continue
		}

		// Add new - create TrackedFile from CSV data
		newFile := TrackedFile{
			Path: path,
			Role: role,
		}
		if err := storage.AddTrackedFile(newFile); err != nil {
			skipped++
			continue
		}
		added++
	}

	fmt.Printf("Import complete: %d added, %d updated, %d skipped\n", added, updated, skipped)
	return nil
}

// ListExportsCmd lists available export files in current directory
func ListExportsCmd(c *cli.Context) error {
	// Find export files
	matches, err := filepath.Glob("gitai-export-*.json")
	if err != nil {
		return err
	}

	csvMatches, err := filepath.Glob("gitai-export-*.csv")
	if err != nil {
		return err
	}

	matches = append(matches, csvMatches...)

	if len(matches) == 0 {
		fmt.Println("No export files found in current directory.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "%s\t%s\t%s\n", "File", "Format", "Size")
	fmt.Fprintf(w, "%s\t%s\t%s\n", "----", "------", "----")

	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		ext := filepath.Ext(match)
		format := "JSON"
		if ext == ".csv" {
			format = "CSV"
		}

		size := info.Size()
		sizeStr := fmt.Sprintf("%d B", size)
		if size > 1024 {
			sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
		}
		if size > 1024*1024 {
			sizeStr = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
		}

		fmt.Fprintf(w, "%s\t%s\t%s\n", match, format, sizeStr)
	}

	w.Flush()
	return nil
}
