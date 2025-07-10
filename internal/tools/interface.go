package tools

import (
	"encoding/json"
	"fmt"
)

// Tool represents a tool that can be called by the AI agent
type Tool interface {
	Name() string
	Description() string
	Execute(args map[string]interface{}) (string, error)
	Schema() ToolSchema
}

// ToolSchema defines the JSON schema for a tool's parameters
type ToolSchema struct {
	Type       string                        `json:"type"`
	Properties map[string]PropertyDefinition `json:"properties"`
	Required   []string                      `json:"required"`
}

// PropertyDefinition defines a parameter property
type PropertyDefinition struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolCall represents a tool call from the AI
type ToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Name    string `json:"name"`
	Result  string `json:"result"`
	Error   string `json:"error,omitempty"`
	Success bool   `json:"success"`
}

// Registry manages all available tools
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get retrieves a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// List returns all registered tools
func (r *Registry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Execute runs a tool with the given arguments
func (r *Registry) Execute(name string, args map[string]interface{}) *ToolResult {
	tool, exists := r.tools[name]
	if !exists {
		return &ToolResult{
			Name:    name,
			Error:   fmt.Sprintf("tool '%s' not found", name),
			Success: false,
		}
	}

	result, err := tool.Execute(args)
	if err != nil {
		return &ToolResult{
			Name:    name,
			Error:   err.Error(),
			Success: false,
		}
	}

	return &ToolResult{
		Name:    name,
		Result:  result,
		Success: true,
	}
}

// ParseToolCalls extracts tool calls from AI response content
func ParseToolCalls(content string) ([]ToolCall, error) {
	var toolCalls []ToolCall

	// Look for JSON blocks that contain tool calls
	// This is a simplified parser - in practice, you might need more sophisticated parsing
	// depending on how your AI model formats tool calls

	// Try to parse the entire content as a tool call first
	var singleCall ToolCall
	if err := json.Unmarshal([]byte(content), &singleCall); err == nil {
		return []ToolCall{singleCall}, nil
	}

	// Try to parse as an array of tool calls
	if err := json.Unmarshal([]byte(content), &toolCalls); err == nil {
		return toolCalls, nil
	}

	// If no valid JSON found, return empty slice
	return []ToolCall{}, nil
}

// GetDefaultRegistry returns a registry with all default tools registered
func GetDefaultRegistry() *Registry {
	registry := NewRegistry()

	// Register all default tools
	registry.Register(&ReadFileTool{})
	registry.Register(&WriteFileTool{})
	registry.Register(&ListDirectoryTool{})
	registry.Register(&EditFileTool{})
	registry.Register(&SearchCodeTool{})
	registry.Register(&ReplaceContentTool{})
	registry.Register(&RunCommandTool{})
	registry.Register(&GetWorkingDirectoryTool{})
	registry.Register(&FindFilesTool{})
	registry.Register(&GrepSearchTool{})
	registry.Register(&DiffTool{})

	return registry
}
