# Makefile for deadlinkr

# configuration variables
BINARY_NAME=deadlinkr
GO=go
GOLANGCI_LINT=golangci-lint

# directories
SRC_DIR=.
BUILD_DIR=./build
COVERAGE_DIR=./coverage

# compilation options
LDFLAGS=-ldflags "-s -w"

# default target
.PHONY: all
all: clean lint test build

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning up build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@$(GO) clean -cache
	@rm -f coverage.out

# Verify dependencies
.PHONY: deps
deps:
	@echo "📦 Verifying and downloading dependencies..."
	@$(GO) mod tidy
	@$(GO) mod verify

# Install development tools
.PHONY: tools
tools:
	@echo "🛠️ Installing development tools..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GO) install github.com/axw/gocov/gocov@latest
	@$(GO) install github.com/matm/gocov-html/cmd/gocov-html@latest

# Linter
.PHONY: lint
lint:
	@echo "🕵️ Running linter..."
	@$(GOLANGCI_LINT) run ./...

# Tests
.PHONY: test
test:
	@echo "🚀 Starting documentation server on http://localhost:8085"
	@$(GO) run ./tests/html_static.go &
	@sleep 1
	@echo "🧪 Running tests..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GO) test -v -covermode=atomic -coverprofile=$(COVERAGE_DIR)/coverage.out ./...

# Generate coverage report
.PHONY: coverage
coverage: test
	@echo "📊 Generating coverage report..."
	@$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out
	@gocov convert $(COVERAGE_DIR)/coverage.out | gocov-html > $(COVERAGE_DIR)/coverage.html
	@echo "Report generated in $(COVERAGE_DIR)/coverage.html"

# Build binary
.PHONY: build
build:
	@echo "🏗️ Building binary..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(SRC_DIR)
	@echo "Binary built in $(BUILD_DIR)/$(BINARY_NAME)"

# Build for all platforms
.PHONY: build-all
build-all:
	@echo "🌍 Building for all platforms..."
	@mkdir -p $(BUILD_DIR)

	# Linux
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(SRC_DIR)
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(SRC_DIR)

	# macOS
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(SRC_DIR)
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(SRC_DIR)

	# Windows
	@GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(SRC_DIR)

	@echo "Binary built in $(BUILD_DIR)"

# Installation
.PHONY: install
install: build
	@echo "📥 Binary installation..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "DeadLinkr installed in /usr/local/bin"

# Uninstallation
.PHONY: uninstall
uninstall:
	@echo "🗑️ Uninstalling..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "DeadLinkr uninstalled"

# Help
.PHONY: help
help:
	@echo "🚀 DeadLinkr - Makefile Commands Available:"
	@echo "  make            - Clean, lint, test and build the project"
	@echo "  make deps       - Check dependencies"
	@echo "  make tools      - Install development tools"
	@echo "  make lint       - Static code analysis"
	@echo "  make test       - Run tests"
	@echo "  make coverage   - Generate test coverage report"
	@echo "  make build      - Build the binary"
	@echo "  make build-all  - Build for all platforms"
	@echo "  make install    - Install the binary"
	@echo "  make uninstall  - Uninstall the binary"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make help       - Show this help message"

# Default
.DEFAULT_GOAL := help