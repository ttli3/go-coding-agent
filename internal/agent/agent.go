package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

func NewAgent(cfg *config.Config) *Agent {
	client := openrouter.NewClient(
		cfg.OpenRouter.APIKey,
		cfg.OpenRouter.BaseURL,
		cfg.OpenRouter.Model,
	)

	sessionCtx := context.NewSessionContext()
	sessionCtx.DetectProjectType()
	
	contextWindow := context.NewContextWindow(getModelContextLimit(cfg.OpenRouter.Model))

	agent := &Agent{
		client:         client,
		toolRegistry:   tools.GetDefaultRegistry(),
		Config:         cfg,
		sessionContext: sessionCtx,
		contextWindow:  contextWindow,
	}
	
	// Automatically load previous session if it exists
	agent.LoadSession()
	
	return agent
}

func (a *Agent) AddMessage(role, content string) {
	important := isImportantMessage(role, content)
	a.contextWindow.AddMessage(role, content, important)
}

func (a *Agent) GetSystemPrompt() string {
	basePrompt := `You are Agent_Go, a powerful AI coding assistant that can execute tasks using function calls.

CRITICAL RULE: You MUST call functions to perform actions. NEVER say "Let me..." or "I'll..." or "I need to..." - just call the function immediately.

FORBIDDEN PHRASES:
- "Let me read..."
- "I'll check..."
- "I need to..."
- "Let me request..."
- "I should..."

INSTEAD: Just call the function directly without any description.

You have access to these functions:
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
- show_diff: Show differences between file versions

WORKFLOW:
1. User gives you a task
2. You IMMEDIATELY call the necessary functions (no description)
3. You continue calling functions until the task is complete
4. You provide a brief summary of what was accomplished

EXAMPLE:
User: "read the README file"
WRONG: "Let me read the README file for you."
CORRECT: [Immediately call read_file function]

NEVER describe what you're going to do - just do it by calling functions immediately.`

	// Add dynamic session context
	contextInfo := a.buildContextInfo()
	if contextInfo != "" {
		return basePrompt + "\n\n" + contextInfo
	}
	return basePrompt
}

func (a *Agent) ProcessMessage(userMessage string) (string, error) {
	messages := a.GetConversationHistory()
	
	if len(messages) == 0 {
		a.AddMessage("system", a.GetSystemPrompt())
	}

	a.AddMessage("user", userMessage)

	availableTools := a.getOpenRouterTools()
	messages = a.GetConversationHistory()
	
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

	if len(message.ToolCalls) > 0 {
		response, err := a.executeOpenRouterToolCalls(message.ToolCalls, aiResponse)
		// Auto-save session after tool execution
		a.autoSaveSession()
		return response, err
	}

	// Auto-save session after regular message processing
	a.autoSaveSession()
	return aiResponse, nil
}

func (a *Agent) ProcessMessageStream(userMessage string, callback func(string)) error {
	response, err := a.ProcessMessage(userMessage)
	if err != nil {
		return err
	}

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

func (a *Agent) executeOpenRouterToolCalls(toolCalls []openrouter.ToolCall, originalResponse string) (string, error) {
	var results []string
	
	filter := ui.NewResponseFilter()
	if originalResponse != "" {
		if a.verbose {
			results = append(results, originalResponse)
		} else if !filter.ShouldSuppressResponse(originalResponse) {
			filteredResponse := filter.FilterResponse(originalResponse)
			if filteredResponse != "" {
				results = append(results, filteredResponse)
			}
		}
	}

	currentToolCalls := toolCalls
	for len(currentToolCalls) > 0 {
		toolDisplay := ui.NewToolExecutionDisplay(len(currentToolCalls))

		var toolResults []string
		for _, toolCall := range currentToolCalls {
			var args map[string]interface{}
			argStr := strings.TrimSpace(toolCall.Function.Arguments)
			if argStr == "" || argStr == "{}" {
				args = make(map[string]interface{})
			} else {
				if err := json.Unmarshal([]byte(argStr), &args); err != nil {
					toolDisplay.StartTool(toolCall.Function.Name, make(map[string]interface{}))
					toolDisplay.FinishTool(false, "", fmt.Errorf("failed to parse arguments '%s': %v", argStr, err))
					continue
				}
			}

			toolDisplay.StartTool(toolCall.Function.Name, args)
			
			result := a.toolRegistry.Execute(toolCall.Function.Name, args)

			// Automatically track files for file-related operations
			if result.Success {
				a.trackFileOperation(toolCall.Function.Name, args)
			}

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
		
		toolDisplay.ShowToolSummary()

		toolResultsMessage := strings.Join(toolResults, "\n")
		a.AddMessage("user", fmt.Sprintf("Tool execution results:\n%s\n\nContinue task execution. Call the next required function immediately or provide task completion summary.", toolResultsMessage))

		followUpResponse, err := a.client.Chat(
			a.GetConversationHistory(),
			a.getOpenRouterTools(),
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

		if len(followUpMessage.ToolCalls) > 0 {
			currentToolCalls = followUpMessage.ToolCalls
		} else {
			break
		}
	}

	return strings.Join(results, "\n"), nil
}

func (a *Agent) ClearConversation() {
	a.contextWindow.ClearConversation()
}
func (a *Agent) GetConversationHistory() []openrouter.Message {
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
	a.client = openrouter.NewClient(
		a.Config.OpenRouter.APIKey,
		a.Config.OpenRouter.BaseURL,
		model,
	)
	return nil
}

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

// buildContextInfo creates dynamic context information for the AI
func (a *Agent) buildContextInfo() string {
	var info strings.Builder
	
	info.WriteString("CURRENT SESSION CONTEXT:\n")
	info.WriteString("========================\n")
	
	// Project information
	if a.sessionContext.ProjectRoot != "" {
		projectName := filepath.Base(a.sessionContext.ProjectRoot)
		info.WriteString(fmt.Sprintf("Project: %s", projectName))
		if a.sessionContext.ProjectType != "" && a.sessionContext.ProjectType != "unknown" {
			info.WriteString(fmt.Sprintf(" (%s)", a.sessionContext.ProjectType))
		}
		info.WriteString("\n")
	}
	
	// Working directory
	if a.sessionContext.WorkingDir != "" {
		info.WriteString(fmt.Sprintf("Working Directory: %s\n", a.sessionContext.WorkingDir))
	}
	
	// Current task
	if a.sessionContext.CurrentTask != "" {
		info.WriteString(fmt.Sprintf("Current Task: %s\n", a.sessionContext.CurrentTask))
	}
	
	// Recently focused files (top 3)
	if len(a.sessionContext.FocusedFiles) > 0 {
		info.WriteString("Recently Active Files:\n")
		maxFiles := len(a.sessionContext.FocusedFiles)
		if maxFiles > 3 {
			maxFiles = 3
		}
		for i := 0; i < maxFiles; i++ {
			file := a.sessionContext.FocusedFiles[i]
			info.WriteString(fmt.Sprintf("  %d. %s\n", i+1, file))
		}
		if len(a.sessionContext.FocusedFiles) > 3 {
			info.WriteString(fmt.Sprintf("  ... and %d more files\n", len(a.sessionContext.FocusedFiles)-3))
		}
	}
	
	// Recent files (top 3, excluding focused files)
	if len(a.sessionContext.RecentFiles) > 0 {
		recentFiles := []string{}
		focusedSet := make(map[string]bool)
		for _, f := range a.sessionContext.FocusedFiles {
			focusedSet[f] = true
		}
		
		for _, file := range a.sessionContext.RecentFiles {
			if !focusedSet[file] && len(recentFiles) < 3 {
				recentFiles = append(recentFiles, file)
			}
		}
		
		if len(recentFiles) > 0 {
			info.WriteString("Other Recent Files:\n")
			for i, file := range recentFiles {
				info.WriteString(fmt.Sprintf("  %d. %s\n", i+1, file))
			}
		}
	}
	
	// Bookmarks
	if len(a.sessionContext.Bookmarks) > 0 {
		info.WriteString(fmt.Sprintf("Available Bookmarks: %d\n", len(a.sessionContext.Bookmarks)))
	}
	
	// Task history summary
	if len(a.sessionContext.TaskHistory) > 0 {
		info.WriteString(fmt.Sprintf("Completed Tasks: %d\n", len(a.sessionContext.TaskHistory)))
		lastTask := a.sessionContext.TaskHistory[len(a.sessionContext.TaskHistory)-1]
		info.WriteString(fmt.Sprintf("Last Completed: %s\n", lastTask.Description))
	}
	
	contextStr := info.String()
	// Only return context if we have meaningful information
	if strings.Contains(contextStr, "Project:") || 
	   strings.Contains(contextStr, "Current Task:") || 
	   strings.Contains(contextStr, "Recently Active Files:") {
		return contextStr
	}
	
	return ""
}

// trackFileOperation automatically tracks files when tools interact with them
func (a *Agent) trackFileOperation(toolName string, args map[string]interface{}) {
	switch toolName {
	case "read_file", "write_file":
		if path, ok := args["path"].(string); ok && path != "" {
			// Convert to absolute path if possible
			if absPath, err := filepath.Abs(path); err == nil {
				path = absPath
			}
			a.sessionContext.AddFocusedFile(path)
			a.sessionContext.AddRecentFile(path)
		}
	case "edit_file":
		if path, ok := args["path"].(string); ok && path != "" {
			if absPath, err := filepath.Abs(path); err == nil {
				path = absPath
			}
			a.sessionContext.AddFocusedFile(path)
			a.sessionContext.AddRecentFile(path)
		}
	case "find_files":
		// For find_files, we don't track individual files since it's a search operation
		// But we could track the search directory as a recent location
		if path, ok := args["path"].(string); ok && path != "" {
			if absPath, err := filepath.Abs(path); err == nil {
				path = absPath
			}
			// Add the search directory to recent files (as a directory context)
			a.sessionContext.AddRecentFile(path)
		}
	case "list_directory":
		if path, ok := args["path"].(string); ok && path != "" {
			if absPath, err := filepath.Abs(path); err == nil {
				path = absPath
			}
			a.sessionContext.AddRecentFile(path)
		}
	}
}

// autoSaveSession automatically saves the current session to a temporary file
func (a *Agent) autoSaveSession() {
	// Create a session file in the user's home directory or temp directory
	sessionFile := a.getSessionFilePath()
	if sessionFile == "" {
		return // Skip if we can't determine a good location
	}
	
	// Save session context (ignore errors for auto-save)
	a.sessionContext.SaveToFile(sessionFile)
}

// getSessionFilePath returns the path where session should be saved
func (a *Agent) getSessionFilePath() string {
	// Try user's home directory first
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, ".agent_go_session.json")
	}
	
	// Fallback to temp directory
	return filepath.Join(os.TempDir(), "agent_go_session.json")
}

// LoadSession attempts to load a previously saved session
func (a *Agent) LoadSession() {
	sessionFile := a.getSessionFilePath()
	if sessionFile == "" {
		return
	}
	
	// Check if session file exists
	if _, err := os.Stat(sessionFile); os.IsNotExist(err) {
		return // No previous session to load
	}
	
	// Load the session (ignore errors - we'll just start fresh if loading fails)
	if loadedSession, err := context.LoadFromFile(sessionFile); err == nil {
		a.sessionContext = loadedSession
		// Re-detect project type in case the project has changed
		a.sessionContext.DetectProjectType()
	}
}

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
		return 32000
	}
}

func isImportantMessage(role, content string) bool {
	if role == "system" {
		return true
	}
	
	content = strings.ToLower(content)
	
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