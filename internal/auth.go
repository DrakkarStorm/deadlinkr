package internal

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/DrakkarStorm/deadlinkr/logger"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	// Basic Auth
	BasicUser     string
	BasicPassword string
	BasicEnabled  bool
	
	// Bearer Token
	BearerToken   string
	BearerEnabled bool
	
	// Custom Headers
	CustomHeaders map[string]string
	HeadersEnabled bool
	
	// Cookies
	Cookies        string
	CookiesEnabled bool
}

// NewAuthConfig creates a new authentication configuration
func NewAuthConfig() *AuthConfig {
	return &AuthConfig{
		CustomHeaders: make(map[string]string),
	}
}

// AuthenticatedHTTPClient wraps an HTTP client with authentication capabilities
type AuthenticatedHTTPClient struct {
	client *http.Client
	config *AuthConfig
}

// NewAuthenticatedHTTPClient creates a new authenticated HTTP client
func NewAuthenticatedHTTPClient(client *http.Client, config *AuthConfig) *AuthenticatedHTTPClient {
	if config == nil {
		config = NewAuthConfig()
	}
	
	return &AuthenticatedHTTPClient{
		client: client,
		config: config,
	}
}

// Do executes an HTTP request with authentication
func (ac *AuthenticatedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Apply authentication to the request
	if err := ac.applyAuthentication(req); err != nil {
		return nil, fmt.Errorf("failed to apply authentication: %w", err)
	}
	
	// Execute the request
	return ac.client.Do(req)
}

// applyAuthentication applies configured authentication to the request
func (ac *AuthenticatedHTTPClient) applyAuthentication(req *http.Request) error {
	// Apply Basic Authentication
	if ac.config.BasicEnabled {
		if err := ac.applyBasicAuth(req); err != nil {
			return fmt.Errorf("basic auth error: %w", err)
		}
	}
	
	// Apply Bearer Token
	if ac.config.BearerEnabled {
		if err := ac.applyBearerToken(req); err != nil {
			return fmt.Errorf("bearer token error: %w", err)
		}
	}
	
	// Apply Custom Headers
	if ac.config.HeadersEnabled {
		ac.applyCustomHeaders(req)
	}
	
	// Apply Cookies
	if ac.config.CookiesEnabled {
		if err := ac.applyCookies(req); err != nil {
			return fmt.Errorf("cookies error: %w", err)
		}
	}
	
	return nil
}

// applyBasicAuth applies Basic Authentication to the request
func (ac *AuthenticatedHTTPClient) applyBasicAuth(req *http.Request) error {
	if ac.config.BasicUser == "" || ac.config.BasicPassword == "" {
		return fmt.Errorf("basic auth enabled but username or password is empty")
	}
	
	// Create basic auth header
	auth := ac.config.BasicUser + ":" + ac.config.BasicPassword
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encodedAuth)
	
	logger.Debugf("Applied basic auth for user: %s", ac.config.BasicUser)
	return nil
}

// applyBearerToken applies Bearer Token authentication to the request
func (ac *AuthenticatedHTTPClient) applyBearerToken(req *http.Request) error {
	if ac.config.BearerToken == "" {
		return fmt.Errorf("bearer token enabled but token is empty")
	}
	
	req.Header.Set("Authorization", "Bearer "+ac.config.BearerToken)
	
	logger.Debugf("Applied bearer token authentication (token length: %d)", len(ac.config.BearerToken))
	return nil
}

// applyCustomHeaders applies custom headers to the request
func (ac *AuthenticatedHTTPClient) applyCustomHeaders(req *http.Request) {
	for key, value := range ac.config.CustomHeaders {
		req.Header.Set(key, value)
		logger.Debugf("Applied custom header: %s", key)
	}
}

// applyCookies applies cookies to the request
func (ac *AuthenticatedHTTPClient) applyCookies(req *http.Request) error {
	if ac.config.Cookies == "" {
		return fmt.Errorf("cookies enabled but cookies string is empty")
	}
	
	req.Header.Set("Cookie", ac.config.Cookies)
	
	logger.Debugf("Applied cookies authentication")
	return nil
}

// SetBasicAuth configures Basic Authentication
func (ac *AuthenticatedHTTPClient) SetBasicAuth(username, password string) {
	ac.config.BasicUser = username
	ac.config.BasicPassword = password
	ac.config.BasicEnabled = true
	
	logger.Debugf("Configured basic auth for user: %s", username)
}

// SetBearerToken configures Bearer Token authentication
func (ac *AuthenticatedHTTPClient) SetBearerToken(token string) {
	ac.config.BearerToken = token
	ac.config.BearerEnabled = true
	
	logger.Debugf("Configured bearer token (length: %d)", len(token))
}

// AddCustomHeader adds a custom header for authentication
func (ac *AuthenticatedHTTPClient) AddCustomHeader(key, value string) {
	if ac.config.CustomHeaders == nil {
		ac.config.CustomHeaders = make(map[string]string)
	}
	
	ac.config.CustomHeaders[key] = value
	ac.config.HeadersEnabled = true
	
	logger.Debugf("Added custom header: %s", key)
}

// SetCookies configures cookie-based authentication
func (ac *AuthenticatedHTTPClient) SetCookies(cookies string) {
	ac.config.Cookies = cookies
	ac.config.CookiesEnabled = true
	
	logger.Debugf("Configured cookies authentication")
}

// ParseBasicAuthFromEnv parses basic auth from environment variables
func ParseBasicAuthFromEnv() (string, string, bool) {
	user := os.Getenv("DEADLINKR_AUTH_USER")
	pass := os.Getenv("DEADLINKR_AUTH_PASS")
	
	if user != "" && pass != "" {
		return user, pass, true
	}
	
	return "", "", false
}

// ParseBearerTokenFromEnv parses bearer token from environment variable
func ParseBearerTokenFromEnv() (string, bool) {
	token := os.Getenv("DEADLINKR_AUTH_TOKEN")
	return token, token != ""
}

// ParseCustomHeadersFromEnv parses custom headers from environment variable
func ParseCustomHeadersFromEnv() map[string]string {
	headers := make(map[string]string)
	
	headerStr := os.Getenv("DEADLINKR_AUTH_HEADERS")
	if headerStr == "" {
		return headers
	}
	
	// Parse format: "Key1:Value1,Key2:Value2"
	pairs := strings.Split(headerStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}
	
	return headers
}

// ParseBasicAuthFromString parses "user:password" format
func ParseBasicAuthFromString(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", fmt.Errorf("empty auth string")
	}
	
	parts := strings.SplitN(authStr, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid basic auth format, expected 'user:password'")
	}
	
	return parts[0], parts[1], nil
}

// ParseCustomHeaderFromString parses "Key: Value" format
func ParseCustomHeaderFromString(headerStr string) (string, string, error) {
	if headerStr == "" {
		return "", "", fmt.Errorf("empty header string")
	}
	
	parts := strings.SplitN(headerStr, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid header format, expected 'Key: Value'")
	}
	
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	
	if key == "" {
		return "", "", fmt.Errorf("header key cannot be empty")
	}
	
	return key, value, nil
}

// GetAuthSummary returns a summary of configured authentication methods
func (ac *AuthenticatedHTTPClient) GetAuthSummary() string {
	var methods []string
	
	if ac.config.BasicEnabled {
		methods = append(methods, fmt.Sprintf("Basic Auth (user: %s)", ac.config.BasicUser))
	}
	
	if ac.config.BearerEnabled {
		tokenPreview := ac.config.BearerToken
		if len(tokenPreview) > 10 {
			tokenPreview = tokenPreview[:10] + "..."
		}
		methods = append(methods, fmt.Sprintf("Bearer Token (%s)", tokenPreview))
	}
	
	if ac.config.HeadersEnabled {
		headerCount := len(ac.config.CustomHeaders)
		methods = append(methods, fmt.Sprintf("Custom Headers (%d)", headerCount))
	}
	
	if ac.config.CookiesEnabled {
		methods = append(methods, "Cookies")
	}
	
	if len(methods) == 0 {
		return "No authentication configured"
	}
	
	return strings.Join(methods, ", ")
}