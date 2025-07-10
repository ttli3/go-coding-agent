package context

import (
	"fmt"
	"strings"
	"time"
)

// ContextWindow manages conversation context within token limits
type ContextWindow struct {
	MaxTokens        int                    `json:"max_tokens"`
	ReservedTokens   int                    `json:"reserved_tokens"`   // Reserved for system prompt + tools
	SummaryTokens    int                    `json:"summary_tokens"`    // Tokens used for conversation summary
	Messages         []ConversationMessage  `json:"messages"`
	ConversationSummary string             `json:"conversation_summary"`
	ImportantMessages []ConversationMessage `json:"important_messages"` // Always keep these
}

// ConversationMessage represents a message with metadata
type ConversationMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Tokens    int    `json:"tokens"`
	Timestamp int64  `json:"timestamp"`
	Important bool   `json:"important"` // Mark as important to preserve
	MessageID string `json:"message_id"`
}

// NewContextWindow creates a new context window manager
func NewContextWindow(maxTokens int) *ContextWindow {
	return &ContextWindow{
		MaxTokens:         maxTokens,
		ReservedTokens:    2000, // Reserve for system prompt and tools
		SummaryTokens:     500,  // Reserve for conversation summary
		Messages:          []ConversationMessage{},
		ImportantMessages: []ConversationMessage{},
	}
}

// EstimateTokens provides a rough token estimate (4 chars ≈ 1 token)
func EstimateTokens(text string) int {
	return len(text) / 4
}

// AddMessage adds a message to the context window
func (cw *ContextWindow) AddMessage(role, content string, important bool) {
	tokens := EstimateTokens(content)
	message := ConversationMessage{
		Role:      role,
		Content:   content,
		Tokens:    tokens,
		Timestamp: getCurrentTimestamp(),
		Important: important,
		MessageID: generateMessageID(),
	}
	
	if important {
		cw.ImportantMessages = append(cw.ImportantMessages, message)
	}
	
	cw.Messages = append(cw.Messages, message)
	cw.trimIfNeeded()
}

// trimIfNeeded trims the conversation if it exceeds token limits
func (cw *ContextWindow) trimIfNeeded() {
	availableTokens := cw.MaxTokens - cw.ReservedTokens - cw.SummaryTokens
	currentTokens := cw.calculateCurrentTokens()
	
	if currentTokens <= availableTokens {
		return // No trimming needed
	}
	
	// Create summary of messages to be removed
	messagesToSummarize := []ConversationMessage{}
	messagesToKeep := []ConversationMessage{}
	
	// Always keep important messages and recent messages
	recentCount := 10 // Keep last 10 messages
	totalMessages := len(cw.Messages)
	
	for i, msg := range cw.Messages {
		isRecent := i >= totalMessages-recentCount
		if msg.Important || isRecent {
			messagesToKeep = append(messagesToKeep, msg)
		} else {
			messagesToSummarize = append(messagesToSummarize, msg)
		}
	}
	
	// Create summary of removed messages
	if len(messagesToSummarize) > 0 {
		cw.ConversationSummary = cw.createSummary(messagesToSummarize)
	}
	
	cw.Messages = messagesToKeep
}

// createSummary creates a summary of messages being removed
func (cw *ContextWindow) createSummary(messages []ConversationMessage) string {
	if len(messages) == 0 {
		return cw.ConversationSummary
	}
	
	var summary strings.Builder
	if cw.ConversationSummary != "" {
		summary.WriteString(cw.ConversationSummary + "\n\n")
	}
	
	summary.WriteString(fmt.Sprintf("=== Conversation Summary (%d messages) ===\n", len(messages)))
	
	// Group by conversation topics/tasks
	taskMessages := []string{}
	codeMessages := []string{}
	otherMessages := []string{}
	
	for _, msg := range messages {
		content := strings.ToLower(msg.Content)
		if strings.Contains(content, "task") || strings.Contains(content, "implement") || strings.Contains(content, "create") {
			taskMessages = append(taskMessages, fmt.Sprintf("- %s: %s", msg.Role, truncateContent(msg.Content, 100)))
		} else if strings.Contains(content, "file") || strings.Contains(content, "code") || strings.Contains(content, "function") {
			codeMessages = append(codeMessages, fmt.Sprintf("- %s: %s", msg.Role, truncateContent(msg.Content, 100)))
		} else {
			otherMessages = append(otherMessages, fmt.Sprintf("- %s: %s", msg.Role, truncateContent(msg.Content, 100)))
		}
	}
	
	if len(taskMessages) > 0 {
		summary.WriteString("Tasks/Implementation:\n")
		for _, msg := range taskMessages {
			summary.WriteString(msg + "\n")
		}
		summary.WriteString("\n")
	}
	
	if len(codeMessages) > 0 {
		summary.WriteString("Code/File Operations:\n")
		for _, msg := range codeMessages {
			summary.WriteString(msg + "\n")
		}
		summary.WriteString("\n")
	}
	
	if len(otherMessages) > 0 {
		summary.WriteString("Other Discussion:\n")
		for _, msg := range otherMessages {
			summary.WriteString(msg + "\n")
		}
	}
	
	return summary.String()
}

// GetContextualMessages returns messages formatted for AI consumption
func (cw *ContextWindow) GetContextualMessages() []ConversationMessage {
	result := []ConversationMessage{}
	
	// Add conversation summary if exists
	if cw.ConversationSummary != "" {
		result = append(result, ConversationMessage{
			Role:    "system",
			Content: cw.ConversationSummary,
			Tokens:  EstimateTokens(cw.ConversationSummary),
		})
	}
	
	// Add current messages
	result = append(result, cw.Messages...)
	
	return result
}

// calculateCurrentTokens calculates total tokens in current context
func (cw *ContextWindow) calculateCurrentTokens() int {
	total := 0
	for _, msg := range cw.Messages {
		total += msg.Tokens
	}
	if cw.ConversationSummary != "" {
		total += EstimateTokens(cw.ConversationSummary)
	}
	return total
}

// GetContextStats returns statistics about the context window
func (cw *ContextWindow) GetContextStats() string {
	currentTokens := cw.calculateCurrentTokens()
	availableTokens := cw.MaxTokens - cw.ReservedTokens
	
	return fmt.Sprintf(`Context Window Stats:
- Current tokens: %d
- Available tokens: %d
- Max tokens: %d
- Reserved tokens: %d
- Messages: %d
- Important messages: %d
- Has summary: %v
- Usage: %.1f%%`,
		currentTokens,
		availableTokens,
		cw.MaxTokens,
		cw.ReservedTokens,
		len(cw.Messages),
		len(cw.ImportantMessages),
		cw.ConversationSummary != "",
		cw.GetUsagePercentage())
}

// returns the context window usage as a percentage
func (cw *ContextWindow) GetUsagePercentage() float64 {
	currentTokens := cw.calculateCurrentTokens()
	availableTokens := cw.MaxTokens - cw.ReservedTokens
	if availableTokens <= 0 {
		return 100.0
	}
	return float64(currentTokens) / float64(availableTokens) * 100
}

// marks a recent message as important
func (cw *ContextWindow) MarkMessageImportant(messageID string) {
	for i := range cw.Messages {
		if cw.Messages[i].MessageID == messageID {
			cw.Messages[i].Important = true
			cw.ImportantMessages = append(cw.ImportantMessages, cw.Messages[i])
			break
		}
	}
}

// ClearConversation clears all messages but keeps important ones
func (cw *ContextWindow) ClearConversation() {
	cw.Messages = cw.ImportantMessages
	cw.ConversationSummary = ""
}

//helpers 
func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

func generateMessageID() string {
	return fmt.Sprintf("msg_%d_%d", time.Now().Unix(), time.Now().UnixNano()%1000)
}
