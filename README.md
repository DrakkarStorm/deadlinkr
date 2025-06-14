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

**Key Features**: lightweight CLI, optimized performance (HTTP keep-alive, HEAD requests), controlled concurrency, flexible output formats (JSON, CSV, HTML), and seamless CI/CD integration.

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

A Docker image is also available:

```bash
docker run --rm drakkarstorm/deadlinkr scan https://example.com --depth 2
```

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

### General Parameters

| Option                  | Alias | Description                                            | Default        |
| ----------------------- | ----- | ------------------------------------------------------ | -------------- |
| `--help`                | `-h`  | Show help message                                      |                |
| `--version`             |       | Display the tool version                               |                |
| `--timeout <s>`         | `-t`  | Global HTTP request timeout in seconds                 | 15             |
| `--user-agent <string>` |       | User-Agent header for requests (e.g., "Deadlinkr/1.0") | Go-http-client |

### Concurrency & Performance

| Option                      | Alias | Description                                                       | Default |
| --------------------------- | ----- | ----------------------------------------------------------------- | ------- |
| `--concurrency <N>`         | `-c`  | Maximum number of simultaneous HTTP requests                      | 20      |
| `--max-idle-conns <N>`      |       | `Transport.MaxIdleConns` (total persistent TCP connections)       | 2       |
| `--max-idle-conns-per-host` |       | `Transport.MaxIdleConnsPerHost` (persistent connections per host) | 2       |
| `--idle-conn-timeout <s>`   |       | `Transport.IdleConnTimeout` in seconds                            | 90      |

> **Recommendation**: To process 10,000 links with 200 workers, use `--concurrency 200 --max-idle-conns-per-host 200 --max-idle-conns 400` for optimal throughput while controlling resource usage.

### Filters & Depth

| Option                      | Alias | Description                                                   | Default |
| --------------------------- | ----- | ------------------------------------------------------------- | ------- |
| `--depth <n>`               | `-d`  | Crawl depth (levels of internal links to follow)              | 1       |
| `--only-internal`           |       | Only check links within the same domain as the base URL       | false   |
| `--only-external`           |       | Only check external links                                     | false   |
| `--include-pattern <regex>` |       | Only include URLs matching the regex                          | —       |
| `--exclude-pattern <regex>` |       | Exclude URLs matching the regex                               | —       |
| `--exclude-html-tags <css>` |       | CSS selector for HTML tags to ignore (e.g., `nav`, `.footer`) | —       |

### Output Formats

| Option            | Description                        | Supported Formats     |
| ----------------- | ---------------------------------- | --------------------- |
| `--format <type>` | Select the report format           | `json`, `csv`, `html` |
| `--output <file>` | Output file (overwrites if exists) | —                     |

### Verbosity & Silent Mode

| Option       | Description                                                  |
| ------------ | ------------------------------------------------------------ |
| `--quiet`    | Show only summary (scanned links count and dead links count) |
| `--verbose`  | Enable detailed logging (debug level)                        |
| `--no-color` | Disable ANSI color output                                    |

---

## Usage Examples

```bash
# Quick scan (depth=1), 50 workers, 10s timeout, JSON output
deadlinkr scan https://example.com \
  --depth 1 --concurrency 50 --timeout 10 \
  --format json --output report.json

# Check a page and generate a CSV report
deadlinkr check https://example.com/index.html \
  --format csv > dead_links.csv

# Exclude social media links
deadlinkr scan https://example.com \
  --exclude-pattern ".*(facebook|twitter|linkedin)\.com.*" \
  --format html --output links.html
```

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
concurrency: 100
timeout: 10
format: json
output: report.json
exclude_html_tags:
  - "nav"
  - ".footer"
```

CLI flags override environment variables, which in turn override configuration file settings.

---

## Project Structure

```
deadlinkr/
├─ cmd/                # CLI entry points (e.g., scan, check)
├─ internal/           # Internal code (crawler, fetcher, parser)
├─ pkg/                # Public packages (modularity, tests)
├─ utils/              # Utility functions (HTTP client, patterns)
├─ model/              # Data structures (LinkResult, Stats)
├─ logger/             # Logging configuration and wrapper
├─ go.mod, go.sum      # Dependency definitions
├─ README.md           # Main documentation
├─ CONTRIBUTING.md     # Contribution guide
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

- 📦 **Fix Mode**: automatically correct broken internal links
- 🔒 **Authentication Support**: scan private sites (Basic, OAuth)
- 🌐 **REST API**: remote control and SaaS integration
- 📊 **Web Dashboard**: real-time report visualization
- 🐳 **Official Docker Image**

---

