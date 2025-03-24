package utils

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/stretchr/testify/assert"
)

// TestResolveURL tests the resolveURL function
func TestResolveURL(t *testing.T) {
	testCases := []struct {
		name        string
		baseURL     string
		pageURL     string
		href        string
		expectedURL string
		expectError bool
	}{
		{
			name:        "Absolute URL",
			baseURL:     "http://127.0.0.1:8085",
			pageURL:     "http://127.0.0.1:8085/installation.html",
			href:        "https://github.com/",
			expectedURL: "https://github.com/",
			expectError: false,
		},
		{
			name:        "Relative URL",
			baseURL:     "http://127.0.0.1:8085",
			pageURL:     "http://127.0.0.1:8085/installation.html",
			href:        "tutoriel.html",
			expectedURL: "http://127.0.0.1:8085/tutoriel.html",
			expectError: false,
		},
		{
			name:        "Root-relative URL",
			baseURL:     "http://127.0.0.1:8085",
			pageURL:     "http://127.0.0.1:8085/tutoriel.html",
			href:        "/index.html",
			expectedURL: "http://127.0.0.1:8085/index.html",
			expectError: false,
		},
		{
			name:        "Invalid href",
			baseURL:     "http://127.0.0.1:8085",
			pageURL:     "http://127.0.0.1:8085/page",
			href:        ":%invalid",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resolvedURL, err := ResolveURL(tc.pageURL, tc.href)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedURL, resolvedURL.String())
			}
		})
	}
}

// TestShouldSkipURL tests the shouldSkipURL function
func TestShouldSkipURL(t *testing.T) {
	testCases := []struct {
		name     string
		baseURL  string
		linkURL  string
		expected bool
	}{
		{
			name:     "HTTP URL",
			baseURL:  "http://127.0.0.1:8085",
			linkURL:  "http://other.com",
			expected: false,
		},
		{
			name:     "HTTPS URL",
			baseURL:  "http://127.0.0.1:8085",
			linkURL:  "https://other.com",
			expected: false,
		},
		{
			name:     "Mailto URL",
			baseURL:  "http://127.0.0.1:8085",
			linkURL:  "mailto:user@example.com",
			expected: true,
		},
		{
			name:     "Tel URL",
			baseURL:  "http://127.0.0.1:8085",
			linkURL:  "tel:1234567890",
			expected: true,
		},
		{
			name:     "JavaScript URL",
			baseURL:  "http://127.0.0.1:8085",
			linkURL:  "javascript:void(0)",
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseURL, _ := url.Parse(tc.baseURL)
			linkURL, _ := url.Parse(tc.linkURL)
			result := ShouldSkipURL(baseURL, linkURL)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestCheckLink tests the checkLink function
func TestCheckLink(t *testing.T) {
	// Set up test servers
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()

	notFoundServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer notFoundServer.Close()

	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, okServer.URL, http.StatusFound)
	}))
	defer redirectServer.Close()

	testCases := []struct {
		name           string
		url            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid URL with 200 response",
			url:            okServer.URL,
			expectedStatus: 200,
			expectedError:  "",
		},
		{
			name:           "Valid URL with 404 response",
			url:            notFoundServer.URL,
			expectedStatus: 404,
			expectedError:  "",
		},
		{
			name:           "Valid URL with redirect",
			url:            redirectServer.URL,
			expectedStatus: 200,
			expectedError:  "",
		},
		{
			name:           "Invalid URL",
			url:            "http://localhost:99999", // Invalid port
			expectedStatus: 0,
			expectedError:  "not empty", // Just check that error is not empty
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			status, errMsg := CheckLink(tc.url)

			if tc.expectedError != "" {
				assert.NotEmpty(t, errMsg)
			} else {
				assert.Empty(t, errMsg)
				assert.Equal(t, tc.expectedStatus, status)
			}
		})
	}
}

// TestCountBrokenLinks tests the CountBrokenLinks function
func TestCountBrokenLinks(t *testing.T) {
	teardown := setupTest()
	defer teardown()

	testCases := []struct {
		name     string
		results  []model.LinkResult
		expected int
	}{
		{
			name:     "No links",
			results:  []model.LinkResult{},
			expected: 0,
		},
		{
			name: "No broken links",
			results: []model.LinkResult{
				{Status: 200, Error: ""},
				{Status: 302, Error: ""},
			},
			expected: 0,
		},
		{
			name: "Only error links",
			results: []model.LinkResult{
				{Status: 0, Error: "connection error"},
				{Status: 0, Error: "timeout"},
			},
			expected: 2,
		},
		{
			name: "Only status error links",
			results: []model.LinkResult{
				{Status: 404, Error: ""},
				{Status: 500, Error: ""},
			},
			expected: 2,
		},
		{
			name: "Mixed links",
			results: []model.LinkResult{
				{Status: 200, Error: ""},
				{Status: 404, Error: ""},
				{Status: 0, Error: "timeout"},
				{Status: 302, Error: ""},
			},
			expected: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			model.Results = tc.results
			count := CountBrokenLinks()
			assert.Equal(t, tc.expected, count)
		})
	}
}
