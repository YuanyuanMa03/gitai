package gitai

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

// SearchResult represents a search result
type SearchResult struct {
	File      string
	Role      string
	Tokens    int
	Matches   []string
	Snippets  []string
	Score     float64
}

// SearchCmd searches across tracked prompt files
func SearchCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Parse options
	query := c.Args().First()
	if query == "" {
		return fmt.Errorf("search query required. Usage: gitai search <query>")
	}

	role := c.String("role")
	maxResults := c.Int("max")
	contextLines := c.Int("context")
	regexMode := c.Bool("regex")
	caseSensitive := c.Bool("case-sensitive")

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files to search.")
		return nil
	}

	// Filter by role if specified
	var filesToSearch []TrackedFile
	if role != "" {
		for _, f := range tracked {
			if strings.EqualFold(f.Role, role) {
				filesToSearch = append(filesToSearch, f)
			}
		}
		if len(filesToSearch) == 0 {
			fmt.Printf("No files found with role: %s\n", role)
			return nil
		}
	} else {
		filesToSearch = tracked
	}

	// Perform search
	results, err := performSearch(filesToSearch, query, regexMode, caseSensitive, contextLines)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(results) == 0 {
		fmt.Printf("No matches found for: %s\n", query)
		return nil
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit results
	if maxResults > 0 && len(results) > maxResults {
		results = results[:maxResults]
	}

	// Display results
	return displaySearchResults(results, query, contextLines > 0)
}

// performSearch executes the search across files
func performSearch(files []TrackedFile, query string, regexMode, caseSensitive bool, contextLines int) ([]SearchResult, error) {
	var results []SearchResult

	for _, f := range files {
		// Read file content
		absPath, err := GetAbsolutePath(f.Path)
		if err != nil {
			continue
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		contentStr := string(content)

		// Find matches
		var matches []string
		var snippets []string
		score := 0.0

		if regexMode {
			// Regex search
			pattern := query
			if !caseSensitive {
				pattern = "(?i)" + pattern
			}
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}

			matches = re.FindAllString(contentStr, -1)
			if len(matches) > 0 {
				// Generate snippets
				snippets = generateSnippets(re, contentStr, contextLines)
			}
		} else {
			// Simple text search
			searchContent := contentStr
			searchQuery := query

			if !caseSensitive {
				searchContent = strings.ToLower(searchContent)
				searchQuery = strings.ToLower(searchQuery)
			}

			// Find all matches
			lines := strings.Split(contentStr, "\n")
			for i, line := range lines {
				lineToSearch := line
				if !caseSensitive {
					lineToSearch = strings.ToLower(line)
				}

				if strings.Contains(lineToSearch, searchQuery) {
					matches = append(matches, strings.TrimSpace(line))

					// Generate context snippet
					if contextLines > 0 {
						start := i - contextLines
						if start < 0 {
							start = 0
						}
						end := i + contextLines + 1
						if end > len(lines) {
							end = len(lines)
						}
						snippet := strings.Join(lines[start:end], "\n")
						snippets = append(snippets, snippet)
					}
				}
			}
		}

		// Calculate score
		if len(matches) > 0 {
			// Base score for having matches
			score = 10.0

			// Bonus for more matches
			score += float64(len(matches)) * 2.0

			// Bonus for role match if query looks like a role
			if strings.EqualFold(f.Role, query) {
				score += 50.0
			}

			// Bonus for shorter files (more focused)
			if f.Tokens > 0 {
				score += 1000.0 / float64(f.Tokens)
			}

			results = append(results, SearchResult{
				File:     f.Path,
				Role:     f.Role,
				Tokens:   f.Tokens,
				Matches:  uniqueMatches(matches),
				Snippets: snippets,
				Score:    score,
			})
		}
	}

	return results, nil
}

// generateSnippets generates context snippets using regex
func generateSnippets(re *regexp.Regexp, content string, contextLines int) []string {
	lines := strings.Split(content, "\n")
	snippets := []string{}

	matches := re.FindAllStringIndex(content, -1)
	for _, match := range matches {
		// Count line number
		lineNum := strings.Count(content[:match[0]], "\n")

		// Generate snippet with context
		start := lineNum - contextLines
		if start < 0 {
			start = 0
		}
		end := lineNum + contextLines + 1
		if end > len(lines) {
			end = len(lines)
		}

		snippet := strings.Join(lines[start:end], "\n")
		snippets = append(snippets, snippet)
	}

	return snippets
}

// uniqueMatches removes duplicate matches
func uniqueMatches(matches []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, m := range matches {
		if !seen[m] {
			seen[m] = true
			result = append(result, m)
		}
	}

	return result
}

// displaySearchResults shows search results
func displaySearchResults(results []SearchResult, query string, showContext bool) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Search results for: %s (%d matches)\n\n", query, len(results))

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "File", "Role", "Tokens", "Score", "Matches")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", "----", "----", "------", "-----", "-------")

	for i, r := range results {
		shortFile := r.File
		if len(shortFile) > 20 {
			shortFile = shortFile[:17] + ".."
		}

		shortRole := r.Role
		if len(shortRole) > 12 {
			shortRole = shortRole[:9] + ".."
		}

		matchCount := len(r.Matches)
		matchStr := fmt.Sprintf("%d", matchCount)
		if matchCount > 9 {
			matchStr = fmt.Sprintf("%d+", matchCount)
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%.1f\t%s\n",
			shortFile, shortRole, r.Tokens, r.Score, matchStr)

		// Show snippets if available
		if showContext && len(r.Snippets) > 0 && i < 3 {
			fmt.Printf("\n  Snippets from %s:\n", filepath.Base(r.File))
			for j, snippet := range r.Snippets {
				if j >= 2 {
					fmt.Printf("    ... %d more snippets\n", len(r.Snippets)-j)
					break
				}
				// Highlight matching lines
				lines := strings.Split(snippet, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						fmt.Printf("    %s\n", line)
					}
				}
				fmt.Println()
			}
		}
	}

	w.Flush()

	return nil
}

// FindByRoleCmd finds files by role pattern
func FindByRoleCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	// Get role pattern
	pattern := c.Args().First()
	if pattern == "" {
		return fmt.Errorf("role pattern required. Usage: gitai find-role <pattern>")
	}

	// Get tracked files
	tracked, err := storage.ListAllTracked()
	if err != nil {
		return fmt.Errorf("failed to get tracked files: %w", err)
	}

	if len(tracked) == 0 {
		fmt.Println("No tracked files.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Files matching role pattern: %s\n\n", pattern)

	fmt.Fprintf(w, "%s\t%s\t%s\n", "File", "Role", "Tokens")
	fmt.Fprintf(w, "%s\t%s\t%s\n", "----", "----", "------")

	count := 0
	for _, f := range tracked {
		if strings.Contains(strings.ToLower(f.Role), strings.ToLower(pattern)) {
			fmt.Fprintf(w, "%s\t%s\t%d\n", f.Path, f.Role, f.Tokens)
			count++
		}
	}

	w.Flush()

	if count == 0 {
		fmt.Printf("No files found matching role pattern: %s\n", pattern)
	} else {
		fmt.Printf("\nFound %d file(s)\n", count)
	}

	return nil
}
