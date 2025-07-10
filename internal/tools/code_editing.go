package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ttli3/go-coding-agent/internal/ui"
)

// EditFileTool edits a file by replacing content at specific lines
type EditFileTool struct{}

func (t *EditFileTool) Name() string {
	return "edit_file"
}

func (t *EditFileTool) Description() string {
	return "Edit a file by replacing content at specific line numbers"
}

func (t *EditFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path parameter is required and must be a string")
	}

	startLine, ok := args["start_line"].(float64)
	if !ok {
		return "", fmt.Errorf("start_line parameter is required and must be a number")
	}

	endLine, ok := args["end_line"].(float64)
	if !ok {
		return "", fmt.Errorf("end_line parameter is required and must be a number")
	}

	newContent, ok := args["new_content"].(string)
	if !ok {
		return "", fmt.Errorf("new_content parameter is required and must be a string")
	}

	// Read the original file content
	originalContent, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(originalContent), "\n")

	// Validate line numbers (convert to 0-based indexing)
	start := int(startLine) - 1
	end := int(endLine) - 1

	if start < 0 || start >= len(lines) {
		return "", fmt.Errorf("start_line %d is out of range (file has %d lines)", int(startLine), len(lines))
	}

	if end < 0 || end >= len(lines) {
		return "", fmt.Errorf("end_line %d is out of range (file has %d lines)", int(endLine), len(lines))
	}

	if start > end {
		return "", fmt.Errorf("start_line cannot be greater than end_line")
	}

	// Create new content
	newLines := make([]string, 0, len(lines))
	newLines = append(newLines, lines[:start]...)

	// Split new content by lines and add them
	contentLines := strings.Split(newContent, "\n")
	newLines = append(newLines, contentLines...)

	// Add remaining lines after the edited section
	if end+1 < len(lines) {
		newLines = append(newLines, lines[end+1:]...)
	}

	newFileContent := strings.Join(newLines, "\n")

	// Show diff before making changes
	diffFormatter := ui.NewDiffFormatter()
	diff := diffFormatter.FormatDiff(path, string(originalContent), newFileContent, 15)

	// Write back to file
	if err := os.WriteFile(path, []byte(newFileContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("%s\n\nSuccessfully applied changes to %s (lines %d-%d)", diff, path, int(startLine), int(endLine)), nil
}

func (t *EditFileTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"path": {
				Type:        "string",
				Description: "Path to the file to edit",
			},
			"start_line": {
				Type:        "number",
				Description: "Starting line number (1-based)",
			},
			"end_line": {
				Type:        "number",
				Description: "Ending line number (1-based, inclusive)",
			},
			"new_content": {
				Type:        "string",
				Description: "New content to replace the specified lines",
			},
		},
		Required: []string{"path", "start_line", "end_line", "new_content"},
	}
}

// SearchCodeTool searches for code patterns in files
type SearchCodeTool struct{}

func (t *SearchCodeTool) Name() string {
	return "search_code"
}

func (t *SearchCodeTool) Description() string {
	return "Search for code patterns in a file"
}

func (t *SearchCodeTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path parameter is required and must be a string")
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return "", fmt.Errorf("pattern parameter is required and must be a string")
	}

	useRegex := false
	if val, exists := args["regex"]; exists {
		if regexVal, ok := val.(bool); ok {
			useRegex = regexVal
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var matches []string

	for i, line := range lines {
		var matched bool
		if useRegex {
			matched, err = regexp.MatchString(pattern, line)
			if err != nil {
				return "", fmt.Errorf("invalid regex pattern: %w", err)
			}
		} else {
			matched = strings.Contains(line, pattern)
		}

		if matched {
			matches = append(matches, fmt.Sprintf("Line %d: %s", i+1, line))
		}
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No matches found for pattern '%s' in %s", pattern, path), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d matches for '%s' in %s:\n", len(matches), pattern, path))
	for _, match := range matches {
		result.WriteString(fmt.Sprintf("  %s\n", match))
	}

	return result.String(), nil
}

func (t *SearchCodeTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"path": {
				Type:        "string",
				Description: "Path to the file to search",
			},
			"pattern": {
				Type:        "string",
				Description: "Pattern to search for",
			},
			"regex": {
				Type:        "boolean",
				Description: "Whether to treat pattern as a regular expression (default: false)",
			},
		},
		Required: []string{"path", "pattern"},
	}
}

// ReplaceContentTool replaces all occurrences of a pattern in a file
type ReplaceContentTool struct{}

func (t *ReplaceContentTool) Name() string {
	return "replace_content"
}

func (t *ReplaceContentTool) Description() string {
	return "Replace all occurrences of a pattern with new content in a file"
}

func (t *ReplaceContentTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path parameter is required and must be a string")
	}

	oldPattern, ok := args["old_pattern"].(string)
	if !ok {
		return "", fmt.Errorf("old_pattern parameter is required and must be a string")
	}

	newContent, ok := args["new_content"].(string)
	if !ok {
		return "", fmt.Errorf("new_content parameter is required and must be a string")
	}

	useRegex := false
	if val, exists := args["regex"]; exists {
		if regexVal, ok := val.(bool); ok {
			useRegex = regexVal
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)
	var modifiedContent string
	var replacements int

	if useRegex {
		re, err := regexp.Compile(oldPattern)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %w", err)
		}

		matches := re.FindAllString(originalContent, -1)
		replacements = len(matches)
		modifiedContent = re.ReplaceAllString(originalContent, newContent)
	} else {
		replacements = strings.Count(originalContent, oldPattern)
		modifiedContent = strings.ReplaceAll(originalContent, oldPattern, newContent)
	}

	if replacements == 0 {
		return fmt.Sprintf("No occurrences of pattern '%s' found in %s", oldPattern, path), nil
	}

	// Show diff before making changes
	diffFormatter := ui.NewDiffFormatter()
	diff := diffFormatter.FormatDiff(path, originalContent, modifiedContent, 15)

	if err := os.WriteFile(path, []byte(modifiedContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("%s\n\nSuccessfully replaced %d occurrences of '%s' in %s", diff, replacements, oldPattern, path), nil
}

func (t *ReplaceContentTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"path": {
				Type:        "string",
				Description: "Path to the file to modify",
			},
			"old_pattern": {
				Type:        "string",
				Description: "Pattern to search for and replace",
			},
			"new_content": {
				Type:        "string",
				Description: "Content to replace the pattern with",
			},
			"regex": {
				Type:        "boolean",
				Description: "Whether to treat old_pattern as a regular expression (default: false)",
			},
		},
		Required: []string{"path", "old_pattern", "new_content"},
	}
}

// GrepSearchTool searches for patterns across multiple files
type GrepSearchTool struct{}

func (t *GrepSearchTool) Name() string {
	return "grep_search"
}

func (t *GrepSearchTool) Description() string {
	return "Search for patterns across multiple files in a directory"
}

func (t *GrepSearchTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path parameter is required and must be a string")
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return "", fmt.Errorf("pattern parameter is required and must be a string")
	}

	filePattern := "*"
	if val, exists := args["file_pattern"]; exists {
		if fpVal, ok := val.(string); ok {
			filePattern = fpVal
		}
	}

	useRegex := false
	if val, exists := args["regex"]; exists {
		if regexVal, ok := val.(bool); ok {
			useRegex = regexVal
		}
	}

	var matches []string
	var regex *regexp.Regexp
	var err error

	if useRegex {
		regex, err = regexp.Compile(pattern)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches the file pattern
		matched, err := filepath.Match(filePattern, info.Name())
		if err != nil || !matched {
			return nil
		}

		// Skip binary files (simple heuristic)
		if strings.Contains(info.Name(), ".exe") || strings.Contains(info.Name(), ".bin") {
			return nil
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil // Continue with other files
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			var lineMatched bool
			if useRegex {
				lineMatched = regex.MatchString(line)
			} else {
				lineMatched = strings.Contains(line, pattern)
			}

			if lineMatched {
				matches = append(matches, fmt.Sprintf("%s:%d: %s", filePath, i+1, strings.TrimSpace(line)))
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to search files: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No matches found for pattern '%s' in %s", pattern, path), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d matches for '%s':\n", len(matches), pattern))
	for _, match := range matches {
		result.WriteString(fmt.Sprintf("  %s\n", match))
	}

	return result.String(), nil
}

func (t *GrepSearchTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"path": {
				Type:        "string",
				Description: "Directory path to search in",
			},
			"pattern": {
				Type:        "string",
				Description: "Pattern to search for",
			},
			"file_pattern": {
				Type:        "string",
				Description: "File name pattern to limit search (default: *)",
			},
			"regex": {
				Type:        "boolean",
				Description: "Whether to treat pattern as a regular expression (default: false)",
			},
		},
		Required: []string{"path", "pattern"},
	}
}
