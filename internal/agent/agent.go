package agent

import (
	"encoding/json"
	"fmt"
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

	return &Agent{
		client:         client,
		toolRegistry:   tools.GetDefaultRegistry(),
		Config:         cfg,
		sessionContext: sessionCtx,
		contextWindow:  contextWindow,
	}
}

func (a *Agent) AddMessage(role, content string) {
	important := isImportantMessage(role, content)
	a.contextWindow.AddMessage(role, content, important)
}

func (a *Agent) GetSystemPrompt() string {
	return `You are Agent_Go, a powerful AI coding assistant that can execute tasks using function calls.

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
}

func (a *Agent) ProcessMessage(userMessage string) (string, error) {
	messages := a.GetConversationHistory()
	
	if len(messages) == 0 {
		a.AddMessage("system", a.GetSystemPrompt())
	}

	a.AddMessage("user", userMessage)

	availableTools := a.getOpenRouterTools()
	fmt.Printf("[DEBUG] Available tools: %d\n", len(availableTools))
	for _, tool := range availableTools {
		fmt.Printf("[DEBUG] Tool: %s - %s\n", tool.Function.Name, tool.Function.Description)
	}

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
	fmt.Printf("[DEBUG] AI Response: %s\n", aiResponse)
	fmt.Printf("[DEBUG] Tool calls received: %d\n", len(message.ToolCalls))
	for i, toolCall := range message.ToolCalls {
		fmt.Printf("[DEBUG] Tool call %d: %s(%s)\n", i, toolCall.Function.Name, toolCall.Function.Arguments)
	}
	a.AddMessage("assistant", aiResponse)

	if len(message.ToolCalls) > 0 {
		return a.executeOpenRouterToolCalls(message.ToolCalls, aiResponse)
	}

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