package commands

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ContextCommand shows current session context
type ContextCommand struct{}

func (c *ContextCommand) Name() string {
	return "context"
}

func (c *ContextCommand) Description() string {
	return "Manage session context, tasks, and view stats"
}

func (c *ContextCommand) Usage() string {
	return "/context [stats|task <description>|task clear|task complete]"
}

func (c *ContextCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	// Handle stats subcommand
	if len(args) > 0 && args[0] == "stats" {
		type StatsProvider interface {
			GetContextStats() string
		}

		if provider, ok := ctx.Agent.(StatsProvider); ok {
			return provider.GetContextStats(), nil
		}
		return "Context statistics not available", nil
	}

	// Handle task subcommands
	if len(args) > 0 && args[0] == "task" {
		type TaskManager interface {
			SetCurrentTask(string)
			GetCurrentTask() string
			CompleteCurrentTask([]string)
		}

		manager, ok := ctx.Agent.(TaskManager)
		if !ok {
			return "", fmt.Errorf("agent does not support task management")
		}

		// Show current task if no additional arguments
		if len(args) == 1 {
			currentTask := manager.GetCurrentTask()
			if currentTask == "" {
				return "No current task set", nil
			}
			return fmt.Sprintf("Current task: %s", currentTask), nil
		}

		// Handle task clear
		if len(args) == 2 && args[1] == "clear" {
			manager.SetCurrentTask("")
			return "Cleared current task", nil
		}

		// Handle task complete
		if len(args) == 2 && args[1] == "complete" {
			currentTask := manager.GetCurrentTask()
			if currentTask == "" {
				return "No current task to complete", nil
			}
			// For now, we're not tracking changed files
			manager.CompleteCurrentTask([]string{})
			return fmt.Sprintf("Completed task: %s", currentTask), nil
		}

		// Set new task
		taskDescription := strings.Join(args[1:], " ")
		manager.SetCurrentTask(taskDescription)
		return fmt.Sprintf("Set current task: %s", taskDescription), nil
	}

	// Handle other subcommands or invalid arguments
	if len(args) > 0 {
		return "", fmt.Errorf("unknown subcommand: %s. Use '/context', '/context stats', or '/context task'", args[0])
	}

	// Get session context from agent
	type ContextProvider interface {
		GetSessionContext() interface{}
	}

	if provider, ok := ctx.Agent.(ContextProvider); ok {
		sessionCtx := provider.GetSessionContext()
		if sessionCtx != nil {
			// Use reflection or type assertion to get context summary
			if sc, ok := sessionCtx.(interface{ GetContextSummary() string }); ok {
				return sc.GetContextSummary(), nil
			}
		}
	}

	return "Session context not available", nil
}

// FocusCommand manages focused files
type FocusCommand struct{}

func (f *FocusCommand) Name() string {
	return "focus"
}

func (f *FocusCommand) Description() string {
	return "Set focus to specific files"
}

func (f *FocusCommand) Usage() string {
	return "/focus <file1> [file2] ... or /focus clear"
}

func (f *FocusCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	type ContextManager interface {
		AddFocusedFile(string)
		ClearFocusedFiles()
		GetFocusedFiles() []string
	}

	manager, ok := ctx.Agent.(ContextManager)
	if !ok {
		return "", fmt.Errorf("agent does not support context management")
	}

	if len(args) == 0 {
		// Show current focused files
		focused := manager.GetFocusedFiles()
		if len(focused) == 0 {
			return "No files currently focused", nil
		}
		
		var result strings.Builder
		result.WriteString(fmt.Sprintf("Focused files (%d):\n", len(focused)))
		for i, file := range focused {
			result.WriteString(fmt.Sprintf("%d. %s\n", i+1, file))
		}
		return result.String(), nil
	}

	if len(args) == 1 && args[0] == "clear" {
		manager.ClearFocusedFiles()
		return "Cleared all focused files", nil
	}

	// Add files to focus
	var added []string
	for _, file := range args {
		// Convert to absolute path if relative
		absPath, err := filepath.Abs(file)
		if err != nil {
			absPath = file // Use as-is if conversion fails
		}
		manager.AddFocusedFile(absPath)
		added = append(added, absPath)
	}

	return fmt.Sprintf("Added %d files to focus:\n%s", len(added), strings.Join(added, "\n")), nil
}

// TaskCommand manages current task
type TaskCommand struct{}

func (t *TaskCommand) Name() string {
	return "task"
}

func (t *TaskCommand) Description() string {
	return "Set or show current task"
}

func (t *TaskCommand) Usage() string {
	return "/task [description] or /task clear or /task complete"
}

func (t *TaskCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	type TaskManager interface {
		SetCurrentTask(string)
		GetCurrentTask() string
		CompleteCurrentTask([]string)
	}

	manager, ok := ctx.Agent.(TaskManager)
	if !ok {
		return "", fmt.Errorf("agent does not support task management")
	}

	if len(args) == 0 {
		// Show current task
		currentTask := manager.GetCurrentTask()
		if currentTask == "" {
			return "No current task set", nil
		}
		return fmt.Sprintf("Current task: %s", currentTask), nil
	}

	if len(args) == 1 && args[0] == "clear" {
		manager.SetCurrentTask("")
		return "Cleared current task", nil
	}

	if len(args) == 1 && args[0] == "complete" {
		currentTask := manager.GetCurrentTask()
		if currentTask == "" {
			return "No current task to complete", nil
		}
		// For now, we're not tracking changed files
		// In the future, we could add a GetRecentFiles method to the interface
		manager.CompleteCurrentTask([]string{})
		return fmt.Sprintf("Completed task: %s", currentTask), nil
	}

	// Set new task
	taskDescription := strings.Join(args, " ")
	manager.SetCurrentTask(taskDescription)
	return fmt.Sprintf("Set current task: %s", taskDescription), nil
}




// ContextStatsCommand shows context window statistics
type ContextStatsCommand struct{}

func (c *ContextStatsCommand) Name() string {
	return "stats"
}

func (c *ContextStatsCommand) Description() string {
	return "Show context window statistics"
}

func (c *ContextStatsCommand) Usage() string {
	return "/stats"
}

func (c *ContextStatsCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("stats command takes no arguments")
	}

	type StatsProvider interface {
		GetContextStats() string
	}

	if provider, ok := ctx.Agent.(StatsProvider); ok {
		return provider.GetContextStats(), nil
	}

	return "Context statistics not available", nil
}
