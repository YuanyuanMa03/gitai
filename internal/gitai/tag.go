package gitai

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli/v2"
)

// TagsFile holds the tags metadata filename
const TagsFile = "tags.json"

// TagsData stores all tag mappings
type TagsData struct {
	Version string            `json:"version"`
	Tags    map[string][]string `json:"tags"` // path -> tags
}

// LoadTags loads tags from .gitai/tags.json
func (s *Storage) LoadTags() (*TagsData, error) {
	tagsPath := filepath.Join(s.gitaiDir, TagsFile)

	// If file doesn't exist, return empty tags
	if _, err := os.Stat(tagsPath); os.IsNotExist(err) {
		return &TagsData{
			Version: "1.0.0",
			Tags:    make(map[string][]string),
		}, nil
	}

	data, err := os.ReadFile(tagsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read tags file: %w", err)
	}

	var tags TagsData
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("failed to parse tags: %w", err)
	}

	if tags.Tags == nil {
		tags.Tags = make(map[string][]string)
	}

	return &tags, nil
}

// SaveTags saves tags to .gitai/tags.json
func (s *Storage) SaveTags(tags *TagsData) error {
	tagsPath := filepath.Join(s.gitaiDir, TagsFile)

	data, err := json.MarshalIndent(tags, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	if err := os.WriteFile(tagsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tags file: %w", err)
	}

	return nil
}

// TagCmd handles tag commands
func TagCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Create storage
	storage := NewStorage(root)
	if !storage.IsInitialized() {
		return fmt.Errorf(".gitai/ not initialized. Run 'gitai init' first")
	}

	args := c.Args()

	if args.Len() == 0 {
		// List all tags
		return ListAllTagsCmd(storage)
	}

	action := args.Get(0)
	switch action {
	case "add":
		return AddTagsCmd(c, storage)
	case "remove", "rm":
		return RemoveTagsCmd(c, storage)
	case "list", "ls":
		return ListTagsCmd(c, storage)
	case "search":
		return SearchByTagCmd(c, storage)
	case "set":
		return SetTagsCmd(c, storage)
	default:
		return fmt.Errorf("unknown tag action: %s\nUsage: gitai tag [add|remove|list|search|set]", action)
	}
}

// AddTagsCmd adds tags to a file
func AddTagsCmd(c *cli.Context, storage *Storage) error {
	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai tag add <file> <tag1> [tag2 ...]")
	}

	filePath := args.Get(1)
	tags := args.Slice()[2:]

	if len(tags) == 0 {
		return fmt.Errorf("at least one tag required")
	}

	// Load existing tags
	tagsData, err := storage.LoadTags()
	if err != nil {
		return err
	}

	// Get existing tags for file
	existingTags := tagsData.Tags[filePath]
	tagSet := make(map[string]bool)
	for _, t := range existingTags {
		tagSet[t] = true
	}

	// Add new tags
	added := 0
	for _, tag := range tags {
		if !tagSet[tag] {
			existingTags = append(existingTags, tag)
			tagSet[tag] = true
			added++
		}
	}

	// Sort and deduplicate
	sort.Strings(existingTags)
	uniqueTags := make([]string, 0, len(existingTags))
	last := ""
	for _, t := range existingTags {
		if t != last {
			uniqueTags = append(uniqueTags, t)
			last = t
		}
	}

	tagsData.Tags[filePath] = uniqueTags

	// Save
	if err := storage.SaveTags(tagsData); err != nil {
		return err
	}

	fmt.Printf("Added %d tag(s) to %s\n", added, filePath)
	if len(uniqueTags) > 0 {
		fmt.Printf("Tags: %s\n", strings.Join(uniqueTags, ", "))
	}

	return nil
}

// RemoveTagsCmd removes tags from a file
func RemoveTagsCmd(c *cli.Context, storage *Storage) error {
	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai tag remove <file> <tag1> [tag2 ...]")
	}

	filePath := args.Get(1)
	tags := args.Slice()[2:]

	if len(tags) == 0 {
		return fmt.Errorf("at least one tag required")
	}

	// Load existing tags
	tagsData, err := storage.LoadTags()
	if err != nil {
		return err
	}

	// Get existing tags
	existingTags := tagsData.Tags[filePath]
	if len(existingTags) == 0 {
		fmt.Printf("File has no tags: %s\n", filePath)
		return nil
	}

	// Remove specified tags
	newTags := make([]string, 0, len(existingTags))
	removed := 0
	for _, t := range existingTags {
		found := false
		for _, tagToRemove := range tags {
			if t == tagToRemove {
				found = true
				removed++
				break
			}
		}
		if !found {
			newTags = append(newTags, t)
		}
	}

	if len(newTags) == 0 {
		delete(tagsData.Tags, filePath)
	} else {
		tagsData.Tags[filePath] = newTags
	}

	// Save
	if err := storage.SaveTags(tagsData); err != nil {
		return err
	}

	fmt.Printf("Removed %d tag(s) from %s\n", removed, filePath)
	if len(newTags) > 0 {
		fmt.Printf("Remaining tags: %s\n", strings.Join(newTags, ", "))
	} else {
		fmt.Println("No tags remaining")
	}

	return nil
}

// SetTagsCmd sets tags for a file (replaces existing)
func SetTagsCmd(c *cli.Context, storage *Storage) error {
	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai tag set <file> <tag1> [tag2 ...]")
	}

	filePath := args.Get(1)
	tags := args.Slice()[2:]

	if len(tags) == 0 {
		return fmt.Errorf("at least one tag required")
	}

	// Load existing tags
	tagsData, err := storage.LoadTags()
	if err != nil {
		return err
	}

	// Sort tags
	sort.Strings(tags)

	// Set tags
	tagsData.Tags[filePath] = tags

	// Save
	if err := storage.SaveTags(tagsData); err != nil {
		return err
	}

	fmt.Printf("Set %d tag(s) for %s\n", len(tags), filePath)
	fmt.Printf("Tags: %s\n", strings.Join(tags, ", "))

	return nil
}

// ListTagsCmd lists tags for a file
func ListTagsCmd(c *cli.Context, storage *Storage) error {
	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai tag list <file>")
	}

	filePath := args.Get(1)

	// Load tags
	tagsData, err := storage.LoadTags()
	if err != nil {
		return err
	}

	tags := tagsData.Tags[filePath]
	if len(tags) == 0 {
		fmt.Printf("No tags found for: %s\n", filePath)
		return nil
	}

	fmt.Printf("Tags for %s:\n", filePath)
	for _, tag := range tags {
		fmt.Printf("  - %s\n", tag)
	}

	return nil
}

// ListAllTagsCmd lists all tags across all files
func ListAllTagsCmd(storage *Storage) error {
	// Load tags
	tagsData, err := storage.LoadTags()
	if err != nil {
		return err
	}

	if len(tagsData.Tags) == 0 {
		fmt.Println("No tags found. Add tags with 'gitai tag add <file> <tag>'")
		return nil
	}

	// Build tag -> files mapping
	tagToFiles := make(map[string][]string)
	totalTags := 0

	for file, tags := range tagsData.Tags {
		for _, tag := range tags {
			tagToFiles[tag] = append(tagToFiles[tag], file)
			totalTags++
		}
	}

	// Sort tags
	sortedTags := make([]string, 0, len(tagToFiles))
	for tag := range tagToFiles {
		sortedTags = append(sortedTags, tag)
	}
	sort.Strings(sortedTags)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("All Tags (%d tags, %d files)\n\n", len(sortedTags), len(tagsData.Tags))

	fmt.Fprintf(w, "%s\t%s\t%s\n", "Tag", "Files", "Filenames")
	fmt.Fprintf(w, "%s\t%s\t%s\n", "---", "-----", "---------")

	for _, tag := range sortedTags {
		files := tagToFiles[tag]
		shortFiles := strings.Join(files, ", ")
		if len(shortFiles) > 50 {
			shortFiles = shortFiles[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%d\t%s\n", tag, len(files), shortFiles)
	}

	w.Flush()

	return nil
}

// SearchByTagCmd searches files by tag
func SearchByTagCmd(c *cli.Context, storage *Storage) error {
	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai tag search <tag>")
	}

	searchTag := args.Get(1)

	// Load tags
	tagsData, err := storage.LoadTags()
	if err != nil {
		return err
	}

	// Find matching files
	matches := []string{}
	for file, tags := range tagsData.Tags {
		for _, tag := range tags {
			if strings.EqualFold(tag, searchTag) {
				matches = append(matches, file)
				break
			}
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No files found with tag: %s\n", searchTag)
		return nil
	}

	// Get tracked files for more info
	tracked, _ := storage.ListAllTracked()
	trackedMap := make(map[string]TrackedFile)
	for _, f := range tracked {
		trackedMap[f.Path] = f
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Files with tag '%s' (%d found)\n\n", searchTag, len(matches))

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "File", "Role", "Tokens", "Tags")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "----", "----", "------", "----")

	for _, match := range matches {
		tf := trackedMap[match]
		role := tf.Role
		tokens := tf.Tokens

		// Get all tags for this file
		fileTags := tagsData.Tags[match]
		tagsStr := strings.Join(fileTags, ", ")

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", match, role, tokens, tagsStr)
	}

	w.Flush()

	return nil
}

// GetTagsForFile gets tags for a specific file
func (s *Storage) GetTagsForFile(path string) ([]string, error) {
	tagsData, err := s.LoadTags()
	if err != nil {
		return nil, err
	}

	return tagsData.Tags[path], nil
}
