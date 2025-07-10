package agent

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ttli3/go-coding-agent/internal/config"
	"github.com/ttli3/go-coding-agent/internal/context"
	"github.com/ttli3/go-coding-agent/internal/openrouter"
	"github.com/ttli3/go-coding-agent/internal/tools"
	"github.com/ttli3/go-coding-agent/internal/ui"
)

type Agent struct {
	client         *openrouter.Client
	toolRegistry   *tools.Registry
	Config         *config.Config
	sessionContext *context.SessionContext
	contextWindow  *context.ContextWindow
	verbose        bool
}

type ToolCallResponse struct {
	ToolCalls []ToolCallData `json:"tool_calls"`
}

type ToolCallData struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

func NewAgent(cfg *config.Config) *Agent {
	client := openrouter.NewClient(
		cfg.OpenRouter.APIKey,
		cfg.OpenRouter.BaseURL,
		cfg.OpenRouter.Model,
	)

	sessionCtx := context.NewSessionContext()
	sessionCtx.DetectProjectType()
	
	// Initialize context window with model-specific limits
	contextWindow := context.NewContextWindow(getModelContextLimit(cfg.OpenRouter.Model))

	return &Agent{
		client:         client,
		toolRegistry:   tools.GetDefaultRegistry(),
		Config:         cfg,
		sessionContext: sessionCtx,
		contextWindow:  contextWindow,
	}
}

func (a *Agent) AddMessage(role, content string) {
	// Add to context window with importance detection
	important := isImportantMessage(role, content)
	a.contextWindow.AddMessage(role, content, important)
}

func (a *Agent) GetSystemPrompt() string {
	return `You are Agent_Go, a powerful AI coding assistant that can execute tasks using various tools.

COMMUNICATION STYLE:
- Be direct and concise in your responses
- Avoid explaining what you're about to do - just do it
- Don't describe tool usage - the user can see tool execution details
- Focus on results and actionable information
- Skip phrases like "I'll help you", "Let me check", "I need to examine"
- Only provide context when it's essential for understanding

You have access to the following tools for file operations, code editing, and system commands:
- read_file: Read the contents of a file
- write_file: Write content to a file
- list_directory: List contents of a directory
- find_files: Find files matching a pattern
- edit_file: Edit file content at specific lines
- search_code: Search for patterns in files
- replace_content: Replace text patterns in files
- grep_search: Search across multiple files
- run_command: Execute system commands
- get_working_directory: Get current directory

When you need to use tools to complete a task, the system will automatically call the appropriate tools for you. Simply describe what you want to do in natural language.

Always:
- Explain what you're doing and why
- Break down complex tasks into steps
- Validate file paths before operations
- Provide clear feedback about results
- Follow coding best practices`
}

func (a *Agent) ProcessMessage(userMessage string) (string, error) {
	// Get messages from context window
	messages := a.GetConversationHistory()
	
	// Add system prompt if this is the first message
	if len(messages) == 0 {
		a.AddMessage("system", a.GetSystemPrompt())
	}

	// Add user message
	a.AddMessage("user", userMessage)

	// Get available tools for the API call
	availableTools := a.getOpenRouterTools()

	// Get updated messages after adding the new message
	messages = a.GetConversationHistory()
	
	// Call the agent
	response, err := a.client.Chat(
		messages,
		availableTools,
		a.Config.Agent.MaxTokens,
		a.Config.Agent.Temperature,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get AI response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	message := response.Choices[0].Message
	aiResponse := message.Content
	a.AddMessage("assistant", aiResponse)

	// check if the response contains tool calls from OpenRouter
	if len(message.ToolCalls) > 0 {
		return a.executeOpenRouterToolCalls(message.ToolCalls, aiResponse)
	}

	// fallback manually extract tool calls the agent wants to run
	textToolCalls := a.extractTextToolCalls(aiResponse)
	if len(textToolCalls) > 0 {
		return a.executeTextToolCalls(textToolCalls, aiResponse)
	}

	return aiResponse, nil
}

func (a *Agent) ProcessMessageStream(userMessage string, callback func(string)) error {
	// we need to use non-streaming first to get tool calls
	// then stream the follow-up response if needed
	response, err := a.ProcessMessage(userMessage)
	if err != nil {
		return err
	}

	// stream the response character by character for better UX
	for _, char := range response {
		callback(string(char))
	}

	return nil
}

func (a *Agent) getOpenRouterTools() []openrouter.Tool {
	availableTools := a.toolRegistry.List()
	orTools := make([]openrouter.Tool, len(availableTools))

	for i, tool := range availableTools {
		orTools[i] = openrouter.Tool{
			Type: "function",
			Function: openrouter.ToolFunction{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  a.convertSchema(tool.Schema()),
			},
		}
	}

	return orTools
}

func (a *Agent) convertSchema(schema tools.ToolSchema) map[string]interface{} {
	result := map[string]interface{}{
		"type":       schema.Type,
		"properties": schema.Properties,
	}

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	return result
}

func (a *Agent) extractToolCalls(response string) []ToolCallData {
	var toolCallResponse ToolCallResponse
	
	// Try to parse as JSON first
	if err := json.Unmarshal([]byte(response), &toolCallResponse); err == nil && len(toolCallResponse.ToolCalls) > 0 {
		return toolCallResponse.ToolCalls
	}
	
	// If not valid JSON, try to extract from text
	return a.extractTextToolCalls(response)
}

func (a *Agent) executeOpenRouterToolCalls(toolCalls []openrouter.ToolCall, originalResponse string) (string, error) {
	var results []string
	
	// Filter the original response based on verbosity setting
	filter := ui.NewResponseFilter()
	if originalResponse != "" {
		if a.verbose {
			// In verbose mode, show original response
			results = append(results, originalResponse)
		} else if !filter.ShouldSuppressResponse(originalResponse) {
			// In quiet mode, filter the response
			filteredResponse := filter.FilterResponse(originalResponse)
			if filteredResponse != "" {
				results = append(results, filteredResponse)
			}
		}
	}

	// Create tool execution display
	toolDisplay := ui.NewToolExecutionDisplay(len(toolCalls))

	// Execute all tools and collect results
	var toolResults []string
	for _, toolCall := range toolCalls {
		// Parse arguments from JSON string
		var args map[string]interface{}
		// Handle empty arguments
		if toolCall.Function.Arguments == "" || toolCall.Function.Arguments == "{}" {
			args = make(map[string]interface{})
		} else {
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				toolDisplay.StartTool(toolCall.Function.Name, args)
				toolDisplay.FinishTool(false, "", fmt.Errorf("failed to parse arguments: %v", err))
				continue
			}
		}

		// Show tool execution start
		toolDisplay.StartTool(toolCall.Function.Name, args)
		
		// Execute the tool
		result := a.toolRegistry.Execute(toolCall.Function.Name, args)

		// Show tool execution completion
		var err error
		if !result.Success && result.Error != "" {
			err = fmt.Errorf("%s", result.Error)
		}
		toolDisplay.FinishTool(result.Success, result.Result, err)

		if result.Success {
			toolResults = append(toolResults, fmt.Sprintf("%s result: %s", result.Name, result.Result))
		} else {
			toolResults = append(toolResults, fmt.Sprintf("%s error: %s", result.Name, result.Error))
		}
	}
	
	// show summary
	toolDisplay.ShowToolSummary()

	// continue conversation until task completion
	toolResultsMessage := strings.Join(toolResults, "\n")
	results = a.continueConversationUntilComplete(results, toolResultsMessage)

	return strings.Join(results, "\n"), nil
}

func (a *Agent) extractTextToolCalls(response string) []ToolCallData {
	// look for common patterns that indicate tool usage
	var toolCalls []ToolCallData

	// lookg for pattern 1: "I'll use the X tool" or "let me use X"
	toolPatterns := map[string]string{
		"read_file":      `(?i)(?:I'll|Let me|I will).*(?:read|check|look at|examine).*file.*(['"]([^'"]+)['"]|\b([\w./\-]+\.[\w]+)\b)`,
		"list_directory": `(?i)(?:I'll|Let me|I will).*(?:list|check|see).*(?:directory|folder|contents).*(['"]([^'"]+)['"]|current|\.)`,
		"find_files":     `(?i)(?:I'll|Let me|I will).*(?:find|search for|locate).*files.*pattern.*(['"]([^'"]+)['"])`,
		"run_command":    `(?i)(?:I'll|Let me|I will).*(?:run|execute).*command.*(['"]([^'"]+)['"])`,
	}

	for toolName, pattern := range toolPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(response)
		if len(matches) > 0 {
			// Extract the argument (file path, command, etc.)
			var arg string
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					arg = matches[i]
					break
				}
			}

			var args map[string]interface{}
			switch toolName {
			case "read_file":
				args = map[string]interface{}{"path": arg}
			case "list_directory":
				if arg == "current" || arg == "." || arg == "" {
					arg = "."
				}
				args = map[string]interface{}{"path": arg}
			case "find_files":
				args = map[string]interface{}{"path": ".", "pattern": arg}
			case "run_command":
				args = map[string]interface{}{"command": arg}
			}

			if args != nil {
				toolCalls = append(toolCalls, ToolCallData{
					Name:      toolName,
					Arguments: args,
				})
			}
		}
	}

	return toolCalls
}

func (a *Agent) executeTextToolCalls(toolCalls []ToolCallData, originalResponse string) (string, error) {
	var results []string
	results = append(results, originalResponse)
	results = append(results, "\nExecuting detected tools...")

	// Execute all tools and collect results
	var toolResults []string
	for _, toolCall := range toolCalls {
		result := a.toolRegistry.Execute(toolCall.Name, toolCall.Arguments)

		if result.Success {
			results = append(results, fmt.Sprintf("\n%s: %s", result.Name, result.Result))
			toolResults = append(toolResults, fmt.Sprintf("%s result: %s", result.Name, result.Result))
		} else {
			results = append(results, fmt.Sprintf("\n%s error: %s", result.Name, result.Error))
			toolResults = append(toolResults, fmt.Sprintf("%s error: %s", result.Name, result.Error))
		}
	}

	// Continue conversation until task completion
	toolResultsMessage := strings.Join(toolResults, "\n")
	results = a.continueConversationUntilComplete(results, toolResultsMessage)

	return strings.Join(results, "\n"), nil
}

func (a *Agent) ClearConversation() {
	a.contextWindow.ClearConversation()
}
func (a *Agent) GetConversationHistory() []openrouter.Message {
	// Convert context window messages to OpenRouter format
	contextMessages := a.contextWindow.GetContextualMessages()
	openRouterMessages := make([]openrouter.Message, 0, len(contextMessages))
	
	for _, msg := range contextMessages {
		openRouterMessages = append(openRouterMessages, openrouter.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	
	return openRouterMessages
}

func (a *Agent) GetCurrentModel() string {
	return a.Config.OpenRouter.Model
}

func (a *Agent) SetModel(model string) error {
	a.Config.OpenRouter.Model = model
	// Update the client with the new model
	a.client = openrouter.NewClient(
		a.Config.OpenRouter.APIKey,
		a.Config.OpenRouter.BaseURL,
		model,
	)
	return nil
}

// Context management methods

func (a *Agent) GetSessionContext() interface{} {
	return a.sessionContext
}

func (a *Agent) AddFocusedFile(filepath string) {
	a.sessionContext.AddFocusedFile(filepath)
}

func (a *Agent) ClearFocusedFiles() {
	a.sessionContext.ClearFocusedFiles()
}

func (a *Agent) GetFocusedFiles() []string {
	return a.sessionContext.FocusedFiles
}

func (a *Agent) SetCurrentTask(task string) {
	a.sessionContext.SetCurrentTask(task)
}

func (a *Agent) GetCurrentTask() string {
	return a.sessionContext.CurrentTask
}

func (a *Agent) CompleteCurrentTask(filesChanged []string) {
	a.sessionContext.CompleteCurrentTask(filesChanged)
}

func (a *Agent) SetBookmark(name, path string) {
	a.sessionContext.SetBookmark(name, path)
}

func (a *Agent) GetBookmark(name string) (string, bool) {
	return a.sessionContext.GetBookmark(name)
}

func (a *Agent) ListBookmarks() map[string]string {
	return a.sessionContext.Bookmarks
}

func (a *Agent) GetWorkspaceInfo() string {
	return fmt.Sprintf("Working Directory: %s\nProject Root: %s\nProject Type: %s",
		a.sessionContext.WorkingDir,
		a.sessionContext.ProjectRoot,
		a.sessionContext.ProjectType)
}

func (a *Agent) AddRecentFile(filepath string) {
	a.sessionContext.AddRecentFile(filepath)
}

func (a *Agent) GetContextStats() string {
	return a.contextWindow.GetContextStats()
}

func (a *Agent) GetContextUsagePercentage() float64 {
	return a.contextWindow.GetUsagePercentage()
}

func (a *Agent) SetVerbose(verbose bool) {
	a.verbose = verbose
}

func (a *Agent) IsVerbose() bool {
	return a.verbose
}

// getModelContextLimit returns the context limit for different models
func getModelContextLimit(model string) int {
	switch {
	case strings.Contains(model, "claude-3.5-sonnet"):
		return 200000
	case strings.Contains(model, "claude-3.5-haiku"):
		return 200000
	case strings.Contains(model, "claude-3-opus"):
		return 200000
	case strings.Contains(model, "gpt-4o"):
		return 128000
	case strings.Contains(model, "gpt-4-turbo"):
		return 128000
	case strings.Contains(model, "gemini-pro"):
		return 1000000
	case strings.Contains(model, "llama"):
		return 32000
	default:
		return 32000 // Conservative default
	}
}

// isImportantMessage determines if a message should be preserved
func isImportantMessage(role, content string) bool {
	if role == "system" {
		return true
	}
	
	content = strings.ToLower(content)
	
	// Mark as important if it contains key indicators
	importantKeywords := []string{
		"error", "failed", "success", "completed",
		"implement", "create", "build", "deploy",
		"task:", "goal:", "objective:",
		"important:", "note:", "warning:",
	}
	
	for _, keyword := range importantKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	
	return false
}

// continueConversationUntilComplete continues the conversation until task is complete
func (a *Agent) continueConversationUntilComplete(results []string, toolResultsMessage string) []string {
	a.AddMessage("user", fmt.Sprintf("Tool execution results:\n%s\n\nPlease analyze these results and continue with the task. If you need to use more tools or take additional actions to complete the user's request, please do so. If the task is complete, provide a summary of what was accomplished.", toolResultsMessage))

	// Continue the conversation until the task is complete
	for maxIterations := 0; maxIterations < 200; maxIterations++ { 
		followUpResponse, err := a.client.Chat(
			a.GetConversationHistory(),
			a.getOpenRouterTools(), // Include tools for continued execution
			a.Config.Agent.MaxTokens,
			a.Config.Agent.Temperature,
		)
		if err != nil {
			results = append(results, fmt.Sprintf("\nFailed to get follow-up response: %v", err))
			break
		}

		if len(followUpResponse.Choices) == 0 {
			break
		}

		followUpMessage := followUpResponse.Choices[0].Message
		followUp := followUpMessage.Content
		a.AddMessage("assistant", followUp)
		results = append(results, fmt.Sprintf("\n\n%s", followUp))

		// Check if follow-up response has tool calls - continue execution
		if len(followUpMessage.ToolCalls) > 0 {
			additionalResult, err := a.executeOpenRouterToolCalls(followUpMessage.ToolCalls, "")
			if err != nil {
				results = append(results, fmt.Sprintf("\nError executing additional tools: %v", err))
				break
			} else {
				results = append(results, fmt.Sprintf("\n%s", additionalResult))
			}
		} else {
			// No more tool calls, check if the AI indicates the task is complete
			// If the response doesn't contain action words, assume task is complete
			responseText := strings.ToLower(followUp)
			actionWords := []string{"let me", "i'll", "i will", "i need to", "next", "now i", "first", "then"}
			hasActionWords := false
			for _, word := range actionWords {
				if strings.Contains(responseText, word) {
					hasActionWords = true
					break
				}
			}

			if !hasActionWords {
				// Task appears to be complete
				break
			}

			a.AddMessage("user", "Is there anything else you need to do to complete this task? If so, please continue. If the task is complete, please confirm.")
		}
	}

	return results
}
