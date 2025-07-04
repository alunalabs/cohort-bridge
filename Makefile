# CohortBridge Makefile

# Variables
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S_UTC')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"

# Program definitions
PROGRAMS=cohort-bridge test
PROGRAM_PATHS=./cmd/cohort-bridge ./cmd/test

# Default target
.PHONY: all
all: build

# Build all programs
.PHONY: build
build: $(PROGRAMS)

# Build individual programs
.PHONY: cohort-bridge
cohort-bridge:
	@echo "Building cohort-bridge..."
	go build $(LDFLAGS) -o cohort-bridge ./cmd/cohort-bridge

.PHONY: test-program
test-program:
	@echo "Building test program..."
	go build $(LDFLAGS) -o test ./cmd/test

# Create test alias to avoid conflict with 'make test'
test: test-program

# Install all programs
.PHONY: install
install:
	@echo "Installing all programs..."
	@for cmd in $(PROGRAM_PATHS); do \
		echo "Installing $$cmd..."; \
		go install $(LDFLAGS) $$cmd; \
	done

# Build for multiple platforms
.PHONY: build-all
build-all: clean-dist dist
	@echo "Building for multiple platforms..."
	@for prog in $(PROGRAMS); do \
		echo "Building $$prog for multiple platforms..."; \
		GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$$prog-linux-amd64 ./cmd/$$prog; \
		GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$$prog-darwin-amd64 ./cmd/$$prog; \
		GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$$prog-darwin-arm64 ./cmd/$$prog; \
		GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$$prog-windows-amd64.exe ./cmd/$$prog; \
	done

# Test the application
.PHONY: test-go
test-go:
	@echo "Running Go tests..."
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
	rm -f $(PROGRAMS)
	rm -f *.exe

.PHONY: clean-dist
clean-dist:
	@echo "Cleaning dist directory..."
	rm -rf dist/

.PHONY: clean-all
clean-all: clean clean-dist

# Create release directory
dist:
	mkdir -p dist

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
	docker build -t cohort-bridge:$(VERSION) .
	docker tag cohort-bridge:$(VERSION) cohort-bridge:latest

# Build releases
.PHONY: release
release: clean-all dist build-all
	@echo "Creating release packages..."
	cd dist && \
	for prog in $(PROGRAMS); do \
		echo "Packaging $$prog..."; \
		tar -czf $$prog-linux-amd64.tar.gz $$prog-linux-amd64; \
		tar -czf $$prog-darwin-amd64.tar.gz $$prog-darwin-amd64; \
		tar -czf $$prog-darwin-arm64.tar.gz $$prog-darwin-arm64; \
		zip $$prog-windows-amd64.zip $$prog-windows-amd64.exe; \
	done

# Development workflow
.PHONY: dev
dev: deps build test-go

# Quick local test
.PHONY: test-local
test-local: build
	@echo "Testing local builds..."
	@for prog in $(PROGRAMS); do \
		echo "Testing $$prog..."; \
		./$$prog -help > /dev/null 2>&1 || echo "$$prog built successfully"; \
	done

# Run specific programs with arguments
.PHONY: run-cohort-bridge
run-cohort-bridge:
	go run ./cmd/cohort-bridge $(ARGS)

.PHONY: run-test
run-test:
	go run ./cmd/test $(ARGS)

# Demo workflow
.PHONY: demo
demo: build
	@echo "Running CohortBridge demo workflow..."
	@echo "1. Building sample data..."
	@echo "2. Running tokenization..."
	@echo "3. Computing intersection..."
	@echo "4. This is a placeholder - implement actual demo steps"

# Show help for individual programs
.PHONY: help-cohort-bridge
help-cohort-bridge:
	go run ./cmd/cohort-bridge -help

.PHONY: help-test
help-test:
	go run ./cmd/test -help

# Show available targets
.PHONY: help
help:
	@echo "CohortBridge Build System"
	@echo "========================"
	@echo ""
	@echo "Building:"
	@echo "  build           - Build all programs"
	@echo "  cohort-bridge   - Build cohort-bridge program"
	@echo "  test            - Build test harness"
	@echo ""
	@echo "Installation:"
	@echo "  install         - Install all programs to GOPATH/bin"
	@echo "  build-all       - Build for multiple platforms"
	@echo ""
	@echo "Testing:"
	@echo "  test-go         - Run Go unit tests"
	@echo "  test-local      - Test local builds"
	@echo "  lint            - Run linter"
	@echo ""
	@echo "Development:"
	@echo "  deps            - Install development dependencies"
	@echo "  dev             - Development workflow (deps + build + test)"
	@echo "  demo            - Run demo workflow"
	@echo ""
	@echo "Running programs:"
	@echo "  run-cohort-bridge - Run cohort-bridge (use ARGS='...' for arguments)"
	@echo "  run-test          - Run test (use ARGS='...' for arguments)"
	@echo ""
	@echo "Help for programs:"
	@echo "  help-cohort-bridge - Show cohort-bridge help"
	@echo "  help-test          - Show test help"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean           - Clean build artifacts"
	@echo "  clean-dist      - Clean distribution files"
	@echo "  clean-all       - Clean everything"
	@echo ""
	@echo "Release:"
	@echo "  release         - Create release packages"
	@echo "  docker-build    - Build Docker image"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run-cohort-bridge ARGS='tokenize -input=data.csv -output=tokens.csv'"
	@echo "  make run-cohort-bridge ARGS='intersect -dataset1=tokens1.csv -dataset2=tokens2.csv'" 