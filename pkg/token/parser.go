package token

import (
	"regexp"
	"strings"
)

// SectionType represents different sections in a prompt file
type SectionType string

const (
	SectionRole        SectionType = "role"
	SectionTask        SectionType = "task"
	SectionConstraints SectionType = "constraints"
	SectionExamples    SectionType = "examples"
	SectionInput       SectionType = "input"
	SectionOutput      SectionType = "output"
	SectionContext     SectionType = "context"
	SectionOther       SectionType = "other"
)

// PromptSection represents a parsed section from a .prompt file
type PromptSection struct {
	Type    SectionType
	Content string
	Tokens  int
}

// ParsedPrompt represents a fully parsed .prompt file
type ParsedPrompt struct {
	Role        string
	Task        string
	Constraints string
	Examples    string
	Input       string
	Output      string
	Context     string
	Other       string
	TotalTokens int
	Sections    []PromptSection
}

// Parser handles parsing of .prompt files
type Parser struct {
	sectionPatterns map[SectionType]*regexp.Regexp
}

// NewParser creates a new prompt parser
func NewParser() *Parser {
	p := &Parser{
		sectionPatterns: make(map[SectionType]*regexp.Regexp),
	}

	// Section patterns: supports # section:, ## section:, ### section:, # Section:, etc.
	p.sectionPatterns[SectionRole] = regexp.MustCompile(`(?i)^#{1,6}\s*role:\s*$`)
	p.sectionPatterns[SectionTask] = regexp.MustCompile(`(?i)^#{1,6}\s*(task|objective|goal|purpose):\s*$`)
	p.sectionPatterns[SectionConstraints] = regexp.MustCompile(`(?i)^#{1,6}\s*(constraints|requirements|rules|limitations):\s*$`)
	p.sectionPatterns[SectionExamples] = regexp.MustCompile(`(?i)^#{1,6}\s*(examples|example|demo|illustration):\s*$`)
	p.sectionPatterns[SectionInput] = regexp.MustCompile(`(?i)^#{1,6}\s*input:\s*$`)
	p.sectionPatterns[SectionOutput] = regexp.MustCompile(`(?i)^#{1,6}\s*output:\s*$`)
	p.sectionPatterns[SectionContext] = regexp.MustCompile(`(?i)^#{1,6}\s*(context|background):\s*$`)

	return p
}

// Parse parses a .prompt file content into structured sections
func (p *Parser) Parse(content string) *ParsedPrompt {
	lines := strings.Split(content, "\n")
	result := &ParsedPrompt{
		Sections: make([]PromptSection, 0),
	}

	currentSection := SectionOther
	currentContent := strings.Builder{}
	inSection := false

	// First pass: extract role from single-line format
	roleLine := regexp.MustCompile(`(?i)^#\s*role:\s*(.+)$`)
	for _, line := range lines {
		if matches := roleLine.FindStringSubmatch(line); matches != nil {
			result.Role = strings.TrimSpace(matches[1])
			break
		}
	}

	// Second pass: parse multi-line sections
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if this is a section header
		sectionType := p.identifySection(trimmed)
		if sectionType != SectionOther {
			// Save previous section
			if inSection && currentContent.Len() > 0 {
				content := strings.TrimSpace(currentContent.String())
				if content != "" {
					section := PromptSection{
						Type:    currentSection,
						Content: content,
						Tokens:  Estimate(content),
					}
					result.Sections = append(result.Sections, section)

					// Store in specific field
					p.assignSection(result, currentSection, content)
				}
			}

			// Start new section
			currentSection = sectionType
			currentContent.Reset()
			inSection = true
			continue
		}

		// Skip empty lines at the start of a section
		if inSection && currentContent.Len() == 0 && trimmed == "" {
			continue
		}

		// Add line to current section
		if inSection {
			if currentContent.Len() > 0 {
				currentContent.WriteString("\n")
			}
			currentContent.WriteString(line)
		}
	}

	// Don't forget the last section
	if inSection && currentContent.Len() > 0 {
		content := strings.TrimSpace(currentContent.String())
		if content != "" {
			section := PromptSection{
				Type:    currentSection,
				Content: content,
				Tokens:  Estimate(content),
			}
			result.Sections = append(result.Sections, section)
			p.assignSection(result, currentSection, content)
		}
	}

	// If no sections were found, treat entire content as "task" or "other"
	if len(result.Sections) == 0 {
		trimmedContent := strings.TrimSpace(content)
		result.Sections = append(result.Sections, PromptSection{
			Type:    SectionTask,
			Content: trimmedContent,
			Tokens:  Estimate(trimmedContent),
		})
		result.Task = trimmedContent
	}

	// Calculate total tokens
	result.TotalTokens = 0
	for _, section := range result.Sections {
		result.TotalTokens += section.Tokens
	}

	return result
}

// identifySection identifies which section type a line represents
func (p *Parser) identifySection(line string) SectionType {
	for sectionType, pattern := range p.sectionPatterns {
		if pattern.MatchString(line) {
			return sectionType
		}
	}
	return SectionOther
}

// assignSection assigns content to the appropriate field
func (p *Parser) assignSection(result *ParsedPrompt, sectionType SectionType, content string) {
	switch sectionType {
	case SectionRole:
		if result.Role == "" {
			result.Role = content
		}
	case SectionTask:
		result.Task = content
	case SectionConstraints:
		result.Constraints = content
	case SectionExamples:
		result.Examples = content
	case SectionInput:
		result.Input = content
	case SectionOutput:
		result.Output = content
	case SectionContext:
		result.Context = content
	default:
		if result.Other == "" {
			result.Other = content
		} else {
			result.Other += "\n" + content
		}
	}
}

// GetSectionByType retrieves a specific section
func (p *ParsedPrompt) GetSectionByType(sectionType SectionType) *PromptSection {
	for i := range p.Sections {
		if p.Sections[i].Type == sectionType {
			return &p.Sections[i]
		}
	}
	return nil
}

// GetSummary returns a summary of the parsed prompt
func (p *ParsedPrompt) GetSummary() string {
	var parts []string
	if p.Role != "" {
		parts = append(parts, "Role: "+p.Role)
	}
	if p.Task != "" {
		task := p.Task
		if len(task) > 50 {
			task = task[:50] + "..."
		}
		parts = append(parts, "Task: "+task)
	}
	parts = append(parts, "Tokens: "+string(rune(p.TotalTokens+'0')))
	return strings.Join(parts, " | ")
}

// Default parser instance
var defaultParser = NewParser()

// Parse is a convenience function using the default parser
func Parse(content string) *ParsedPrompt {
	return defaultParser.Parse(content)
}
