package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ttli3/go-coding-agent/internal/ui"
)

// RunCommandTool executes system commands
type RunCommandTool struct{}

func (t *RunCommandTool) Name() string {
	return "run_command"
}

func (t *RunCommandTool) Description() string {
	return "Execute a system command and return its output"
}

func (t *RunCommandTool) Execute(args map[string]interface{}) (string, error) {
	command, ok := args["command"].(string)
	if !ok {
		return "", fmt.Errorf("command parameter is required and must be a string")
	}

	workingDir := ""
	if val, exists := args["working_dir"]; exists {
		if wd, ok := val.(string); ok {
			workingDir = wd
		}
	}

	timeout := 30 * time.Second
	if val, exists := args["timeout"]; exists {
		if timeoutVal, ok := val.(float64); ok {
			timeout = time.Duration(timeoutVal) * time.Second
		}
	}

	// Ask for user confirmation before executing the command
	prompt := ui.NewCommandPrompt()
	if !prompt.ConfirmCommand(command, workingDir) {
		return "Command execution cancelled by user.", nil
	}

	// Split command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Set timeout
	if timeout > 0 {
		go func() {
			time.Sleep(timeout)
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}()
	}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	
	// Display the result using the formatted UI
	prompt.DisplayCommandResult(command, workingDir, outputStr, err == nil)
	
	if err != nil {
		return fmt.Sprintf("Command failed: %v\nOutput: %s", err, outputStr), fmt.Errorf("command execution failed: %w", err)
	}

	// Return a simple success message since the detailed output was already displayed
	return fmt.Sprintf("Command executed successfully: %s", command), nil
}

func (t *RunCommandTool) Schema() ToolSchema {
	return ToolSchema{
		Type: "object",
		Properties: map[string]PropertyDefinition{
			"command": {
				Type:        "string",
				Description: "Command to execute",
			},
			"working_dir": {
				Type:        "string",
				Description: "Working directory for the command (optional)",
			},
			"timeout": {
				Type:        "number",
				Description: "Timeout in seconds (default: 30)",
			},
		},
		Required: []string{"command"},
	}
}

// GetWorkingDirectoryTool gets the current working directory
type GetWorkingDirectoryTool struct{}

func (t *GetWorkingDirectoryTool) Name() string {
	return "get_working_directory"
}

func (t *GetWorkingDirectoryTool) Description() string {
	return "Get the current working directory"
}

func (t *GetWorkingDirectoryTool) Execute(args map[string]interface{}) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	return fmt.Sprintf("Current working directory: %s", wd), nil
}

func (t *GetWorkingDirectoryTool) Schema() ToolSchema {
	return ToolSchema{
		Type:       "object",
		Properties: map[string]PropertyDefinition{},
		Required:   []string{},
	}
}
