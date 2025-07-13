# Agent_Go

Agent_Go is a terminal-based AI coding agent/assistant written in Go, heavily inspired by Claude Code and OpenCode. Built to help with developer tasks through natural language interaction and functional tool execution. The main goal of this project was to highlight and understand how powerful LLMs can be when enabled with tools/environment interaction + smart session memory.

## Features

- **Terminal-Based Interface**: Beautiful terminal UI with color-coded responses
- **Tool Execution**: Read files, list dirs, edit code, run commands, etc.
- **Smart Session Memory**: Automatically tracks files and context without manual commands
- **Persistent Sessions**: Sessions are auto-saved and restored between runs
- **Dynamic System Prompts**: Context is automatically injected into AI prompts
- **Context Management**: Tracks conversation context with token management
- **Consistent Command Interface**: All commands use slash prefix for consistency

### Available Commands

- **Session/chat Management**: 
  - `/clear` - Clear the conversation history
  - `/exit` - Exit the application
- **Context & Focus**: 
  - `/context` - Show session context, manage tasks, view stats
    - `/context stats` - Show context window statistics
    - `/context task <description>` - Set current task
    - `/context task clear` - Clear current task
    - `/context task complete` - Complete current task
  - `/focus <files...>` - Set focus to specific files
    - `/focus clear` - Clear focused files
- **LLM Behavior Config**: 
  - `/model [model-name]` - Switch AI model
- **System Info**: 
  - `/help [command]` - Show help information
  - `/history` - Show command history

## Installation

### Prerequisites

- Go 1.21 or higher
- OpenRouter API key (https://openrouter.ai/settings/keys)

### Quick Install (Recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/ttli3/go-coding-agent/main/install.sh | bash
```

### Install via go install

```bash
go install github.com/ttli3/go-coding-agent/cmd@latest
```

This will install the `cmd` binary to your `$GOPATH/bin` directory. Make sure `$GOPATH/bin` is in your `PATH`.

**Note:** The binary will be named `cmd`. You may want to create an alias:
```bash
# Add to your ~/.bashrc, ~/.zshrc, etc.
alias agent="cmd"
```

### Building from Source

```bash
git clone https://github.com/ttli3/go-coding-agent.git
cd go-coding-agent

# Using Makefile (recommended)
make build    # builds ./agent_go
make install  # installs to GOPATH/bin

# Or manually
go build -o agent_go ./cmd
go install ./cmd
```

### Configuration

Create a `.agent_go.yaml` file in your home directory:

```yaml
openrouter:
  api_key: "your_openrouter_api_key"  # Required: Get from https://openrouter.ai/settings/keys
  model: "anthropic/claude-3-7-sonnet"  # Optional: Default model
  timeout: 30                          # Optional: Request timeout

ui:
  colors: true     # Optional: Enable colored output
```

## Usage

### Quick Start

```bash
# Start the agent (if installed via go install)
cmd

# Or if built from source
./agent_go

# Start with a specific message
cmd "help me refactor this function"

# Use in any directory - the agent will automatically detect your project type
# and track files you interact with
```

### Session Persistence

Your session is automatically saved to `~/.agent_go_session.json` and will be restored the next time you start the agent. This includes:

- Recently focused files
- Current and completed tasks
- Project information
- Working directory

### Example Commands
```
/focus main.go  # Manually focus on a specific file (optional with auto-tracking)
/context        # View current session context (files, tasks, etc.)
/context task "Fix error handling"  # Set current task
/model claude-3-sonnet  # Switch to a different model
/help           # Show all available commands
/exit           # Exit the application
```

### AI Tool Execution

Agent_Go enables the AI to perform actions on your system through tool execution:

- **File Operations**: Read, write, and edit files
- **Directory Operations**: List directories, find files
- **Command Execution**: Run shell commands and capture output
- **Code Analysis**: Search for patterns, analyze code structure

All file interactions are automatically tracked in the session memory, building context for the AI without manual intervention.

## Smart Session Memory

### Automatic File Tracking

- Automatically tracks files when AI tools interact with them
- No need to manually set focus or remember important files
- Tracks files for operations like read_file, write_file, edit_file, find_files, and list_directory
- Converts paths to absolute paths for consistency

### Persistent Sessions

- Sessions are automatically saved to `~/.agent_go_session.json` after each interaction
- Sessions are automatically loaded when the agent starts
- Graceful session loading - if loading fails, agent starts fresh
- All context is preserved between application restarts

### Dynamic Context Injection

The system prompt is dynamically enhanced with session context including:
- Project name and type
- Working directory
- Current task (if any)
- Recently active files (top 3)
- Other recent files (excluding focused ones)
- Task history summary

## Context Window Management

- Tracks token/context-window usage with percentage in the UI
- Preserves important messages (system messages, errors, task instructions)
- Summarizes removed messages to maintain context
- Trims older messages when approaching token limits

## License

Licensed under MIT License
