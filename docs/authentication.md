# Authentication Support

Deadlinkr supports multiple authentication methods to scan private websites, intranets, and APIs that require authentication. This document covers all supported authentication methods and how to use them.

## Supported Authentication Methods

### 1. Basic Authentication

Basic HTTP authentication using username and password.

#### Command Line Usage

```bash
# Using command line flag
deadlinkr scan https://private-site.com --auth-basic "username:password"

# Using environment variables
export DEADLINKR_AUTH_USER="username"
export DEADLINKR_AUTH_PASS="password"
deadlinkr scan https://private-site.com
```

#### Examples

```bash
# Scan a private documentation site
deadlinkr scan https://docs.company.com --auth-basic "admin:secret123"

# Scan with specific depth and concurrency
deadlinkr scan https://intranet.company.com \
  --auth-basic "employee:password" \
  --depth 3 \
  --concurrency 10
```

### 2. Bearer Token Authentication

Token-based authentication commonly used with APIs and modern web applications.

#### Command Line Usage

```bash
# Using command line flag
deadlinkr scan https://api.example.com --auth-bearer "your-api-token-here"

# Using environment variable
export DEADLINKR_AUTH_TOKEN="your-api-token-here"
deadlinkr scan https://api.example.com
```

#### Examples

```bash
# Scan an API documentation site
deadlinkr scan https://api-docs.company.com --auth-bearer "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Scan with custom user agent
deadlinkr scan https://protected-site.com \
  --auth-bearer "abc123def456" \
  --user-agent "MyBot/1.0"
```

### 3. Custom Headers Authentication

Flexible authentication method using custom HTTP headers. Useful for API keys, custom authentication schemes, or additional security headers.

#### Command Line Usage

```bash
# Single custom header
deadlinkr scan https://api.example.com --auth-header "X-API-Key: your-api-key"

# Multiple custom headers
deadlinkr scan https://api.example.com \
  --auth-header "X-API-Key: your-api-key" \
  --auth-header "X-Client-Version: 1.0" \
  --auth-header "X-Request-ID: unique-request-id"

# Using environment variable (comma-separated)
export DEADLINKR_AUTH_HEADERS="X-API-Key:secret123,X-Version:2.0"
deadlinkr scan https://api.example.com
```

#### Examples

```bash
# Scan with API key authentication
deadlinkr scan https://docs.api.com --auth-header "X-API-Key: sk_live_1234567890"

# Scan with multiple authentication headers
deadlinkr scan https://internal-api.company.com \
  --auth-header "Authorization: ApiKey your-key-here" \
  --auth-header "X-Tenant-ID: company-123" \
  --auth-header "X-Client-App: deadlinkr"

# Scan with custom content negotiation
deadlinkr scan https://api.example.com \
  --auth-header "X-API-Key: secret" \
  --auth-header "Accept: application/vnd.api+json" \
  --auth-header "Content-Type: application/json"
```

### 4. Cookie-based Authentication

Session-based authentication using HTTP cookies.

#### Command Line Usage

```bash
# Using command line flag
deadlinkr scan https://app.example.com --auth-cookies "sessionid=abc123; csrftoken=xyz789"

# Multiple cookies
deadlinkr scan https://app.example.com --auth-cookies "session=value1; auth_token=value2; preferences=value3"
```

#### Examples

```bash
# Scan authenticated web application
deadlinkr scan https://app.company.com --auth-cookies "JSESSIONID=A1B2C3D4E5F6"

# Scan with complex cookie string
deadlinkr scan https://portal.company.com \
  --auth-cookies "sessionid=1a2b3c4d5e6f; csrftoken=abcdef123456; user_prefs=theme:dark"
```

## Combining Authentication Methods

You can combine multiple authentication methods for complex authentication schemes:

```bash
# Basic auth + custom headers
deadlinkr scan https://api.example.com \
  --auth-basic "user:pass" \
  --auth-header "X-API-Version: v2" \
  --auth-header "X-Client-ID: deadlinkr"

# Bearer token + cookies
deadlinkr scan https://hybrid-auth.com \
  --auth-bearer "jwt-token-here" \
  --auth-cookies "session=abc123"

# All methods combined
deadlinkr scan https://complex-auth.com \
  --auth-basic "user:pass" \
  --auth-bearer "token123" \
  --auth-header "X-API-Key: secret" \
  --auth-cookies "session=xyz789"
```

## Environment Variables

All authentication methods support environment variables as an alternative to command-line flags:

```bash
# Basic Authentication
export DEADLINKR_AUTH_USER="username"
export DEADLINKR_AUTH_PASS="password"

# Bearer Token
export DEADLINKR_AUTH_TOKEN="your-token-here"

# Custom Headers (comma-separated key:value pairs)
export DEADLINKR_AUTH_HEADERS="X-API-Key:secret,X-Version:1.0"

# Now run without explicit auth flags
deadlinkr scan https://private-site.com
```

## Security Best Practices

### 1. Use Environment Variables for Sensitive Data

Instead of passing credentials on the command line (which may be visible in process lists or shell history), use environment variables:

```bash
# ❌ Avoid: credentials visible in command line
deadlinkr scan https://site.com --auth-basic "admin:super-secret-password"

# ✅ Better: use environment variables
export DEADLINKR_AUTH_USER="admin"
export DEADLINKR_AUTH_PASS="super-secret-password"
deadlinkr scan https://site.com
```

### 2. Use Configuration Files

For complex setups, consider using a configuration file (not tracked in version control):

```bash
# Create .env file (add to .gitignore)
echo "DEADLINKR_AUTH_TOKEN=your-secret-token" > .env
echo "DEADLINKR_AUTH_HEADERS=X-API-Key:secret123" >> .env

# Load environment variables
source .env
deadlinkr scan https://api.example.com
```

### 3. Rotate Credentials Regularly

Regularly rotate API keys, tokens, and passwords used for authentication.

### 4. Use Least Privilege

Ensure authentication credentials have only the minimum permissions necessary for link checking.

## Common Use Cases

### Scanning Internal Documentation

```bash
# Company wiki with basic auth
deadlinkr scan https://wiki.company.com \
  --auth-basic "employee:password" \
  --depth 2 \
  --only-internal

# API documentation with token auth
deadlinkr scan https://api-docs.company.com \
  --auth-bearer "$API_TOKEN" \
  --include-pattern "^https://api-docs\.company\.com"
```

### Scanning APIs

```bash
# REST API with API key
deadlinkr scan https://api.service.com/docs \
  --auth-header "X-API-Key: $API_KEY" \
  --auth-header "Accept: application/json"

# GraphQL API with JWT token
deadlinkr scan https://graphql.api.com \
  --auth-bearer "$JWT_TOKEN" \
  --auth-header "Content-Type: application/json"
```

### Scanning Protected Web Applications

```bash
# Web app with session cookies
deadlinkr scan https://app.company.com \
  --auth-cookies "sessionid=$SESSION_ID; csrftoken=$CSRF_TOKEN" \
  --user-agent "DeadLinkr/1.0 (Company Internal Scan)"
```

## Troubleshooting

### Authentication Not Working

1. **Check credentials**: Verify username, password, tokens, and keys are correct
2. **Check logs**: Use `--log-level debug` to see authentication details
3. **Test manually**: Verify credentials work with curl or browser
4. **Check headers**: Ensure custom headers have correct format ("Key: Value")

### Common Issues

#### Invalid Basic Auth Format

```bash
# ❌ Wrong: missing colon
deadlinkr scan https://site.com --auth-basic "userpass"

# ✅ Correct: colon-separated
deadlinkr scan https://site.com --auth-basic "user:pass"
```

#### Invalid Header Format

```bash
# ❌ Wrong: missing colon or space
deadlinkr scan https://site.com --auth-header "X-API-Key=secret"

# ✅ Correct: colon with space
deadlinkr scan https://site.com --auth-header "X-API-Key: secret"
```

### Debug Mode

Use debug mode to see authentication in action:

```bash
deadlinkr scan https://site.com \
  --auth-basic "user:pass" \
  --log-level debug
```

Look for log messages like:
- "Configured basic authentication for user: username"
- "Applied basic auth for user: username"
- "Authentication configured: Basic Auth (user: username)"

## Technical Implementation

The authentication system is built on:

- **HTTPClient Interface**: Supports both regular and authenticated HTTP clients
- **AuthenticatedHTTPClient**: Wraps HTTP requests with authentication
- **Factory Pattern**: Seamlessly integrates authentication across all services
- **Automatic Detection**: Environment variables are automatically detected as fallback
- **Security**: Sensitive data is masked in logs for security

All authentication methods work with all deadlinkr features including:
- Worker pools and concurrency
- Rate limiting
- HEAD request optimization
- Intelligent caching
- Progress tracking
- Result export formats