package commands

import (
	"fmt"
	"strings"
)

// ModelCommand switches the AI model
type ModelCommand struct{}

func (m *ModelCommand) Name() string {
	return "model"
}

func (m *ModelCommand) Description() string {
	return "Switch the AI model"
}

func (m *ModelCommand) Usage() string {
	return "/model <model-name> or /model to see current model"
}

func (m *ModelCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	// Define interface for agents that have configurable models
	type ModelConfigurable interface {
		GetCurrentModel() string
		SetModel(string) error
	}

	// If no arguments, show current model
	if len(args) == 0 {
		if configurable, ok := ctx.Agent.(ModelConfigurable); ok {
			currentModel := configurable.GetCurrentModel()
			return fmt.Sprintf("Current model: %s", currentModel), nil
		}
		return "", fmt.Errorf("agent does not support model configuration")
	}

	// Set new model
	newModel := strings.Join(args, " ")
	
	// Validate model name (basic validation)
	validModels := []string{
		"anthropic/claude-3-5-sonnet-20241022",
		"anthropic/claude-3-5-haiku-20241022", 
		"anthropic/claude-4-opus-20240229",
		"openai/gpt-4o",
		"openai/gpt-4o-mini",
		"openai/gpt-4-turbo",
		"google/gemini-pro-1.5",
		"meta-llama/llama-3.1-405b-instruct",
		"meta-llama/llama-3.1-70b-instruct",
		"meta-llama/llama-3.1-8b-instruct",
	}

	isValid := false
	for _, valid := range validModels {
		if newModel == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		return "", fmt.Errorf("invalid model: %s\nValid models: %s", newModel, strings.Join(validModels, ", "))
	}

	// Try to set the model
	if configurable, ok := ctx.Agent.(ModelConfigurable); ok {
		oldModel := configurable.GetCurrentModel()
		err := configurable.SetModel(newModel)
		if err != nil {
			return "", fmt.Errorf("failed to set model: %w", err)
		}
		return fmt.Sprintf("Model changed from %s to %s", oldModel, newModel), nil
	}

	return "", fmt.Errorf("agent does not support model configuration")
}
