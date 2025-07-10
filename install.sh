#!/bin/bash

# Agent_Go Installation Script
set -e

echo "Installing Agent_Go..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.21 or higher first."
    echo "   Visit: https://golang.org/doc/install"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "Go version $GO_VERSION is too old. Please upgrade to Go 1.21 or higher."
    exit 1
fi

echo "Go version $GO_VERSION detected"

# Install the package
echo "Installing Agent_Go..."
go install github.com/ttli3/go-coding-agent/cmd@latest

# Check if GOPATH/bin is in PATH
GOPATH_BIN=$(go env GOPATH)/bin
if [[ ":$PATH:" != *":$GOPATH_BIN:"* ]]; then
    echo " WARNING: $GOPATH_BIN is not in your PATH"
    echo "   Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo "   export PATH=\$PATH:$GOPATH_BIN"
    echo ""
fi

echo "Agent_Go installed successfully!"
echo ""
echo " Quick start:"
echo "   1. Set up your OpenRouter API key in ~/.agent_go.yaml"
echo "   2. Run: cmd"
echo "   3. Or create an alias: alias agent='cmd'"
echo ""
echo " For more information, visit: https://github.com/ttli3/go-coding-agent"
