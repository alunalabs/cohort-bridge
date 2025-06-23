# Cohort Bridge Makefile

# Variables
BINARY_NAME=cohort-tokenize
CLI_PKG=./cmd/cohort-tokenize
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Default target
.PHONY: all
all: build

# Build the CLI tool
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(CLI_PKG)

# Install the CLI tool
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) $(CLI_PKG)

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(CLI_PKG)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(CLI_PKG)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(CLI_PKG)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(CLI_PKG)

# Test the application
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f *.exe

# Run the CLI tool locally
.PHONY: run
run:
	go run $(CLI_PKG) $(ARGS)

# Show help for the CLI tool
.PHONY: help-cli
help-cli:
	go run $(CLI_PKG) -help

# Development dependencies
.PHONY: deps
deps:
	@echo "Installing development dependencies..."
	go mod download
	go mod tidy

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t cohort-tokenize:$(VERSION) .
	docker tag cohort-tokenize:$(VERSION) cohort-tokenize:latest

# Create release directory
dist:
	mkdir -p dist

# Build releases
.PHONY: release
release: clean dist build-all
	@echo "Creating release packages..."
	cd dist && \
	tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64 && \
	tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64 && \
	tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64 && \
	zip $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe

# Development workflow
.PHONY: dev
dev: deps build test

# Quick local test
.PHONY: test-local
test-local:
	@echo "Testing local build..."
	./$(BINARY_NAME) -version

# Show available targets
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       - Build the CLI tool"
	@echo "  install     - Install the CLI tool to GOPATH/bin"
	@echo "  build-all   - Build for multiple platforms"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  clean       - Clean build artifacts"
	@echo "  run         - Run the CLI tool locally (use ARGS='-help' for arguments)"
	@echo "  help-cli    - Show CLI tool help"
	@echo "  deps        - Install development dependencies"
	@echo "  docker-build- Build Docker image"
	@echo "  release     - Create release packages"
	@echo "  dev         - Development workflow (deps + build + test)"
	@echo "  test-local  - Test the local build"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make install"
	@echo "  make run ARGS='-database -main-config postgres.yaml'"
	@echo "  make build-all" 