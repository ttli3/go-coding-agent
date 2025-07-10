package commands

import (
	"fmt"
)

// ClearCommand clears the chat history
type ClearCommand struct{}

func (c *ClearCommand) Name() string {
	return "clear"
}

func (c *ClearCommand) Description() string {
	return "Clear the chat history and start fresh"
}

func (c *ClearCommand) Usage() string {
	return "/clear"
}

func (c *ClearCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	if len(args) > 0 {
		return "", fmt.Errorf("clear command takes no arguments")
	}

	// Use interface{} to call ClearConversation - we'll need to define an interface
	type ConversationClearer interface {
		ClearConversation()
	}

	if clearer, ok := ctx.Agent.(ConversationClearer); ok {
		clearer.ClearConversation()
		return "Chat history cleared successfully", nil
	}

	return "", fmt.Errorf("agent does not support clearing conversation")
}
