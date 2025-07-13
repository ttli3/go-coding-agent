package commands

import (
	"fmt"
)

// ExitCommand exits the application
type ExitCommand struct{}

func (e *ExitCommand) Name() string {
	return "exit"
}

func (e *ExitCommand) Description() string {
	return "Exit the application"
}

func (e *ExitCommand) Usage() string {
	return "/exit"
}

func (e *ExitCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("exit command takes no arguments")
	}

	// Return a special marker that the main loop can detect
	return "EXIT_APPLICATION", nil
}
