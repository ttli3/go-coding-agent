package commands

import (
	"fmt"
	"strings"
	"sync"
)

// Command represents a slash command
type Command interface {
	Name() string
	Description() string
	Usage() string
	Execute(args []string, ctx *CommandContext) (string, error)
}

// CommandContext provides context for command execution
type CommandContext struct {
	Agent    interface{}       // Will be *agent.Agent but avoiding circular import
	Registry *CommandRegistry  // Access to command registry for help
}

// CommandRegistry manages all available commands
type CommandRegistry struct {
	commands map[string]Command
	mu       sync.RWMutex
}

// NewCommandRegistry creates a new command registry
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name()] = cmd
}

// Execute parses and executes a command
func (r *CommandRegistry) Execute(input string, ctx *CommandContext) (string, bool, error) {
	if !strings.HasPrefix(input, "/") {
		return "", false, nil // Not a command
	}

	// Parse command and arguments
	parts := strings.Fields(input[1:]) // Remove leading "/"
	if len(parts) == 0 {
		return "", true, fmt.Errorf("empty command")
	}

	cmdName := parts[0]
	args := parts[1:]

	r.mu.RLock()
	cmd, exists := r.commands[cmdName]
	r.mu.RUnlock()
	
	if !exists {
		return "", true, fmt.Errorf("unknown command: /%s", cmdName)
	}

	result, err := cmd.Execute(args, ctx)
	return result, true, err
}

// ListCommands returns all available commands
func (r *CommandRegistry) ListCommands() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	commands := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		commands = append(commands, cmd)
	}
	return commands
}

// GetCommand returns a specific command by name
func (r *CommandRegistry) GetCommand(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	cmd, exists := r.commands[name]
	return cmd, exists
}


