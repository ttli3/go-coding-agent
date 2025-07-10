package tools

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ttli3/go-coding-agent/internal/ui"
)

// ReadFileTool reads the contents of a file
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path parameter is required and must be a string")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

func (t *ReadFileTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"path": {
				Type:        "string",
				Description: "Path to the file to read",
			},
		},
		Required: []string{"path"},
	}
}

// WriteFileTool writes content to a file
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file, creating it if it doesn't exist"
}

func (t *WriteFileTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path parameter is required and must be a string")
	}

	content, ok := args["content"].(string)
	if !ok {
		return "", fmt.Errorf("content parameter is required and must be a string")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Check if file exists to show diff
	var diffOutput string
	if existingContent, err := os.ReadFile(path); err == nil {
		// File exists, show diff
		diffFormatter := ui.NewDiffFormatter()
		diffOutput = diffFormatter.FormatDiff(path, string(existingContent), content, 15) + "\n\n"
	} else {
		// New file
		diffFormatter := ui.NewDiffFormatter()
		diffOutput = diffFormatter.FormatDiff(path, "", content, 15) + "\n\n"
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("%sSuccessfully wrote %d bytes to %s", diffOutput, len(content), path), nil
}

func (t *WriteFileTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"path": {
				Type:        "string",
				Description: "Path to the file to write",
			},
			"content": {
				Type:        "string",
				Description: "Content to write to the file",
			},
		},
		Required: []string{"path", "content"},
	}
}

// ListDirectoryTool lists the contents of a directory
type ListDirectoryTool struct{}

func (t *ListDirectoryTool) Name() string {
	return "list_directory"
}

func (t *ListDirectoryTool) Description() string {
	return "List the contents of a directory"
}

func (t *ListDirectoryTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path parameter is required and must be a string")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Contents of %s:\n", path))

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileType := "file"
		if entry.IsDir() {
			fileType = "directory"
		}

		result.WriteString(fmt.Sprintf("  %s (%s, %d bytes)\n",
			entry.Name(), fileType, info.Size()))
	}

	return result.String(), nil
}

func (t *ListDirectoryTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"path": {
				Type:        "string",
				Description: "Path to the directory to list",
			},
		},
		Required: []string{"path"},
	}
}

// FindFilesTool finds files matching a pattern
type FindFilesTool struct{}

func (t *FindFilesTool) Name() string {
	return "find_files"
}

func (t *FindFilesTool) Description() string {
	return "Find files matching a pattern recursively in directories and subdirectories"
}

func (t *FindFilesTool) Execute(args map[string]interface{}) (string, error) {
	path, ok := args["path"].(string)
	if !ok {
		return "", fmt.Errorf("path parameter is required and must be a string")
	}

	pattern, ok := args["pattern"].(string)
	if !ok {
		return "", fmt.Errorf("pattern parameter is required and must be a string")
	}

	// Optional: include hidden files (default: false)
	includeHidden := false
	if val, exists := args["include_hidden"]; exists {
		if hidden, ok := val.(bool); ok {
			includeHidden = hidden
		}
	}

	// Optional: max depth (default: unlimited)
	maxDepth := -1
	if val, exists := args["max_depth"]; exists {
		if depth, ok := val.(float64); ok {
			maxDepth = int(depth)
		}
	}

	var matches []string
	baseDepth := strings.Count(path, string(filepath.Separator))

	err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Continue walking even if there's an error
		}

		// Check depth limit
		if maxDepth >= 0 {
			currentDepth := strings.Count(filePath, string(filepath.Separator)) - baseDepth
			if currentDepth > maxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip hidden directories and files unless explicitly included
		// But don't skip the root directory even if it's "."
		if !includeHidden && strings.HasPrefix(d.Name(), ".") && filePath != path {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			// Support both filename matching and full path matching
			filenameMatched, err := filepath.Match(pattern, d.Name())
			if err != nil {
				return nil
			}

			// Also try matching against the relative path
			relPath, _ := filepath.Rel(path, filePath)
			pathMatched, err := filepath.Match(pattern, relPath)
			if err != nil {
				pathMatched = false
			}

			// Check if filename contains the pattern (case-insensitive)
			containsPattern := strings.Contains(strings.ToLower(d.Name()), strings.ToLower(pattern))

			if filenameMatched || pathMatched || containsPattern {
				matches = append(matches, filePath)
			}
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(matches) == 0 {
		return fmt.Sprintf("No files matching pattern '%s' found in %s (searched recursively)", pattern, path), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Found %d files matching '%s' (searched recursively):\n", len(matches), pattern))
	for _, match := range matches {
		// Show relative path for cleaner output
		relPath, err := filepath.Rel(path, match)
		if err != nil {
			relPath = match
		}
		result.WriteString(fmt.Sprintf("  %s\n", relPath))
	}

	return result.String(), nil
}

func (t *FindFilesTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"path": {
				Type:        "string",
				Description: "Directory path to search in",
			},
			"pattern": {
				Type:        "string",
				Description: "File name pattern to match (supports wildcards, substring matching, and path matching)",
			},
			"include_hidden": {
				Type:        "boolean",
				Description: "Whether to include hidden files and directories (default: false)",
			},
			"max_depth": {
				Type:        "number",
				Description: "Maximum depth to search (default: unlimited)",
			},
		},
		Required: []string{"path", "pattern"},
	}
}
