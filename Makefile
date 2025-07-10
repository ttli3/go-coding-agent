.PHONY: build install clean test fmt vet

# Build the binary
build:
	go build -o agent_go ./cmd

# Install the binary to GOPATH/bin
install:
	go install ./cmd

# Clean build artifacts
clean:
	rm -f agent_go
	go clean

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run all checks
check: fmt vet test

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o agent_go-linux-amd64 ./cmd
	GOOS=darwin GOARCH=amd64 go build -o agent_go-darwin-amd64 ./cmd
	GOOS=darwin GOARCH=arm64 go build -o agent_go-darwin-arm64 ./cmd
	GOOS=windows GOARCH=amd64 go build -o agent_go-windows-amd64.exe ./cmd

# Development build with race detection
dev:
	go build -race -o agent_go ./cmd
