package commands

import (
	"fmt"
	"strings"
)

// VerbosityCommand controls the verbosity of AI responses
type VerbosityCommand struct{}

func (v *VerbosityCommand) Name() string {
	return "verbose"
}

func (v *VerbosityCommand) Description() string {
	return "Control verbosity of AI responses (on/off)"
}

func (v *VerbosityCommand) Usage() string {
	return "/verbose [on|off]"
}

func (v *VerbosityCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	// Define interface for agents that support verbosity control
	type VerbosityController interface {
		SetVerbose(bool)
		IsVerbose() bool
	}

	controller, ok := ctx.Agent.(VerbosityController)
	if !ok {
		return "", fmt.Errorf("agent does not support verbosity control")
	}

	// If no arguments, toggle current state
	if len(args) == 0 {
		newState := !controller.IsVerbose()
		controller.SetVerbose(newState)
		if newState {
			return "Verbose mode enabled. AI will provide detailed responses.", nil
		}
		return "Quiet mode enabled. AI responses will be more concise.", nil
	}

	// With argument, set specific state
	arg := strings.ToLower(args[0])
	switch arg {
	case "on", "true", "1", "yes":
		controller.SetVerbose(true)
		return "Verbose mode enabled. AI will provide detailed responses.", nil
	case "off", "false", "0", "no":
		controller.SetVerbose(false)
		return "Quiet mode enabled. AI responses will be more concise.", nil
	default:
		return "", fmt.Errorf("invalid argument: %s. Use 'on' or 'off'", arg)
	}
}
