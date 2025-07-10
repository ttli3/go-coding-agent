package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
)

// CommandPrompt handles user confirmation for command execution
type CommandPrompt struct {
	formatter *ResponseFormatter
}

// NewCommandPrompt creates a new command prompt handler
func NewCommandPrompt() *CommandPrompt {
	return &CommandPrompt{
		formatter: NewResponseFormatter(),
	}
}

// ConfirmCommand displays a command and asks for user confirmation
func (cp *CommandPrompt) ConfirmCommand(command, workingDir string) bool {
	// Create a visually distinct command block
	fmt.Println()
	color.New(color.FgYellow, color.Bold).Println("COMMAND EXECUTION REQUEST")
	fmt.Println(strings.Repeat("─", 60))
	
	// Display command details
	color.New(color.FgCyan, color.Bold).Print("Command: ")
	color.New(color.FgWhite, color.BgBlack).Printf(" %s ", command)
	fmt.Println()
	
	if workingDir != "" {
		color.New(color.FgCyan, color.Bold).Print("Working Directory: ")
		color.New(color.FgWhite).Println(workingDir)
	}
	
	fmt.Println(strings.Repeat("─", 60))
	
	// Safety warning for potentially dangerous commands
	if cp.isDangerousCommand(command) {
		color.New(color.FgRed, color.Bold).Println("WARNING: This command may modify your system!")
		fmt.Println()
	}
	
	// Prompt for confirmation
	color.New(color.FgGreen, color.Bold).Print("Do you want to execute this command? ")
	color.New(color.FgWhite).Print("[y/N]: ")
	
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// isDangerousCommand checks if a command might be dangerous
func (cp *CommandPrompt) isDangerousCommand(command string) bool {
	dangerousCommands := []string{
		"rm", "rmdir", "del", "delete",
		"mv", "move", "cp", "copy",
		"chmod", "chown", "sudo",
		"dd", "fdisk", "mkfs",
		"kill", "killall", "pkill",
		"shutdown", "reboot", "halt",
		"format", "diskpart",
		"git reset --hard", "git clean -fd",
	}
	
	commandLower := strings.ToLower(command)
	for _, dangerous := range dangerousCommands {
		if strings.Contains(commandLower, dangerous) {
			return true
		}
	}
	
	return false
}

// DisplayCommandResult formats and displays command execution results
func (cp *CommandPrompt) DisplayCommandResult(command, workingDir, output string, success bool) {
	fmt.Println()
	
	if success {
		color.New(color.FgGreen, color.Bold).Println("COMMAND EXECUTED SUCCESSFULLY")
	} else {
		color.New(color.FgRed, color.Bold).Println("COMMAND EXECUTION FAILED")
	}
	
	fmt.Println(strings.Repeat("─", 60))
	
	// Display command details
	color.New(color.FgCyan, color.Bold).Print("Command: ")
	color.New(color.FgWhite, color.BgBlack).Printf(" %s ", command)
	fmt.Println()
	
	if workingDir != "" {
		color.New(color.FgCyan, color.Bold).Print("Working Directory: ")
		color.New(color.FgWhite).Println(workingDir)
	}
	
	fmt.Println(strings.Repeat("─", 60))
	
	// Display output
	if output != "" {
		color.New(color.FgCyan, color.Bold).Println("Output:")
		
		// Display output with simple formatting
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			color.New(color.FgWhite).Println(line)
		}
	} else {
		color.New(color.FgWhite).Println("(No output)")
	}
	
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
}

// looksLikeCode determines if output should be syntax highlighted
func (cp *CommandPrompt) looksLikeCode(output string) bool {
	codeIndicators := []string{
		"#!/", "function", "def ", "class ", "import ", "package ",
		"<html", "<?xml", "{", "}", "[", "]", "=>", "->",
	}
	
	outputLower := strings.ToLower(output)
	for _, indicator := range codeIndicators {
		if strings.Contains(outputLower, indicator) {
			return true
		}
	}
	
	return false
}
