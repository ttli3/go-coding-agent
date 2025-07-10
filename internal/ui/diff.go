package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// DiffLine represents a single line in a diff
type DiffLine struct {
	Type    DiffLineType
	Content string
	LineNum int
}

// DiffLineType represents the type of diff line
type DiffLineType int

const (
	DiffLineContext DiffLineType = iota
	DiffLineAdded
	DiffLineRemoved
)

// DiffFormatter handles formatting of code diffs
type DiffFormatter struct {
	addedColor   *color.Color
	removedColor *color.Color
	contextColor *color.Color
	headerColor  *color.Color
	lineNumColor *color.Color
}

// NewDiffFormatter creates a new diff formatter with default colors
func NewDiffFormatter() *DiffFormatter {
	return &DiffFormatter{
		addedColor:   color.New(color.FgGreen, color.BgHiBlack),
		removedColor: color.New(color.FgRed, color.BgHiBlack),
		contextColor: color.New(color.FgHiBlack),
		headerColor:  color.New(color.FgCyan, color.Bold),
		lineNumColor: color.New(color.FgHiBlack),
	}
}

// FormatDiff formats a diff with proper colors and optional collapsing
func (df *DiffFormatter) FormatDiff(filename string, oldContent, newContent string, maxLines int) string {
	lines := df.generateDiff(oldContent, newContent)
	
	if len(lines) == 0 {
		return df.headerColor.Sprintf("No changes detected in %s", filename)
	}

	var result strings.Builder
	
	// Header
	result.WriteString(df.headerColor.Sprintf("Changes in %s\n", filename))
	result.WriteString(df.contextColor.Sprintf("─%s\n", strings.Repeat("─", 50)))
	
	// Check if we need to collapse
	shouldCollapse := maxLines > 0 && len(lines) > maxLines
	
	if shouldCollapse {
		// Show first few lines
		showLines := maxLines / 2
		for i := 0; i < showLines && i < len(lines); i++ {
			result.WriteString(df.formatLine(lines[i]))
		}
		
		// Collapse indicator
		hiddenCount := len(lines) - maxLines
		if hiddenCount > 0 {
			result.WriteString(df.contextColor.Sprintf("┊ ... %d more lines (use --full-diff to see all) ...\n", hiddenCount))
		}
		
		// Show last few lines
		startIdx := len(lines) - (maxLines - showLines)
		if startIdx < showLines {
			startIdx = showLines
		}
		for i := startIdx; i < len(lines); i++ {
			result.WriteString(df.formatLine(lines[i]))
		}
	} else {
		// Show all lines
		for _, line := range lines {
			result.WriteString(df.formatLine(line))
		}
	}
	
	// Footer with stats
	added, removed := df.countChanges(lines)
	result.WriteString(df.contextColor.Sprintf("─%s\n", strings.Repeat("─", 50)))
	result.WriteString(df.addedColor.Sprintf("+%d ", added))
	result.WriteString(df.removedColor.Sprintf("-%d ", removed))
	result.WriteString(df.contextColor.Sprintf("lines changed\n"))
	
	return result.String()
}

// formatLine formats a single diff line with appropriate colors
func (df *DiffFormatter) formatLine(line DiffLine) string {
	var prefix, content string
	var colorFunc func(format string, a ...interface{}) string
	
	switch line.Type {
	case DiffLineAdded:
		prefix = "+"
		colorFunc = df.addedColor.Sprintf
	case DiffLineRemoved:
		prefix = "-"
		colorFunc = df.removedColor.Sprintf
	case DiffLineContext:
		prefix = " "
		colorFunc = df.contextColor.Sprintf
	}
	
	// Format line number if available
	lineNumStr := ""
	if line.LineNum > 0 {
		lineNumStr = df.lineNumColor.Sprintf("%4d ", line.LineNum)
	}
	
	content = strings.TrimRight(line.Content, "\n\r")
	if content == "" {
		content = " " // Show empty lines
	}
	
	return fmt.Sprintf("%s%s%s %s\n", lineNumStr, colorFunc(prefix), colorFunc(" "), colorFunc(content))
}

// generateDiff creates a simple diff between old and new content
func (df *DiffFormatter) generateDiff(oldContent, newContent string) []DiffLine {
	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")
	
	var diff []DiffLine
	
	// Simple line-by-line diff (could be improved with proper diff algorithm)
	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}
	
	oldIdx, newIdx := 0, 0
	
	for oldIdx < len(oldLines) || newIdx < len(newLines) {
		if oldIdx >= len(oldLines) {
			// Only new lines remaining
			diff = append(diff, DiffLine{
				Type:    DiffLineAdded,
				Content: newLines[newIdx],
				LineNum: newIdx + 1,
			})
			newIdx++
		} else if newIdx >= len(newLines) {
			// Only old lines remaining
			diff = append(diff, DiffLine{
				Type:    DiffLineRemoved,
				Content: oldLines[oldIdx],
				LineNum: oldIdx + 1,
			})
			oldIdx++
		} else if oldLines[oldIdx] == newLines[newIdx] {
			// Lines are the same
			diff = append(diff, DiffLine{
				Type:    DiffLineContext,
				Content: oldLines[oldIdx],
				LineNum: oldIdx + 1,
			})
			oldIdx++
			newIdx++
		} else {
			// Lines are different - show both
			diff = append(diff, DiffLine{
				Type:    DiffLineRemoved,
				Content: oldLines[oldIdx],
				LineNum: oldIdx + 1,
			})
			diff = append(diff, DiffLine{
				Type:    DiffLineAdded,
				Content: newLines[newIdx],
				LineNum: newIdx + 1,
			})
			oldIdx++
			newIdx++
		}
	}
	
	return df.optimizeDiff(diff)
}

// optimizeDiff removes unnecessary context lines and groups changes
func (df *DiffFormatter) optimizeDiff(lines []DiffLine) []DiffLine {
	if len(lines) == 0 {
		return lines
	}
	
	var optimized []DiffLine
	contextWindow := 3 // Show 3 lines of context around changes
	
	// Find all change positions
	changePositions := make([]bool, len(lines))
	for i, line := range lines {
		if line.Type != DiffLineContext {
			changePositions[i] = true
			// Mark surrounding context
			for j := max(0, i-contextWindow); j <= min(len(lines)-1, i+contextWindow); j++ {
				changePositions[j] = true
			}
		}
	}
	
	// Build optimized diff
	inSkipSection := false
	for i, line := range lines {
		if changePositions[i] {
			if inSkipSection {
				// End skip section
				inSkipSection = false
			}
			optimized = append(optimized, line)
		} else if !inSkipSection {
			// Start skip section
			inSkipSection = true
			if len(optimized) > 0 {
				optimized = append(optimized, DiffLine{
					Type:    DiffLineContext,
					Content: "...",
					LineNum: -1,
				})
			}
		}
	}
	
	return optimized
}

// countChanges counts added and removed lines
func (df *DiffFormatter) countChanges(lines []DiffLine) (added, removed int) {
	for _, line := range lines {
		switch line.Type {
		case DiffLineAdded:
			added++
		case DiffLineRemoved:
			removed++
		}
	}
	return
}

// FormatSimpleDiff creates a simple before/after diff
func (df *DiffFormatter) FormatSimpleDiff(filename, oldContent, newContent string) string {
	if oldContent == newContent {
		return df.headerColor.Sprintf("No changes in %s", filename)
	}
	
	var result strings.Builder
	result.WriteString(df.headerColor.Sprintf("Modified %s\n", filename))
	result.WriteString(df.contextColor.Sprintf("─%s\n", strings.Repeat("─", 50)))
	
	// Show a few lines of old content
	oldLines := strings.Split(strings.TrimSpace(oldContent), "\n")
	newLines := strings.Split(strings.TrimSpace(newContent), "\n")
	
	// Show removed lines
	if len(oldLines) > 0 && oldLines[0] != "" {
		result.WriteString(df.removedColor.Sprintf("- Removed:\n"))
		for i, line := range oldLines {
			if i >= 5 { // Limit to 5 lines
				result.WriteString(df.contextColor.Sprintf("  ... (%d more lines)\n", len(oldLines)-i))
				break
			}
			result.WriteString(df.removedColor.Sprintf("  %s\n", line))
		}
	}
	
	// Show added lines
	if len(newLines) > 0 && newLines[0] != "" {
		result.WriteString(df.addedColor.Sprintf("+ Added:\n"))
		for i, line := range newLines {
			if i >= 5 { // Limit to 5 lines
				result.WriteString(df.contextColor.Sprintf("  ... (%d more lines)\n", len(newLines)-i))
				break
			}
			result.WriteString(df.addedColor.Sprintf("  %s\n", line))
		}
	}
	
	result.WriteString(df.contextColor.Sprintf("─%s\n", strings.Repeat("─", 50)))
	return result.String()
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
