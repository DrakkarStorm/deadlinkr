---

# Deadlinkr

Command-line dead link detector for websites.

---

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Introduction](#introduction)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Install via `go install`](#install-via-go-install)
  - [Download Pre-built Binaries](#download-pre-built-binaries)
  - [Homebrew (Linux/Mac)](#homebrew-linuxmac)
  - [Docker](#docker)
- [Configuration](#configuration)
- [Quick Start](#quick-start)
- [CLI Options](#cli-options)
  - [General Parameters](#general-parameters)
  - [Concurrency \& Performance](#concurrency--performance)
  - [Filters \& Depth](#filters--depth)
  - [Output Formats](#output-formats)
  - [Verbosity \& Silent Mode](#verbosity--silent-mode)
- [Usage Examples](#usage-examples)
- [Authentication Support](#authentication-support)
- [CI/CD Integration](#cicd-integration)
  - [GitHub Actions](#github-actions)
  - [GitLab CI/CD](#gitlab-cicd)
- [Exit Codes \& Machine-Friendly Output](#exit-codes--machine-friendly-output)
- [Advanced Configuration](#advanced-configuration)
- [Project Structure](#project-structure)
- [Contributing](#contributing)
- [Roadmap \& Future Enhancements](#roadmap--future-enhancements)

---

## Introduction

Deadlinkr is an open-source command-line tool written in Go designed to automatically detect dead links (HTTP 4xx/5xx or no response) on a website. It offers a **scan** mode (recursive domain crawl) and **check** mode (single page verification), with multiple options for filtering, exporting results, and tuning performance.

**Key Features**: lightweight CLI, advanced performance optimizations (HTTP keep-alive, HEAD requests, intelligent caching, rate limiting), controlled concurrency with worker pools, flexible output formats (JSON, CSV, HTML), and seamless CI/CD integration.

---

## Installation

### Prerequisites

- Go ≥ 1.20 (for installing from source)
- Git (if cloning the repository)

### Install via `go install`

```bash
go install github.com/DrakkarStorm/deadlinkr@latest
```

The `deadlinkr` binary will be installed to `$GOPATH/bin` (or `$(go env GOPATH)/bin`).

### Download Pre-built Binaries

Go to the [Releases](https://github.com/DrakkarStorm/deadlinkr/releases) page and download the archive for your platform (Linux x86\_64, Linux ARM64, macOS x86\_64/ARM64). Unpack and move the binary to your `$PATH`.

### Homebrew (Linux/Mac)

```bash
brew install drakkarstorm/tap/deadlinkr
```

### Docker

Official Docker images are available on multiple registries:

```bash
# GitHub Container Registry (recommended)
docker run --rm ghcr.io/drakkarstorm/deadlinkr:latest scan https://example.com --depth 2

# Docker Hub (mirror)
docker run --rm deadlinkr/deadlinkr:latest scan https://example.com --depth 2

# With authentication
docker run --rm ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://private-site.com \
  --auth-basic "user:pass" \
  --depth 2

# Mount volumes for reports
docker run --rm \
  -v $(pwd)/reports:/app/reports \
  ghcr.io/drakkarstorm/deadlinkr:latest \
  scan https://example.com \
  --format json \
  --output /app/reports/report.json
```

**Available Images:**
- `ghcr.io/drakkarstorm/deadlinkr:latest` (multi-arch: linux/amd64, linux/arm64)
- `deadlinkr/deadlinkr:latest` (Docker Hub mirror)

For detailed Docker usage, see [docs/docker.md](docs/docker.md).

---

## Configuration

Deadlinkr can be configured via:

- **CLI Flags** (`--concurrency`, `--timeout`, etc.)
- **Environment Variables**:
  - `DEADLINKR_CONCURRENCY` (default number of concurrent HTTP requests)
  - `DEADLINKR_TIMEOUT` (global HTTP timeout in seconds)

Flag ↔ environment variable mapping is handled automatically using Viper + Cobra.

---

## Quick Start

```bash
# Full site scan with depth level 2
deadlinkr scan https://example.com --depth 2

# Check a single page
deadlinkr check https://example.com/page.html
```

---

## CLI Options

### Core Commands

| Command | Description |
| ------- | ----------- |
| `scan [url]` | Recursively scan a website for broken links |
| `check [url]` | Check links on a single page only |

### General Parameters

| Option                  | Alias | Description                                            | Default        |
| ----------------------- | ----- | ------------------------------------------------------ | -------------- |
| `--help`                | `-h`  | Show help message                                      |                |
| `--version`             |       | Display the tool version                               |                |
| `--timeout <s>`         | `-t`  | Global HTTP request timeout in seconds                 | 15             |
| `--user-agent <string>` |       | User-Agent header for requests                         | DeadLinkr/1.0  |

### Concurrency & Performance

| Option                      | Alias | Description                                                       | Default |
| --------------------------- | ----- | ----------------------------------------------------------------- | ------- |
| `--concurrency <N>`         | `-c`  | Maximum number of simultaneous HTTP requests                      | 20      |
| `--rate-limit <float>`      |       | Requests per second per domain (prevents server bans)            | 2.0     |
| `--rate-burst <float>`      |       | Burst capacity for rate limiting                                  | 5.0     |
| `--optimize-head`           |       | Use HEAD requests when possible to reduce bandwidth               | true    |
| `--cache`                   |       | Enable intelligent caching of link check results                 | true    |
| `--cache-size <int>`        |       | Maximum number of entries in the cache                           | 1000    |
| `--cache-ttl <int>`         |       | Cache time-to-live in minutes                                    | 60      |

> **Performance Tips**: 
> - HEAD requests can reduce bandwidth by 60-80%
> - Intelligent caching can improve performance by 50-90% on repeated scans
> - Rate limiting prevents server bans and respects website resources
> - Worker pools provide controlled concurrency without memory explosion

### Crawling & Filtering

| Option                      | Alias | Description                                                   | Default |
| --------------------------- | ----- | ------------------------------------------------------------- | ------- |
| `--depth <n>`               | `-d`  | Crawl depth (levels of internal links to follow)              | 1       |
| `--only-internal`           |       | Only check links within the same domain as the base URL       | false   |
| `--only-external`           |       | Only check external links                                     | false   |
| `--include-pattern <regex>` |       | Only include URLs matching the regex                          | —       |
| `--exclude-pattern <regex>` |       | Exclude URLs matching the regex                               | —       |
| `--exclude-html-tags <css>` |       | CSS selector for HTML tags to ignore (e.g., `nav`, `.footer`) | —       |

### Output & Display

| Option                | Alias | Description                                                         | Default |
| --------------------- | ----- | ------------------------------------------------------------------- | ------- |
| `--output <file>`     | `-o`  | Output file path (format auto-detected from extension)             | —       |
| `--format <type>`     | `-f`  | Export format (csv, json, html) - overrides auto-detection         | —       |
| `--show-all`          |       | Show all links including working ones (default: only broken links) | false   |
| `--quiet`             |       | Show only summary (scanned links count and dead links count)       | false   |
| `--log-level <level>` |       | Log level (debug, info, warn, error, fatal)                        | info    |

### Authentication Options

| Option                          | Description                                                         | Default |
| ------------------------------- | ------------------------------------------------------------------- | ------- |
| `--auth-basic <user:pass>`      | Basic authentication (or use DEADLINKR_AUTH_USER/DEADLINKR_AUTH_PASS env vars) | —       |
| `--auth-bearer <token>`         | Bearer token authentication (or use DEADLINKR_AUTH_TOKEN env var)  | —       |
| `--auth-header <"Key: Value">`  | Custom authentication headers (can be used multiple times)         | —       |
| `--auth-cookies <cookie_string>`| Cookie-based authentication                                        | —       |

---

## Usage Examples

### Basic Usage

```bash
# Quick scan with auto-detected JSON format
deadlinkr scan https://example.com -o report.json

# Check a single page with CSV output
deadlinkr check https://example.com/page.html -o results.csv

# Deep scan with custom settings using shortcuts
deadlinkr scan https://example.com -d 3 -c 50 -t 20 -o deep-scan.html
```

### Advanced Usage

```bash
# High-performance scan with optimizations
deadlinkr scan https://example.com \
  -d 2 -c 100 \
  --rate-limit 5.0 --rate-burst 10.0 \
  --cache-size 2000 \
  -o optimized-report.json

# Show all links (including working ones)
deadlinkr scan https://example.com --show-all -o full-report.html

# Exclude social media links and disable caching
deadlinkr scan https://example.com \
  --exclude-pattern ".*(facebook|twitter|linkedin)\.com.*" \
  --cache=false \
  -o filtered-links.json

# Conservative scan for fragile websites
deadlinkr scan https://fragile-site.com \
  --rate-limit 0.5 --rate-burst 1.0 \
  -c 5 --optimize-head=false

# Export format override (ignores file extension)
deadlinkr scan https://example.com -o report.txt -f json
```

### Format Auto-Detection

```bash
# Format automatically detected from file extension
deadlinkr scan https://example.com -o report.json    # → JSON format
deadlinkr scan https://example.com -o report.csv     # → CSV format  
deadlinkr scan https://example.com -o report.html    # → HTML format

# Manual format override
deadlinkr scan https://example.com -o data.txt -f csv  # → CSV in .txt file
```

---

## Authentication Support

Deadlinkr supports multiple authentication methods to scan private websites, intranets, and APIs:

```bash
# Basic authentication
deadlinkr scan https://private-docs.company.com \
  --auth-basic "username:password" --depth 2

# Bearer token authentication  
deadlinkr scan https://api-docs.service.com \
  --auth-bearer "eyJhbGciOiJIUzI1NiIs..." --depth 1

# API key authentication using custom headers
deadlinkr scan https://api.example.com \
  --auth-header "X-API-Key: secret123" \
  --auth-header "X-Client-Version: 1.0"

# Session-based authentication with cookies
deadlinkr scan https://internal-app.company.com \
  --auth-cookies "sessionid=abc123; csrftoken=xyz789"

# Multiple authentication methods combined
deadlinkr scan https://complex-auth.com \
  --auth-basic "user:pass" \
  --auth-header "X-API-Key: secret" \
  --auth-cookies "session=xyz"

# Using environment variables for security
export DEADLINKR_AUTH_USER="username"
export DEADLINKR_AUTH_PASS="password"
deadlinkr scan https://private-site.com
```

**Supported Authentication Methods:**
- **Basic Auth**: Username/password authentication (`--auth-basic "user:pass"`)
- **Bearer Tokens**: JWT or API tokens (`--auth-bearer "token"`)  
- **Custom Headers**: API keys and custom auth (`--auth-header "Key: Value"`)
- **Cookies**: Session-based auth (`--auth-cookies "session=value"`)

For detailed documentation, see [docs/authentication.md](docs/authentication.md).

---

## CI/CD Integration

### GitHub Actions

```yaml
name: Dead Link Check
on: [push]
jobs:
  deadlinkr:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Install Deadlinkr
        run: go install github.com/DrakkarStorm/deadlinkr@latest
      - name: Run Deadlinkr
        run: |
          deadlinkr scan https://example.com \
            --depth 2 --concurrency 50 \
            --format json --output deadlinkr-report.json || exit 1
      - name: Upload Report
        uses: actions/upload-artifact@v3
        with:
          name: deadlinkr-report
          path: deadlinkr-report.json
```

### GitLab CI/CD

```yaml
stages:
  - test

link_check:
  stage: test
  image: golang:1.20
  script:
    - go install github.com/DrakkarStorm/deadlinkr@latest
    - deadlinkr scan https://example.com --depth 2 --format json --output report.json || exit 1
  artifacts:
    paths:
      - report.json
```

---

## Exit Codes & Machine-Friendly Output

- **0**: No dead links detected
- **1**: At least one dead link detected
- **>1**: Execution error (timeout, parsing, etc.)

JSON format is recommended for automated parsing (e.g., `jq .`), while HTML is suitable for human review.

---

## Advanced Configuration

Deadlinkr also supports a `deadlinkr.yaml` configuration file located in:

1. The current working directory
2. `$HOME/.config/deadlinkr/`

```yaml
# Core settings
concurrency: 100
timeout: 10
format: json
output: report.json

# Performance optimizations
rate_limit: 3.0
rate_burst: 8.0
optimize_head: true
cache: true
cache_size: 1500
cache_ttl: 90

# Filtering
exclude_html_tags:
  - "nav"
  - ".footer"
exclude_pattern: ".*(facebook|twitter|linkedin)\\.com.*"
```

CLI flags override environment variables, which in turn override configuration file settings.

---

## Project Structure

```
deadlinkr/
├─ cmd/                # CLI entry points (scan, check commands)
├─ internal/           # Service-oriented architecture
│  ├─ interfaces.go    # Service interfaces and contracts
│  ├─ factory.go       # Service creation and dependency injection
│  ├─ workerpool.go    # Worker pool with controlled concurrency
│  ├─ ratelimiter.go   # Token bucket rate limiting per domain
│  ├─ cache.go         # Intelligent caching with adaptive TTL
│  ├─ linkchecker_*.go # Standard and optimized link checkers
│  └─ *_test.go        # Comprehensive test coverage
├─ utils/              # HTTP client, URL handling, adapters
├─ model/              # Data structures and global configuration
├─ logger/             # Centralized logging with levels
├─ tests/              # Test utilities and static HTML server
├─ go.mod, go.sum      # Dependency management
├─ README.md           # Main documentation
├─ CLAUDE.md           # Development guide for Claude Code
└─ .goreleaser.yml     # Automated release configuration
```

---

## Contributing

Contributions are welcome! Please:

1. **Fork** the repository
2. Create a **branch** (`feat/`, `fix/`, `doc/`)
3. **Commit** using [Conventional Commits] style
4. Run **tests**: `go test ./...`
5. Open a **Pull Request** describing your changes

See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.

---

## Roadmap & Future Enhancements

### Performance & Scalability ✅ 
- ✅ **HEAD Request Optimization**: bandwidth reduction (60-80%)
- ✅ **Intelligent Caching**: adaptive TTL strategies 
- ✅ **Rate Limiting**: token bucket algorithm per domain
- ✅ **Worker Pools**: controlled concurrency with job queues

### Authentication Support ✅ 
- ✅ **Basic Authentication**: username/password authentication
- ✅ **Bearer Token Authentication**: JWT and API token support
- ✅ **Custom Headers Authentication**: API keys and custom auth schemes
- ✅ **Cookie-based Authentication**: session-based authentication
- ✅ **Environment Variable Support**: secure credential management

### Docker & DevOps ✅ 
- ✅ **Multi-Architecture Docker Images**: linux/amd64, linux/arm64
- ✅ **Optimized Distroless Images**: < 20MB, security-hardened
- ✅ **Docker Compose Support**: development and production setups
- ✅ **CI/CD Pipeline**: automated build, test, security scan, and publish
- ✅ **Multi-Registry Publishing**: GitHub Container Registry + Docker Hub

### Future Features
- 📦 **Fix Mode**: automatically correct broken internal links
- 🌐 **REST API**: remote control and SaaS integration
- 📊 **Web Dashboard**: real-time report visualization with performance metrics
- 📈 **Monitoring Integration**: Prometheus metrics and health checks

---

