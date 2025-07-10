package ui

import (
	"fmt"
	"regexp"
	"strings"
)

// ResponseFilter cleans up and formats AI responses
type ResponseFilter struct {
	// Patterns to remove or simplify
	verbosePatterns []string
	cleanupRules    map[string]string
}

// NewResponseFilter creates a new response filter
func NewResponseFilter() *ResponseFilter {
	return &ResponseFilter{
		verbosePatterns: []string{
			// Remove excessive explanations about what the AI is doing
			`(?i)I'll (now )?(?:help you|assist you|proceed to|start by|begin by|first).*?\.`,
			`(?i)Let me (?:help you|assist you|check|examine|look at|analyze).*?\.`,
			`(?i)I need to (?:check|examine|look at|analyze|understand).*?\.`,
			`(?i)(?:First|Now|Next), I'll (?:need to|have to).*?\.`,
			
			// Remove redundant file operation descriptions
			`(?i)I'll (?:read|examine|check) the (?:file|contents of).*?\.`,
			`(?i)Let me (?:read|examine|check) the (?:file|contents of).*?\.`,
			
			// Remove excessive tool call explanations
			`(?i)I'll use the .* tool to.*?\.`,
			`(?i)Let me use the .* tool to.*?\.`,
			`(?i)Let's use the .* tool to.*?\.`,
			`(?i)I'll call the .* tool to.*?\.`,
			`(?i)Let me call the .* tool to.*?\.`,
			`(?i)Now I'll use the .* tool.*?\.`,
			`(?i)I will use the .* tool.*?\.`,
			`(?i)Let me find .* and view its contents.*?\.`,
			`(?i)I'll find .* and view its contents.*?\.`,
			
			// Remove redundant confirmations
			`(?i)(?:Perfect|Great|Excellent|Good)! (?:I can see|I found|I notice).*?\.`,
			`(?i)Now (that )?I have (?:all|the) (?:necessary|required) information.*?\.`,
			`(?i)Based on the (?:code|file|output) (?:above|I've examined).*?\.`,
		},
		cleanupRules: map[string]string{
			// Simplify common verbose phrases
			`(?i)I can see that`: "I see",
			`(?i)I notice that`: "I notice",
			`(?i)It appears that`: "It appears",
			`(?i)It looks like`: "It looks like",
			`(?i)I understand that`: "I understand",
			`(?i)Based on (?:my analysis|what I can see)`: "Based on this",
			`(?i)Let me (?:help you|assist you) (?:with|by)`: "I'll",
			`(?i)I'll (?:help you|assist you) (?:with|by)`: "I'll",
			
			// Tool usage simplifications
			`(?i)Let me check if the diff is working`: "Testing the diff tool",
			`(?i)Let me update the response filter to better handle`: "Updating response filter for",
			`(?i)Now I'll build the project to make sure our changes work`: "Building the project",
			`(?i)Let's build the project to verify`: "Building to verify",
			`(?i)Let me examine the`: "Checking the",
			`(?i)I'll examine the`: "Checking the",
			`(?i)Let me continue examining`: "Continuing with",
			`(?i)Let me check how the`: "Checking how the",
			`(?i)I'll check how the`: "Checking how the",
			`(?i)Let me look at the`: "Looking at the",
			`(?i)I'll look at the`: "Looking at the",
		},
	}
}

// FilterResponse cleans up a response from the AI
func (rf *ResponseFilter) FilterResponse(response string) string {
	// Don't filter if response is very short
	if len(response) < 100 {
		return response
	}
	
	// Preserve formatted diffs and code blocks
	// First, extract and save code blocks and diffs
	preservedBlocks := make(map[string]string)
	blockID := 0
	
	// Preserve diff blocks (they typically have +/- and line formatting)
	diffPattern := regexp.MustCompile(`(?s)(Changes in .*?\n─+\n.*?lines changed\n)`)
	filtered := diffPattern.ReplaceAllStringFunc(response, func(match string) string {
		placeholder := fmt.Sprintf("__PRESERVED_BLOCK_%d__", blockID)
		preservedBlocks[placeholder] = match
		blockID++
		return placeholder
	})
	
	// Apply cleanup rules (simple replacements)
	for pattern, replacement := range rf.cleanupRules {
		re := regexp.MustCompile(pattern)
		filtered = re.ReplaceAllString(filtered, replacement)
	}
	
	// Remove verbose patterns
	for _, pattern := range rf.verbosePatterns {
		re := regexp.MustCompile(pattern)
		filtered = re.ReplaceAllString(filtered, "")
	}
	
	// Clean up multiple spaces and newlines (but not in preserved blocks)
	filtered = regexp.MustCompile(`\s+`).ReplaceAllString(filtered, " ")
	filtered = regexp.MustCompile(`\n\s*\n\s*\n`).ReplaceAllString(filtered, "\n\n")
	
	// Restore preserved blocks
	for placeholder, block := range preservedBlocks {
		filtered = strings.Replace(filtered, placeholder, block, 1)
	}
	
	// Remove leading/trailing whitespace
	filtered = strings.TrimSpace(filtered)
	
	// If we removed too much, return original
	if len(filtered) < len(response)/3 {
		return response
	}
	
	return filtered
}

// ShouldSuppressResponse determines if a response should be suppressed entirely
func (rf *ResponseFilter) ShouldSuppressResponse(response string) bool {
	response = strings.ToLower(strings.TrimSpace(response))
	
	// Suppress very generic responses
	suppressPatterns := []string{
		`^(?:ok|okay|sure|alright|got it|understood)\.?$`,
		`^(?:i'll|i will) (?:help|assist) (?:you )?(?:with )?(?:that|this)\.?$`,
		`^(?:let me|i'll) (?:check|look at|examine) (?:that|this)\.?$`,
		`^(?:analyzing|checking|examining|processing)\.{0,3}$`,
	}
	
	for _, pattern := range suppressPatterns {
		if matched, _ := regexp.MatchString(pattern, response); matched {
			return true
		}
	}
	
	return false
}

// ExtractActionableContent extracts the main actionable content from a response
func (rf *ResponseFilter) ExtractActionableContent(response string) string {
	lines := strings.Split(response, "\n")
	var actionableLines []string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Skip purely descriptive lines
		if rf.isDescriptiveLine(line) {
			continue
		}
		
		actionableLines = append(actionableLines, line)
	}
	
	return strings.Join(actionableLines, "\n")
}

// isDescriptiveLine checks if a line is purely descriptive
func (rf *ResponseFilter) isDescriptiveLine(line string) bool {
	line = strings.ToLower(line)
	
	descriptivePatterns := []string{
		`^(?:i'll|i will|let me|i need to|i'm going to)`,
		`^(?:first|next|then|now|after that)`,
		`^(?:this (?:will|should)|that (?:will|should))`,
		`^(?:the (?:file|code|function|method)).*(?:contains|has|shows|indicates)`,
	}
	
	for _, pattern := range descriptivePatterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			return true
		}
	}
	
	return false
}
