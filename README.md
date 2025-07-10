# Agent_Go

Agent_Go is a terminal-based AI coding agent/assistant written in Go, heavily inspired by Claude Code and OpenCode. Built to help with developer tasks through natural language interaction and functional tool execution. The main goal of this project was to highlight and understand how powerful LLMs can be when enabled with tools/environment interaction + session context.

## Features

- **Terminal-Based Interface**: beautiful terminal UI with color-coded responses
- **Tool Execution**: read files, list dirs, edit code, run commands, etc..
- **Context Management**: tracks conversation context with token management
- **Session Context**: maintains awareness of focused files, current tasks, and project structure
- **Slash Commands**: comprehensive command system for controlling agentic behavior and enabling visibility

### Available Commands

- **Session/chat Management**: `/clear`
- **Context & Focus**: 
  - `/context` - Show session context, manage tasks, view stats
    - `/context stats` - Show context window statistics
    - `/context task <description>` - Set current task
    - `/context task clear` - Clear current task
    - `/context task complete` - Complete current task
  - `/focus <files...>` - Set focus to specific files
    - `/focus clear` - Clear focused files
- **Workspace & Navigation**: 
  - `/workspace` - Show workspace information
  - `/bookmark <name> <path>` - Set bookmarks
    - `/bookmark list` - List bookmarks
    - `/bookmark goto <name>` - Navigate to bookmark
- **LLM Behavior Config**: 
  - `/model [model-name]` - Switch AI model
  - `/verbose [on|off]` - Control response verbosity
- **System Info**: `/help [command]` - Show help information

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

```bash
# Copy the example configuration
cp .agent_go.example.yaml ~/.agent_go.yaml

# Edit with your API key
vim ~/.agent_go.yaml  # or use your preferred editor
```

**Configuration file (`~/.agent_go.yaml`):**
```yaml
openrouter:
  api_key: "your_openrouter_api_key"  # Required: Get from https://openrouter.ai/settings/keys
  model: "anthropic/claude-3-5-sonnet"  # Optional: Default model
  timeout: 30                          # Optional: Request timeout

ui:
  colors: true     # Optional: Enable colored output
  verbose: false   # Optional: Default verbosity
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

# Use in any directory - the agent will detect your project type
cd /path/to/your/project
cmd
```

### Using in Your Own Repositories

The agent works in any directory and automatically detects your project type:

```bash
# Navigate to any project
cd /path/to/your/go-project
cmd "help me optimize this function"

# Works with different project types
cd /path/to/your/python-project
cmd "add error handling to this script"

# Focus on specific files for context
cmd
> /focus main.py utils.py
> "refactor these files to use dependency injection"
```

### Slash Commands

```bash
/help           # show available commands
/focus file.go  # focus on a specific file
/context        # show current context
/context stats  # Show context window usage
/model          # switch AI model
/verbose on     # enable verbose mode
/workspace      # show workspace info
```

## Context Window Management

Tracks token/context-window usage with percentage in the UI
Preserves important messages (system messages, errors, task instructions)
Summarizes removed messages to maintain context
Trims older messages when approaching token limits


## License

Licensed under MIT License
