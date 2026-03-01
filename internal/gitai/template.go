package gitai

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/mayuanyuan/gitai/pkg/token"
	"github.com/urfave/cli/v2"
)

// Template represents a prompt template
type Template struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	Role        string   `json:"role"`
	Content     string   `json:"content"`
	Variables   []string `json:"variables,omitempty"` // Placeholder variables like {{name}}
	Tags        []string `json:"tags,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

// TemplatesDir holds the templates directory name
const TemplatesDir = "templates"

// TemplateCmd handles template commands
func TemplateCmd(c *cli.Context) error {
	args := c.Args()

	if args.Len() == 0 {
		// List templates
		return ListTemplatesCmd(c)
	}

	action := args.Get(0)
	switch action {
	case "create":
		return CreateTemplateCmd(c)
	case "use":
		return UseTemplateCmd(c)
	case "delete":
		return DeleteTemplateCmd(c)
	case "edit":
		return EditTemplateCmd(c)
	case "show":
		return ShowTemplateCmd(c)
	case "list":
		return ListTemplatesCmd(c)
	default:
		return fmt.Errorf("unknown template action: %s\nUsage: gitai template [create|use|delete|edit|show|list]", action)
	}
}

// CreateTemplateCmd creates a new template
func CreateTemplateCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Parse options
	name := c.String("name")
	category := c.String("category")
	from := c.String("from")
	description := c.String("description")

	if name == "" {
		return fmt.Errorf("template name required. Use --name <name>")
	}

	// Get templates directory
	templatesDir := filepath.Join(root, GitaiDirName, TemplatesDir)
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return fmt.Errorf("failed to create templates directory: %w", err)
	}

	templatePath := filepath.Join(templatesDir, name+".prompt")

	// Check if template already exists
	if _, err := os.Stat(templatePath); err == nil {
		return fmt.Errorf("template already exists: %s", name)
	}

	var content string
	var variables []string

	// Create from existing file
	if from != "" {
		absPath, err := GetAbsolutePath(from)
		if err != nil {
			return fmt.Errorf("invalid file path: %w", err)
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		content = string(data)

		// Find potential variables
		variables = findVariables(content)
	} else {
		// Create empty template
		content = fmt.Sprintf(`# role: %s

## Task:
Describe what the AI should do here.

## Constraints:
List any rules or limitations.

## Examples:
Provide examples of expected input/output.

## Input:
Describe the input format.

## Output:
Describe the expected output format.
`, name)
	}

	// Create template file content with metadata
	// Template metadata will be stored in .gitai/templates/

	// Write template file
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	fmt.Printf("Created template: %s\n", name)
	if description != "" {
		fmt.Printf("  Description: %s\n", description)
	}
	if category != "" {
		fmt.Printf("  Category: %s\n", category)
	}
	if len(variables) > 0 {
		fmt.Printf("  Variables: %s\n", strings.Join(variables, ", "))
	}

	return nil
}

// UseTemplateCmd uses a template to create a new prompt file
func UseTemplateCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai template use <template-name> <output-file>")
	}

	templateName := args.Get(1)
	outputFile := c.String("output")

	// Get templates directory
	templatesDir := filepath.Join(root, GitaiDirName, TemplatesDir)
	templatePath := filepath.Join(templatesDir, templateName+".prompt")

	// Read template
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Replace variables if provided
	variables := c.StringSlice("var")
	templateContent := string(content)

	for _, v := range variables {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			placeholder := "{{" + parts[0] + "}}"
			templateContent = strings.ReplaceAll(templateContent, placeholder, parts[1])
		}
	}

	// Write output file
	absOutput, err := GetAbsolutePath(outputFile)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	if err := os.WriteFile(absOutput, []byte(templateContent), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Created file from template '%s': %s\n", templateName, outputFile)

	return nil
}

// DeleteTemplateCmd deletes a template
func DeleteTemplateCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai template delete <template-name>")
	}

	templateName := args.Get(1)

	// Get templates directory
	templatesDir := filepath.Join(root, GitaiDirName, TemplatesDir)
	templatePath := filepath.Join(templatesDir, templateName+".prompt")

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Delete template
	if err := os.Remove(templatePath); err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	fmt.Printf("Deleted template: %s\n", templateName)

	return nil
}

// EditTemplateCmd edits a template
func EditTemplateCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai template edit <template-name>")
	}

	templateName := args.Get(1)

	// Get templates directory
	templatesDir := filepath.Join(root, GitaiDirName, TemplatesDir)
	templatePath := filepath.Join(templatesDir, templateName+".prompt")

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template not found: %s", templateName)
	}

	fmt.Printf("Template location: %s\n", templatePath)
	fmt.Println("Edit the file directly, then run 'gitai template show' to verify.")

	return nil
}

// ShowTemplateCmd shows template details
func ShowTemplateCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	args := c.Args()
	if args.Len() < 2 {
		return fmt.Errorf("usage: gitai template show <template-name>")
	}

	templateName := args.Get(1)

	// Get templates directory
	templatesDir := filepath.Join(root, GitaiDirName, TemplatesDir)
	templatePath := filepath.Join(templatesDir, templateName+".prompt")

	// Read template
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Parse template
	parsed := token.Parse(string(content))

	fmt.Printf("Template: %s\n", templateName)
	fmt.Printf("Role: %s\n", parsed.Role)
	fmt.Printf("Tokens: %d\n", EstimateTokensQuick(string(content)))
	fmt.Println()

	// Show content
	fmt.Println("--- Content ---")
	fmt.Println(string(content))

	return nil
}

// ListTemplatesCmd lists all templates
func ListTemplatesCmd(c *cli.Context) error {
	// Get git root
	_, root := MustGetGitRepo()

	// Parse options
	category := c.String("category")

	// Get templates directory
	templatesDir := filepath.Join(root, GitaiDirName, TemplatesDir)

	// Check if directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		fmt.Println("No templates found. Create one with 'gitai template create'")
		return nil
	}

	// Find all template files
	matches, err := filepath.Glob(filepath.Join(templatesDir, "*.prompt"))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		fmt.Println("No templates found. Create one with 'gitai template create'")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Printf("Templates (%d)\n\n", len(matches))

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "Name", "Role", "Tokens", "Category")
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", "----", "----", "------", "--------")

	for _, match := range matches {
		name := strings.TrimSuffix(filepath.Base(match), ".prompt")

		// Read and parse template
		content, err := os.ReadFile(match)
		if err != nil {
			continue
		}

		parsed := token.Parse(string(content))
		tokens := EstimateTokensQuick(string(content))

		// Filter by category
		if category != "" && parsed.Category != category {
			continue
		}

		shortName := name
		if len(shortName) > 20 {
			shortName = shortName[:17] + ".."
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\n",
			shortName,
			parsed.Role,
			tokens,
			parsed.Category,
		)
	}

	w.Flush()

	return nil
}

// findVariables finds placeholder variables in template content
func findVariables(content string) []string {
	// Find {{variable}} patterns
	start := strings.Index(content, "{{")
	vars := make(map[string]bool)

	for start != -1 {
		end := strings.Index(content[start:], "}}")
		if end == -1 {
			break
		}

		varName := strings.TrimSpace(content[start+2 : start+end])
		if varName != "" {
			vars[varName] = true
		}

		content = content[start+end+2:]
		start = strings.Index(content, "{{")
	}

	result := make([]string, 0, len(vars))
	for v := range vars {
		result = append(result, v)
	}

	return result
}
