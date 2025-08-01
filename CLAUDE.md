# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Rules

When working with this repository, follow these rules:
- never commit with your signature
- never add tag comment with your signature

## Commands

### Development Commands
- `make` - Full development cycle: clean, lint, test, and build
- `make build` - Build the binary to `./build/deadlinkr`
- `make test` - Run tests with coverage (starts test server on port 8085)
- `make lint` - Run golangci-lint static analysis
- `make coverage` - Generate and view test coverage report
- `make clean` - Remove build artifacts and clean cache
- `make deps` - Verify and download dependencies
- `make tools` - Install development tools (golangci-lint, gocov, gocov-html)

### Single Commands
- `go build -o build/deadlinkr .` - Quick build
- `go test ./...` - Run all tests
- `go run ./tests/html_static.go` - Start test documentation server

## Architecture

**Deadlinkr** is a Go-based CLI tool for detecting broken links on websites with two main modes:
- `scan` - Recursive website crawling with configurable depth
- `check` - Single page link verification

### Core Structure
- **cmd/** - CLI commands using Cobra framework (root.go, scan.go, check.go)
- **utils/** - HTTP client, URL handling, result processing, and site utilities
- **model/** - Data structures (config.go, struct.go) and global configuration
- **logger/** - Centralized logging with configurable levels
- **internal/** - Service-oriented architecture with optimized components
- **tests/** - Test utilities including static HTML server

### Service Architecture (internal/)
- **LinkChecker Services**: Standard and optimized link checking with HEAD request optimization
- **Worker Pool**: Controlled concurrency with job queues and statistics tracking
- **Rate Limiting**: Token bucket algorithm with domain-specific configurations
- **Caching**: Intelligent caching with TTL strategies based on HTTP response codes
- **Factory Pattern**: Service creation and dependency injection
- **Interfaces**: Clean abstractions for testability and modularity

### Key Components
- **HTTP Client**: Optimized with keep-alive, HEAD requests, and configurable connection pooling
- **Concurrency**: Worker pool architecture with controlled job queuing (configurable workers)
- **Performance Optimizations**: HEAD requests, intelligent caching, domain-based rate limiting
- **Output Formats**: JSON, CSV, and HTML report generation
- **Filtering**: Regex patterns, HTML tag exclusion, internal/external link filtering

### Configuration Flow
CLI flags → Environment variables → YAML config file (deadlinkr.yaml)

## Performance Optimizations

### HEAD Request Optimization
- Automatically detects domain support for HEAD requests
- Falls back to GET when HEAD is not supported (405/501 status codes)
- Tracks HEAD support per domain to avoid unnecessary retries
- Reduces bandwidth usage by 60-80% for supported domains

### Intelligent Caching
- Adaptive TTL based on HTTP response codes:
  - Success responses (2xx): 2x default TTL
  - Not found (404/410): default TTL
  - Server errors (5xx): TTL/4
  - Rate limited (429): TTL/10
- Automatic cache size management with LRU eviction
- Hit rate statistics tracking

### Rate Limiting
- Token bucket algorithm per domain
- Configurable requests per second and burst capacity
- Domain-specific rate limit overrides
- Prevents server bans and respects robots.txt spirit

### Worker Pool Architecture
- Controlled concurrency with job queuing
- Graceful shutdown and resource management
- Real-time statistics (jobs queued/completed/active)
- Memory usage optimization

## CLI Flags Reference

### Core Options
- `--depth <int>` / `-d` - Maximum crawl depth (default: 1)
- `--concurrency <int>` / `-c` - Number of concurrent workers (default: 20)
- `--timeout <int>` / `-t` - Request timeout in seconds (default: 15)
- `--user-agent <string>` - Custom user agent (default: "DeadLinkr/1.0")

### Output Options
- `--output <path>` / `-o` - Export results to file (format auto-detected from extension)
- `--format <type>` / `-f` - Export format (csv/json/html) - overrides auto-detection
- `--show-all` - Show all links including working ones (default: only broken links)
- `--quiet` - Disable output
- `--log-level <level>` - Set log level (debug/info/warn/error/fatal) (default: info)

### Filtering Options
- `--only-internal` - Check only internal links
- `--only-external` - Check only external links
- `--include-pattern <regex>` - Include URLs matching regex pattern
- `--exclude-pattern <regex>` - Exclude URLs matching regex pattern
- `--exclude-html-tags <tags>` - Skip specific HTML tags (comma-separated)

### Performance Options
- `--rate-limit <float>` - Requests per second per domain (default: 2.0)
- `--rate-burst <float>` - Burst capacity for rate limiting (default: 5.0)
- `--optimize-head` - Use HEAD requests when possible (default: true)
- `--cache` - Enable intelligent caching (default: true)
- `--cache-size <int>` - Maximum cache entries (default: 1000)
- `--cache-ttl <int>` - Cache TTL in minutes (default: 60)

## Testing
Tests require a local documentation server that serves static HTML files from `tests/static/`. The `make test` command automatically starts this server on port 8085 before running tests.

### Test Coverage
- Unit tests for all service components
- Integration tests for optimized crawling
- Performance tests for rate limiting and caching
- Mock servers for reliable testing