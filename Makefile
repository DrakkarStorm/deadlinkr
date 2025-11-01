# Makefile for deadlinkr

# configuration variables
BINARY_NAME=deadlinkr
GO=go
GOLANGCI_LINT=golangci-lint

# version information
VERSION=$(shell git describe --tags --always --dirty)
COMMIT=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# directories
SRC_DIR=.
BUILD_DIR=./build
COVERAGE_DIR=./coverage

# compilation options
LDFLAGS=-ldflags "-s -w -X 'github.com/DrakkarStorm/deadlinkr/cmd.Version=$(VERSION)' -X 'github.com/DrakkarStorm/deadlinkr/cmd.Commit=$(COMMIT)' -X 'github.com/DrakkarStorm/deadlinkr/cmd.BuildDate=$(BUILD_DATE)'"

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
	@$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
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


# ==============================================================================
# Docker Commands
# ==============================================================================

# Docker configuration
DOCKER_IMAGE=$(BINARY_NAME)
DOCKER_TAG=latest
DOCKER_REGISTRY=ghcr.io
DOCKER_REPO=drakkarstorm/deadlinkr

# Build Docker image
.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker image..."
	@./scripts/docker-build.sh --tag $(DOCKER_TAG)

# Build and push Docker image
.PHONY: docker-push
docker-push:
	@echo "🐳 Building and pushing Docker image..."
	@./scripts/docker-build.sh --tag $(DOCKER_TAG) --push

# Build multi-architecture Docker image
.PHONY: docker-build-multi
docker-build-multi:
	@echo "🐳 Building multi-architecture Docker image..."
	@./scripts/docker-build.sh --tag $(DOCKER_TAG) --multi-arch --push

# Run Docker container
.PHONY: docker-run
docker-run:
	@echo "🐳 Running Docker container..."
	@docker run --rm $(DOCKER_IMAGE):$(DOCKER_TAG) --help

# Test Docker image
.PHONY: docker-test
docker-test:
	@echo "🧪 Testing Docker image..."
	@docker run --rm $(DOCKER_IMAGE):$(DOCKER_TAG) scan https://httpbin.org/get --timeout 10

# Security scan Docker image
.PHONY: docker-scan
docker-scan:
	@echo "🛡️ Scanning Docker image for vulnerabilities..."
	@./scripts/docker-build.sh --tag $(DOCKER_TAG) --scan


# Clean Docker artifacts
.PHONY: docker-clean
docker-clean:
	@echo "🧹 Cleaning Docker artifacts..."
	@docker image prune -f
	@docker container prune -f
	@docker volume prune -f

# ==============================================================================
# Help
# ==============================================================================
.PHONY: help
help:
	@echo "🚀 DeadLinkr - Makefile Commands Available:"
	@echo ""
	@echo "📦 Build Commands:"
	@echo "  make            - Clean, lint, test and build the project"
	@echo "  make deps       - Check dependencies"
	@echo "  make tools      - Install development tools"
	@echo "  make build      - Build the binary"
	@echo "  make build-all  - Build for all platforms"
	@echo ""
	@echo "🧪 Test Commands:"
	@echo "  make lint       - Static code analysis"
	@echo "  make test       - Run tests"
	@echo "  make coverage   - Generate test coverage report"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-push        - Build and push Docker image"
	@echo "  make docker-build-multi - Build multi-architecture image"
	@echo "  make docker-run         - Run Docker container"
	@echo "  make docker-test        - Test Docker image"
	@echo "  make docker-scan        - Security scan Docker image"
	@echo "  make docker-clean       - Clean Docker artifacts"
	@echo ""
	@echo "🔧 System Commands:"
	@echo "  make install    - Install the binary"
	@echo "  make uninstall  - Uninstall the binary"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "📖 Examples:"
	@echo "  make docker-build DOCKER_TAG=v1.0.0"
	@echo "  make docker-push DOCKER_TAG=latest"
	@echo "  make docker-test DOCKER_TAG=dev"

# Default
.DEFAULT_GOAL := help