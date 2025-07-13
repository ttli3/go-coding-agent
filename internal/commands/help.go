package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
)

// HelpCommand displays available commands and usage information
type HelpCommand struct{}

func (c *HelpCommand) Name() string {
	return "help"
}

func (c *HelpCommand) Description() string {
	return "Show available commands and usage information"
}

func (c *HelpCommand) Usage() string {
	return "/help [command]"
}

func (c *HelpCommand) Execute(args []string, ctx *CommandContext) (string, error) {
	if len(args) > 0 {
		// Show help for specific command
		commandName := args[0]
		cmd, exists := ctx.Registry.GetCommand(commandName)
		if exists {
			return c.formatCommandHelp(cmd), nil
		}
		return fmt.Sprintf("Command '%s' not found. Use /help to see all available commands.", commandName), nil
	}

	// Show all commands grouped by category
	return c.formatAllCommands(ctx.Registry), nil
}

func (c *HelpCommand) formatCommandHelp(cmd Command) string {
	var result strings.Builder
	
	color.New(color.FgCyan, color.Bold).Fprintf(&result, "Command: /%s\n", cmd.Name())
	result.WriteString(fmt.Sprintf("Description: %s\n", cmd.Description()))
	result.WriteString(fmt.Sprintf("Usage: %s\n", cmd.Usage()))
	
	return result.String()
}

func (c *HelpCommand) formatAllCommands(registry *CommandRegistry) string {
	var result strings.Builder
	
	// Header
	color.New(color.FgCyan, color.Bold).Fprint(&result, "Agent_Go - Available Commands\n")
	color.New(color.FgHiBlack).Fprint(&result, strings.Repeat("─", 50)+"\n\n")
	
	categories := map[string][]Command{
		"Chat & Session Management": {},
		"Context & Focus": {},
		"Model Control": {},
		"System Information": {},
	}
	
	// Get all commands and categorize them
	commands := registry.ListCommands()
	for _, cmd := range commands {
		switch cmd.Name() {
		case "clear", "exit":
			categories["Chat & Session Management"] = append(categories["Chat & Session Management"], cmd)
		case "context", "focus", "task", "stats":
			categories["Context & Focus"] = append(categories["Context & Focus"], cmd)
		case "model":
			categories["Model Control"] = append(categories["Model Control"], cmd)
		case "help", "history":
			categories["System Information"] = append(categories["System Information"], cmd)
		default:
			categories["System Information"] = append(categories["System Information"], cmd)
		}
	}
	
	// Display each category
	categoryOrder := []string{
		"Chat & Session Management",
		"Context & Focus", 
		"Model Control",
		"System Information",
	}
	
	for _, categoryName := range categoryOrder {
		cmds := categories[categoryName]
		if len(cmds) == 0 {
			continue
		}
		
		// Sort commands within category
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})
		
		// Category header
		color.New(color.FgYellow, color.Bold).Fprintf(&result, "%s:\n", categoryName)
		
		// Commands in category
		for _, cmd := range cmds {
			color.New(color.FgGreen).Fprintf(&result, "  /%s", cmd.Name())
			color.New(color.FgHiBlack).Fprintf(&result, " - %s\n", cmd.Description())
		}
		result.WriteString("\n")
	}
	
	// Footer with usage tips
	color.New(color.FgHiBlack).Fprint(&result, "Usage Tips:\n")
	color.New(color.FgHiBlack).Fprint(&result, "• Use /help <command> for detailed help on a specific command\n")
	color.New(color.FgHiBlack).Fprint(&result, "• Context window usage is shown in your prompt: ")
	color.New(color.FgGreen).Fprint(&result, "[25.3%] ")
	color.New(color.FgHiBlack).Fprint(&result, "agent_go> \n")
	color.New(color.FgHiBlack).Fprint(&result, "• Use /focus to set files for the AI to pay attention to\n")
	
	return result.String()
}
