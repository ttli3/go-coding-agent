package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/ttli3/go-coding-agent/internal/agent"
	"github.com/ttli3/go-coding-agent/internal/commands"
	"github.com/ttli3/go-coding-agent/internal/config"
	"github.com/ttli3/go-coding-agent/internal/ui"
)

var (
	configPath string
	model      string
	stream     bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "agent_go [message]",
		Short: "Agent_Go - AI-powered terminal coding assistant",
		Long: `Agent_Go is a powerful AI coding assistant that can execute tasks using various tools.
It integrates with OpenRouter to provide access to multiple LLM models and includes
features like context management, file operations, and more.`,
		Run: runAgent,
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "Override model from config")
	rootCmd.Flags().BoolVarP(&stream, "stream", "s", true, "Enable streaming responses")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runAgent(cmd *cobra.Command, args []string) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		color.New(color.FgRed).Printf("Configuration error: %v\n", err)
		color.New(color.FgHiBlack).Println("Tip: Set your OpenRouter API key with: export OPENROUTER_API_KEY=\"your-key\"")
		os.Exit(1)
	}

	// override model if specified 
	if model != "" {
		cfg.OpenRouter.Model = model
	}

	// create agent
	aiAgent := agent.NewAgent(cfg)

	// print welcome message
	printWelcome(cfg)

	// handle direct command
	if len(args) > 0 {
		message := strings.Join(args, " ")
		handleMessage(aiAgent, message)
		return
	}

	runInteractiveMode(aiAgent)
}

func printWelcome(cfg *config.Config) {
	//welcome message
	fmt.Println()
	color.New(color.FgCyan, color.Bold).Println("Agent_Go")
	color.New(color.FgHiBlack).Printf("AI Coding Assistant • Model: %s\n", cfg.OpenRouter.Model)
	fmt.Println()
	color.New(color.FgHiBlack).Println("Type your message, '/help' for commands, or '/exit' to quit")
	color.New(color.FgHiBlack).Println(strings.Repeat("─", 60))
	fmt.Println()
}

func runInteractiveMode(aiAgent *agent.Agent) {
	scanner := bufio.NewScanner(os.Stdin)
	cmdRegistry := commands.NewDefaultRegistry()

	for {
		// Show context window status in prompt
		contextStatus := getContextStatus(aiAgent)
		color.New(color.FgCyan).Printf("> %s", contextStatus)
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle built-in commands first
		switch input {
		case "exit", "quit", "q":
			color.New(color.FgHiBlack).Println("\nGoodbye!")
			return
		}

		// Try to execute as a slash command
		ctx := &commands.CommandContext{
			Agent:    aiAgent,
			Registry: cmdRegistry,
		}
		result, isCommand, err := cmdRegistry.Execute(input, ctx)
		if isCommand {
			if err != nil {
				color.Red("Command error: %v", err)
			} else {
				// Handle exit command
				if result == "EXIT_APPLICATION" {
					color.New(color.FgHiBlack).Println("\nGoodbye!")
					return
				}
				color.Green("%s", result)
			}
			continue
		}

		// Handle as regular message
		handleMessage(aiAgent, input)
	}

	if err := scanner.Err(); err != nil {
		color.Red("Error reading input: %v", err)
	}
}

func determineActivityType(message string) string {
	messageLower := strings.ToLower(message)

	// Check for coding/implementation keywords
	codingKeywords := []string{"create", "build", "implement", "write", "code", "develop", "make", "generate", "add", "fix", "update", "modify", "change"}
	for _, keyword := range codingKeywords {
		if strings.Contains(messageLower, keyword) {
			return "coding"
		}
	}

	// Check for reading/viewing keywords
	readingKeywords := []string{"read", "show", "view", "display", "examine", "check", "look", "see", "what does", "explain", "describe"}
	for _, keyword := range readingKeywords {
		if strings.Contains(messageLower, keyword) {
			return "reading"
		}
	}

	// Check for searching/finding keywords
	searchingKeywords := []string{"find", "search", "locate", "look for", "where is", "list", "show me"}
	for _, keyword := range searchingKeywords {
		if strings.Contains(messageLower, keyword) {
			return "searching"
		}
	}

	// Default to thinking
	return "thinking"
}

func handleMessage(aiAgent *agent.Agent, message string) {
	if stream {
		handleStreamingMessage(aiAgent, message)
	} else {
		handleBlockingMessage(aiAgent, message)
	}
}

func handleStreamingMessage(aiAgent *agent.Agent, message string) {
	// determine activity type based on msg content
	activityType := determineActivityType(message)
	loading := ui.NewLoadingIndicator(activityType)
	loading.Start()

	formatter := ui.NewResponseFormatter()
	var responseBuffer strings.Builder
	firstChunk := true

	err := aiAgent.ProcessMessageStream(message, func(chunk string) {
		if firstChunk {
			loading.Stop()
			firstChunk = false
		}
		responseBuffer.WriteString(chunk)
	})

	loading.Stop()

	if err != nil {
		color.New(color.FgRed).Printf("\nError: %v\n", err)
		return
	}

	if responseBuffer.Len() > 0 {
		formattedResponse := formatter.FormatResponse(responseBuffer.String())
		fmt.Print(formattedResponse)
	}

	fmt.Println()
}

func handleBlockingMessage(aiAgent *agent.Agent, message string) {
	activityType := determineActivityType(message)
	loading := ui.NewLoadingIndicator(activityType)
	loading.Start()

	response, err := aiAgent.ProcessMessage(message)
	loading.Stop()

	if err != nil {
		color.New(color.FgRed).Printf("Error: %v\n", err)
		return
	}

	formatter := ui.NewResponseFormatter()
	formattedResponse := formatter.FormatResponse(response)
	fmt.Print(formattedResponse)
	fmt.Println()
}



// getContextStatus returns a formatted context window status for the prompt
func getContextStatus(aiAgent *agent.Agent) string {
	percent := aiAgent.GetContextUsagePercentage()
	usageStr := fmt.Sprintf("%.1f%%", percent)
	
	// Color code based on usage
	if percent < 50 {
		return color.New(color.FgGreen).Sprintf("[%s] ", usageStr)
	} else if percent < 80 {
		return color.New(color.FgYellow).Sprintf("[%s] ", usageStr)
	} else {
		return color.New(color.FgRed).Sprintf("[%s] ", usageStr)
	}
}
