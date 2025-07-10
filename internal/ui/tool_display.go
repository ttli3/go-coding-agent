package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// ToolExecutionDisplay manages the display of tool execution progress
type ToolExecutionDisplay struct {
	currentTool string
	startTime   time.Time
	totalTools  int
	completed   int
}

// NewToolExecutionDisplay creates a new tool execution display
func NewToolExecutionDisplay(totalTools int) *ToolExecutionDisplay {
	return &ToolExecutionDisplay{
		totalTools: totalTools,
		completed:  0,
	}
}

// StartTool displays the start of a tool execution
func (ted *ToolExecutionDisplay) StartTool(toolName string, args map[string]interface{}) {
	ted.currentTool = toolName
	ted.startTime = time.Now()
	
	// Show tool execution header
	color.New(color.FgHiBlack).Printf("\n┌─ Tool %d/%d: ", ted.completed+1, ted.totalTools)
	color.New(color.FgCyan, color.Bold).Printf("%s", toolName)
	color.New(color.FgHiBlack).Println(" ─────────────────────")
	
	// Show simplified arguments
	if len(args) > 0 {
		color.New(color.FgHiBlack).Print("│ ")
		ted.displaySimplifiedArgs(args)
	}
	
	color.New(color.FgHiBlack).Print("│ ")
	color.New(color.FgYellow).Print("[EXEC] Executing...")
}

// FinishTool displays the completion of a tool execution
func (ted *ToolExecutionDisplay) FinishTool(success bool, result string, err error) {
	duration := time.Since(ted.startTime)
	ted.completed++
	
	// Clear the "Executing..." line
	fmt.Print("\r")
	color.New(color.FgHiBlack).Print("│ ")
	
	if success {
		color.New(color.FgGreen).Printf("[OK] Completed")
	} else {
		color.New(color.FgRed).Printf("[FAIL] Failed")
	}
	
	color.New(color.FgHiBlack).Printf(" (%s)\n", formatDuration(duration))
	
	// Show result summary (truncated)
	if success && result != "" {
		summary := ted.summarizeResult(ted.currentTool, result)
		if summary != "" {
			color.New(color.FgHiBlack).Print("│ ")
			color.New(color.FgWhite).Println(summary)
		}
	} else if err != nil {
		color.New(color.FgHiBlack).Print("│ ")
		color.New(color.FgRed).Printf("Error: %s\n", ted.truncateString(err.Error(), 60))
	}
	
	color.New(color.FgHiBlack).Println("└─────────────────────────────────────────")
}

// displaySimplifiedArgs shows a simplified view of tool arguments
func (ted *ToolExecutionDisplay) displaySimplifiedArgs(args map[string]interface{}) {
	var parts []string
	
	for key, value := range args {
		switch key {
		case "path", "file_path", "filename":
			if str, ok := value.(string); ok {
				parts = append(parts, fmt.Sprintf("File: %s", ted.shortenPath(str)))
			}
		case "content":
			if str, ok := value.(string); ok {
				lines := strings.Count(str, "\n") + 1
				chars := len(str)
				parts = append(parts, fmt.Sprintf("Content: %d lines, %d chars", lines, chars))
			}
		case "pattern", "query", "search":
			if str, ok := value.(string); ok {
				parts = append(parts, fmt.Sprintf("Query: \"%s\"", ted.truncateString(str, 30)))
			}
		case "command":
			if str, ok := value.(string); ok {
				parts = append(parts, fmt.Sprintf("Command: %s", ted.truncateString(str, 40)))
			}
		case "directory", "dir":
			if str, ok := value.(string); ok {
				parts = append(parts, fmt.Sprintf("Dir: %s", ted.shortenPath(str)))
			}
		}
	}
	
	if len(parts) > 0 {
		color.New(color.FgHiBlack).Println(strings.Join(parts, " • "))
	}
}

// summarizeResult creates a brief summary of tool results
func (ted *ToolExecutionDisplay) summarizeResult(toolName, result string) string {
	switch toolName {
	case "read_file":
		lines := strings.Count(result, "\n") + 1
		return fmt.Sprintf("Read %d lines", lines)
	case "write_file":
		if strings.Contains(result, "Successfully wrote") {
			return "File written successfully"
		}
		return "File operation completed"
	case "list_directory":
		lines := strings.Count(result, "\n")
		return fmt.Sprintf("Found %d items", lines)
	case "find_files":
		lines := strings.Count(result, "\n")
		if lines > 0 {
			return fmt.Sprintf("Found %d files", lines)
		}
		return "No files found"
	case "run_command":
		lines := strings.Count(result, "\n")
		if lines > 5 {
			return fmt.Sprintf("Command output: %d lines", lines)
		}
		return ted.truncateString(result, 50)
	case "search_code", "grep_search":
		matches := strings.Count(result, "\n")
		return fmt.Sprintf("Found %d matches", matches)
	default:
		// Generic summary
		if len(result) > 100 {
			return fmt.Sprintf("Result: %d chars", len(result))
		}
		return ted.truncateString(result, 50)
	}
}

// Helper functions

func (ted *ToolExecutionDisplay) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (ted *ToolExecutionDisplay) shortenPath(path string) string {
	if len(path) <= 40 {
		return path
	}
	
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return ".../" + strings.Join(parts[len(parts)-2:], "/")
	}
	
	return ted.truncateString(path, 40)
}

// ShowToolSummary displays a summary after all tools complete
func (ted *ToolExecutionDisplay) ShowToolSummary() {
	if ted.totalTools > 1 {
		color.New(color.FgHiBlack).Printf("\n[SUMMARY] Completed %d tools\n", ted.totalTools)
	}
}
