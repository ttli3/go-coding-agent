package tools

import (
	"fmt"
	"os"

	"github.com/ttli3/go-coding-agent/internal/ui"
)

type DiffTool struct{}


func (t *DiffTool) Name() string {
	return "show_diff"
}

func (t *DiffTool) Description() string {
	return "Show a diff view between two versions of a file or content"
}

func (t *DiffTool) Execute(args map[string]interface{}) (string, error) {
	filename, ok := args["filename"].(string)
	if !ok {
		return "", fmt.Errorf("filename is required")
	}

	oldContent, hasOld := args["old_content"].(string)
	newContent, hasNew := args["new_content"].(string)
	if !hasOld {
		if content, err := os.ReadFile(filename); err == nil {
			oldContent = string(content)
		} else {
			oldContent = "" 
		}
	}

	if !hasNew {
		return "", fmt.Errorf("new_content is required")
	}

	// How many lines to show before we collapse stuff
	maxLines := 20 // Default
	if ml, ok := args["max_lines"].(float64); ok {
		maxLines = int(ml)
	}

	// make diff look good  
	formatter := ui.NewDiffFormatter()
	diff := formatter.FormatDiff(filename, oldContent, newContent, maxLines)

	return diff, nil
}

func (t *DiffTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"filename": {
				Type:        "string",
				Description: "The filename being modified",
			},
			"old_content": {
				Type:        "string",
				Description: "The original content (optional, will read from file if not provided)",
			},
			"new_content": {
				Type:        "string",
				Description: "The new content to compare against",
			},
			"max_lines": {
				Type:        "number",
				Description: "Maximum number of lines to show before collapsing (default: 20)",
			},
		},
		Required: []string{"filename", "new_content"},
	}
}
