package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

func init() {
	// Initialize logger for tests
	logger.InitLogger("error") // Use error level to reduce test output
}

func TestAuthConfig(t *testing.T) {
	t.Run("NewAuthConfig creates empty config", func(t *testing.T) {
		config := NewAuthConfig()
		
		if config == nil {
			t.Fatal("Expected non-nil config")
		}
		
		if config.BasicEnabled || config.BearerEnabled || config.HeadersEnabled || config.CookiesEnabled {
			t.Error("Expected all auth methods to be disabled by default")
		}
		
		if config.CustomHeaders == nil {
			t.Error("Expected CustomHeaders map to be initialized")
		}
	})
}

func TestAuthenticatedHTTPClient(t *testing.T) {
	t.Run("Basic Authentication", func(t *testing.T) {
		// Create test server that requires basic auth
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Basic ") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		}))
		defer server.Close()

		// Create authenticated client
		config := NewAuthConfig()
		config.BasicUser = "testuser"
		config.BasicPassword = "testpass"
		config.BasicEnabled = true
		
		client := NewAuthenticatedHTTPClient(&http.Client{}, config)
		
		// Make request
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Bearer Token Authentication", func(t *testing.T) {
		expectedToken := "test-bearer-token-123"
		
		// Create test server that checks bearer token
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			expected := "Bearer " + expectedToken
			if auth != expected {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		}))
		defer server.Close()

		// Create authenticated client
		config := NewAuthConfig()
		config.BearerToken = expectedToken
		config.BearerEnabled = true
		
		client := NewAuthenticatedHTTPClient(&http.Client{}, config)
		
		// Make request
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Custom Headers Authentication", func(t *testing.T) {
		// Create test server that checks custom headers
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")
			clientVersion := r.Header.Get("X-Client-Version")
			
			if apiKey != "secret123" || clientVersion != "1.0" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		}))
		defer server.Close()

		// Create authenticated client
		config := NewAuthConfig()
		config.CustomHeaders = map[string]string{
			"X-API-Key":        "secret123",
			"X-Client-Version": "1.0",
		}
		config.HeadersEnabled = true
		
		client := NewAuthenticatedHTTPClient(&http.Client{}, config)
		
		// Make request
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Cookie Authentication", func(t *testing.T) {
		expectedCookies := "sessionid=abc123; csrftoken=xyz789"
		
		// Create test server that checks cookies
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie := r.Header.Get("Cookie")
			if cookie != expectedCookies {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		}))
		defer server.Close()

		// Create authenticated client
		config := NewAuthConfig()
		config.Cookies = expectedCookies
		config.CookiesEnabled = true
		
		client := NewAuthenticatedHTTPClient(&http.Client{}, config)
		
		// Make request
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("Multiple Authentication Methods", func(t *testing.T) {
		// Create test server that checks multiple auth methods
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			apiKey := r.Header.Get("X-API-Key")
			
			if !strings.HasPrefix(auth, "Basic ") || apiKey != "secret123" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated"))
		}))
		defer server.Close()

		// Create authenticated client with multiple methods
		config := NewAuthConfig()
		config.BasicUser = "user"
		config.BasicPassword = "pass"
		config.BasicEnabled = true
		config.CustomHeaders = map[string]string{"X-API-Key": "secret123"}
		config.HeadersEnabled = true
		
		client := NewAuthenticatedHTTPClient(&http.Client{}, config)
		
		// Make request
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})
}

func TestAuthenticationParsing(t *testing.T) {
	t.Run("ParseBasicAuthFromString", func(t *testing.T) {
		tests := []struct {
			input    string
			wantUser string
			wantPass string
			wantErr  bool
		}{
			{"user:password", "user", "password", false},
			{"admin:secret123", "admin", "secret123", false},
			{"", "", "", true},
			{"invalid", "", "", true},
			{"user:", "user", "", false},
			{":password", "", "password", false},
		}

		for _, tt := range tests {
			user, pass, err := ParseBasicAuthFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBasicAuthFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				continue
			}
			if user != tt.wantUser || pass != tt.wantPass {
				t.Errorf("ParseBasicAuthFromString(%q) = (%q, %q), want (%q, %q)", tt.input, user, pass, tt.wantUser, tt.wantPass)
			}
		}
	})

	t.Run("ParseCustomHeaderFromString", func(t *testing.T) {
		tests := []struct {
			input     string
			wantKey   string
			wantValue string
			wantErr   bool
		}{
			{"X-API-Key: secret123", "X-API-Key", "secret123", false},
			{"Authorization: Bearer token", "Authorization", "Bearer token", false},
			{"Content-Type:application/json", "Content-Type", "application/json", false},
			{"", "", "", true},
			{"invalid", "", "", true},
			{": value", "", "", true},
			{"Key:", "Key", "", false},
		}

		for _, tt := range tests {
			key, value, err := ParseCustomHeaderFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCustomHeaderFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				continue
			}
			if key != tt.wantKey || value != tt.wantValue {
				t.Errorf("ParseCustomHeaderFromString(%q) = (%q, %q), want (%q, %q)", tt.input, key, value, tt.wantKey, tt.wantValue)
			}
		}
	})
}

func TestAuthenticationMethods(t *testing.T) {
	client := NewAuthenticatedHTTPClient(&http.Client{}, NewAuthConfig())
	
	t.Run("SetBasicAuth", func(t *testing.T) {
		client.SetBasicAuth("testuser", "testpass")
		
		if !client.config.BasicEnabled {
			t.Error("Expected BasicEnabled to be true")
		}
		if client.config.BasicUser != "testuser" || client.config.BasicPassword != "testpass" {
			t.Error("Basic auth credentials not set correctly")
		}
	})

	t.Run("SetBearerToken", func(t *testing.T) {
		client.SetBearerToken("test-token")
		
		if !client.config.BearerEnabled {
			t.Error("Expected BearerEnabled to be true")
		}
		if client.config.BearerToken != "test-token" {
			t.Error("Bearer token not set correctly")
		}
	})

	t.Run("AddCustomHeader", func(t *testing.T) {
		client.AddCustomHeader("X-API-Key", "secret123")
		
		if !client.config.HeadersEnabled {
			t.Error("Expected HeadersEnabled to be true")
		}
		if client.config.CustomHeaders["X-API-Key"] != "secret123" {
			t.Error("Custom header not set correctly")
		}
	})

	t.Run("SetCookies", func(t *testing.T) {
		client.SetCookies("sessionid=abc123")
		
		if !client.config.CookiesEnabled {
			t.Error("Expected CookiesEnabled to be true")
		}
		if client.config.Cookies != "sessionid=abc123" {
			t.Error("Cookies not set correctly")
		}
	})
}

func TestAuthSummary(t *testing.T) {
	t.Run("No authentication", func(t *testing.T) {
		client := NewAuthenticatedHTTPClient(&http.Client{}, NewAuthConfig())
		summary := client.GetAuthSummary()
		
		if summary != "No authentication configured" {
			t.Errorf("Expected 'No authentication configured', got %q", summary)
		}
	})

	t.Run("All authentication methods", func(t *testing.T) {
		config := NewAuthConfig()
		config.BasicUser = "user"
		config.BasicPassword = "pass"
		config.BasicEnabled = true
		config.BearerToken = "very-long-token-that-should-be-truncated"
		config.BearerEnabled = true
		config.CustomHeaders = map[string]string{"X-Key": "value"}
		config.HeadersEnabled = true
		config.Cookies = "session=123"
		config.CookiesEnabled = true
		
		client := NewAuthenticatedHTTPClient(&http.Client{}, config)
		summary := client.GetAuthSummary()
		
		if !strings.Contains(summary, "Basic Auth (user: user)") {
			t.Error("Summary should contain basic auth info")
		}
		if !strings.Contains(summary, "Bearer Token (very-long-...") {
			t.Error("Summary should contain truncated bearer token")
		}
		if !strings.Contains(summary, "Custom Headers (1)") {
			t.Error("Summary should contain custom headers count")
		}
		if !strings.Contains(summary, "Cookies") {
			t.Error("Summary should contain cookies")
		}
	})
}

// Test integration with factory
func TestFactoryAuthIntegration(t *testing.T) {
	// Save original model values
	origBasic := model.AuthBasic
	origBearer := model.AuthBearer
	origHeaders := model.AuthHeaders
	origCookies := model.AuthCookies
	
	// Restore after test
	defer func() {
		model.AuthBasic = origBasic
		model.AuthBearer = origBearer
		model.AuthHeaders = origHeaders
		model.AuthCookies = origCookies
	}()

	t.Run("Factory creates authenticated client from model config", func(t *testing.T) {
		// Set up model config
		model.AuthBasic = "user:pass"
		model.AuthBearer = "token123"
		model.AuthHeaders = []string{"X-API-Key: secret", "X-Version: 1.0"}
		model.AuthCookies = "session=abc123"
		
		factory := NewServiceFactory()
		authClient := factory.createAuthenticatedClient(&http.Client{})
		
		if authClient == nil {
			t.Fatal("Expected non-nil authenticated client")
		}
		
		// Verify config was applied
		summary := authClient.GetAuthSummary()
		if !strings.Contains(summary, "Basic Auth") {
			t.Error("Expected basic auth to be configured")
		}
		if !strings.Contains(summary, "Bearer Token") {
			t.Error("Expected bearer token to be configured")
		}
		if !strings.Contains(summary, "Custom Headers (2)") {
			t.Error("Expected 2 custom headers to be configured")
		}
		if !strings.Contains(summary, "Cookies") {
			t.Error("Expected cookies to be configured")
		}
	})
}