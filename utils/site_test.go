package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestCheckLinks(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		pageURL   string
		wantLinks []model.LinkResult
	}{
		{
			name:    "Valid BaseURL and PageURL",
			baseURL: "http://127.0.0.1:8085",
			pageURL: "http://127.0.0.1:8085/installation.html",
			wantLinks: []model.LinkResult{
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/index.html",
					Status:     200,
					IsExternal: false,
				},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/installation.html",
					Status:     200,
					IsExternal: false,
				},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/tutoriel.html",
					Status:     200,
					IsExternal: false,
				},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/api.html",
					Status:     404,
					IsExternal: false},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/exemples.html",
					Status:     404,
					IsExternal: false},
				{
					SourceURL: "http://127.0.0.1:8085/installation.html",
					TargetURL: "http://127.0.0.1:8085/faq.html",
					Status:    404, IsExternal: false},
				{
					SourceURL: "http://127.0.0.1:8085/installation.html",
					TargetURL: "http://127.0.0.1:8085/ressources.html",
					Status:    404, IsExternal: false},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/contributeurs.html",
					Status:     404,
					IsExternal: false},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/page-inexistante.html",
					Status:     404,
					IsExternal: false},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/contact.html",
					Status:     404,
					IsExternal: false,
				},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "https://golang.org/dl/",
					Status:     405,
					IsExternal: true},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "https://git-scm.com/",
					Status:     200,
					IsExternal: true},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "https://non-existent-domain-123456.xyz/",
					Status:     0,
					Error:      "Head \"https://non-existent-domain-123456.xyz/\": dial tcp: lookup non-existent-domain-123456.xyz: no such host",
					IsExternal: true},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/configuration.html",
					Status:     404,
					IsExternal: false},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "https://another-wrong-domain.org/docs",
					Status:     0,
					Error:      "Head \"https://another-wrong-domain.org/docs\": dial tcp: lookup another-wrong-domain.org: no such host",
					IsExternal: true},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/forum.html",
					Status:     404,
					IsExternal: false},

				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/faq.html",
					Status:     404,
					IsExternal: false},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/contact.html",
					Status:     404,
					IsExternal: false},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "http://127.0.0.1:8085/mentions-legales.html",
					Status:     404,
					IsExternal: false},
				{
					SourceURL:  "http://127.0.0.1:8085/installation.html",
					TargetURL:  "https://github.com/",
					Status:     200,
					IsExternal: true},
			},
		},
		{
			name:      "Invalid BaseURL",
			baseURL:   "invalid_url",
			pageURL:   "http://127.0.0.1:8085/installation.html",
			wantLinks: []model.LinkResult{},
		},
		{
			name:      "Invalid PageURL",
			baseURL:   "http://127.0.0.1:8085",
			pageURL:   "invalid_url",
			wantLinks: []model.LinkResult{},
		},
	}
	fmt.Printf("Running tests: %v\n", tests)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLinks := CheckLinks(tt.baseURL, tt.pageURL)
			if !compareLinkResults(gotLinks, tt.wantLinks) {
				t.Errorf("CheckLinks() = %v, want %v", gotLinks, tt.wantLinks)
			}
		})
	}
}

func compareLinkResults(a, b []model.LinkResult) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParseBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		expected *url.URL
		wantErr  bool
	}{
		{
			name:    "Valid URL",
			baseURL: "http://127.0.0.1:8085",
			expected: &url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:8085",
			},
			wantErr: false,
		},
		{
			name:     "Invalid URL",
			baseURL:  "invalid-url",
			expected: nil,
			wantErr:  true,
		},
		{
			name:    "URL with path",
			baseURL: "http://127.0.0.1:8085/installation.html",
			expected: &url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:8085",
				Path:   "/installation.html",
			},
			wantErr: false,
		},
		{
			name:    "URL with query",
			baseURL: "https://127.0.0.1:8085/installation.html?query=value",
			expected: &url.URL{
				Scheme:   "https",
				Host:     "127.0.0.1:8085",
				Path:     "/installation.html",
				RawQuery: "query=value",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBaseURL(tt.baseURL)
			if tt.wantErr && got != nil {
				t.Errorf("parseBaseURL() = %v, wantErr %v", got, tt.wantErr)
			}
			if !tt.wantErr && got == nil {
				t.Errorf("parseBaseURL() = %v, wantErr %v", got, tt.wantErr)
			}
			if !tt.wantErr && got != nil && got.String() != tt.expected.String() {
				t.Errorf("parseBaseURL() = %v, want %v", got.String(), tt.expected.String())
			}
		})
	}
}

func TestFetchAndParseDocument(t *testing.T) {
	t.Run("valid HTML document", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte("<html><body><h1>Hello, world!</h1></body></html>"))
			if err != nil {
				t.Fatalf("Failed to write response: %v", err)
			}
		}))
		defer server.Close()

		doc := fetchAndParseDocument(server.URL)
		assert.NotNil(t, doc)
		assert.Equal(t, "Hello, world!", doc.Find("h1").Text())
	})

	t.Run("invalid content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"message":"Hello, world!"}`))
			if err != nil {
				t.Fatalf("Failed to write response: %v", err)
			}
		}))
		defer server.Close()

		doc := fetchAndParseDocument(server.URL)
		assert.Nil(t, doc)
	})

	t.Run("HTTP error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		doc := fetchAndParseDocument(server.URL)
		assert.Nil(t, doc)
	})

	t.Run("invalid URL", func(t *testing.T) {
		doc := fetchAndParseDocument("invalid URL")
		assert.Nil(t, doc)
	})
}

func TestExtractLinks(t *testing.T) {
	// Sauvegarder les valeurs originales des variables globales pour les restaurer après
	originalIgnoreExternal := model.IgnoreExternal
	originalOnlyExternal := model.OnlyExternal
	defer func() {
		model.IgnoreExternal = originalIgnoreExternal
		model.OnlyExternal = originalOnlyExternal
	}()

	// Configurer un serveur de test pour les URLs internes et externes
	internalServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer internalServer.Close()

	baseURL, _ := url.Parse(internalServer.URL)

	testCases := []struct {
		name           string
		html           string
		ignoreExternal bool
		onlyExternal   bool
		expectedCount  int
		expectedURLs   []string
	}{
		{
			name: "Basic internal links",
			html: `<html><body>
				<a href="/page1">Page 1</a>
				<a href="/page2">Page 2</a>
				<a href="#section">Section</a>
			</body></html>`,
			ignoreExternal: false,
			onlyExternal:   false,
			expectedCount:  2,
			expectedURLs:   []string{internalServer.URL + "/page1", internalServer.URL + "/page2"},
		},
		{
			name: "Mix of internal and external links",
			html: `<html><body>
				<a href="/internal">Internal</a>
				<a href="http://external.server.com/external">External</a>
			</body></html>`,
			ignoreExternal: false,
			onlyExternal:   false,
			expectedCount:  2,
			expectedURLs:   []string{internalServer.URL + "/internal", "http://external.server.com/external"},
		},
		{
			name: "Ignore external links",
			html: `<html><body>
				<a href="/internal1">Internal 1</a>
				<a href="/internal2">Internal 2</a>
				<a href="http://external.server.com/external">External</a>
			</body></html>`,
			ignoreExternal: true,
			onlyExternal:   false,
			expectedCount:  2,
			expectedURLs:   []string{internalServer.URL + "/internal1", internalServer.URL + "/internal2"},
		},
		{
			name: "Only external links",
			html: `<html><body>
				<a href="/internal">Internal</a>
				<a href="http://external.server.com/external1">External 1</a>
				<a href="http://external.server.com/external2">External 2</a>
			</body></html>`,
			ignoreExternal: false,
			onlyExternal:   true,
			expectedCount:  2,
			expectedURLs:   []string{"http://external.server.com/external1", "http://external.server.com/external2"},
		},
		{
			name: "Empty hrefs and fragment links should be ignored",
			html: `<html><body>
				<a href="">Empty</a>
				<a href="#">Fragment</a>
				<a href="/valid">Valid</a>
			</body></html>`,
			ignoreExternal: false,
			onlyExternal:   false,
			expectedCount:  1,
			expectedURLs:   []string{internalServer.URL + "/valid"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Configurer les variables globales pour ce test
			model.IgnoreExternal = tc.ignoreExternal
			model.OnlyExternal = tc.onlyExternal

			// Créer le document à partir du HTML
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			assert.NoError(t, err)

			// Appeler la fonction à tester
			links := extractLinks(baseURL, internalServer.URL, doc)

			// Vérifier le nombre de liens extraits
			assert.Equal(t, tc.expectedCount, len(links))

			// Vérifier que les URLs attendues sont présentes
			foundURLs := make([]string, 0)
			for _, link := range links {
				foundURLs = append(foundURLs, link.TargetURL)
			}

			for _, expectedURL := range tc.expectedURLs {
				assert.Contains(t, foundURLs, expectedURL)
			}
		})
	}
}

func TestResolveAndFilterURL(t *testing.T) {
	// Préparer une URL de base pour les tests
	baseURL, _ := url.Parse("http://127.0.0.1:8085")

	tests := []struct {
		name     string
		baseURL  *url.URL
		pageURL  string
		href     string
		expected *url.URL
	}{
		{
			name:     "Valid relative URL",
			baseURL:  baseURL,
			pageURL:  "http://127.0.0.1:8085/index.html",
			href:     "installation.html",
			expected: &url.URL{Scheme: "http", Host: "127.0.0.1:8085", Path: "/installation.html"},
		},
		{
			name:     "External URL should be filtered",
			baseURL:  baseURL,
			pageURL:  "http://127.0.0.1:8085/installation.html",
			href:     "https://golang.org/dl/",
			expected: &url.URL{Scheme: "https", Host: "golang.org", Path: "/dl/"},
		},
		{
			name:     "Invalid URL should return nil",
			baseURL:  baseURL,
			pageURL:  "http://127.0.0.1:8085/installation.html",
			href:     "http://[::1]:namedport",
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := resolveAndFilterURL(tc.baseURL, tc.pageURL, tc.href)

			// Vérifier si le résultat est nil quand on l'attend
			if tc.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}

			// Vérifier si le résultat est nil quand on ne l'attend pas
			if result == nil {
				t.Errorf("Expected %v, got nil", tc.expected)
				return
			}

			// Comparer les parties importantes de l'URL
			if result.Scheme != tc.expected.Scheme {
				t.Errorf("Expected scheme %s, got %s", tc.expected.Scheme, result.Scheme)
			}
			if result.Host != tc.expected.Host {
				t.Errorf("Expected host %s, got %s", tc.expected.Host, result.Host)
			}
			if result.Path != tc.expected.Path {
				t.Errorf("Expected path %s, got %s", tc.expected.Path, result.Path)
			}
			if result.RawQuery != tc.expected.RawQuery {
				t.Errorf("Expected query %s, got %s", tc.expected.RawQuery, result.RawQuery)
			}
			if result.Fragment != tc.expected.Fragment {
				t.Errorf("Expected fragment %s, got %s", tc.expected.Fragment, result.Fragment)
			}
		})
	}
}

func TestShouldSkipURLBasedOnPattern(t *testing.T) {
	// Save original values to restore after test
	originalIncludePattern := model.IncludePattern
	originalExcludePattern := model.ExcludePattern

	// Restore after test
	defer func() {
		model.IncludePattern = originalIncludePattern
		model.ExcludePattern = originalExcludePattern
	}()

	tests := []struct {
		name           string
		includePattern string
		excludePattern string
		urlString      string
		expected       bool
	}{
		// Tests with only include pattern
		{
			name:           "Include pattern matches URL",
			includePattern: "127.0.0.1:8085",
			excludePattern: "",
			urlString:      "http://127.0.0.1:8085/page",
			expected:       false, // Should not skip URLs that match include pattern
		},
		{
			name:           "Include pattern does not match URL",
			includePattern: "127.0.0.1:8085",
			excludePattern: "",
			urlString:      "https://different.com/page",
			expected:       true, // Should skip URLs that don't match include pattern
		},
		{
			name:           "Invalid include pattern",
			includePattern: "[", // Invalid regex
			excludePattern: "",
			urlString:      "http://127.0.0.1:8085/page",
			expected:       true, // Should skip due to regex error
		},

		// Tests with only exclude pattern
		{
			name:           "Exclude pattern matches URL",
			includePattern: "",
			excludePattern: "private",
			urlString:      "http://127.0.0.1:8085/private/page",
			expected:       true, // Should skip URLs that match exclude pattern
		},
		{
			name:           "Exclude pattern does not match URL",
			includePattern: "",
			excludePattern: "private",
			urlString:      "http://127.0.0.1:8085/public/page",
			expected:       false, // Should not skip URLs that don't match exclude pattern
		},
		{
			name:           "Invalid exclude pattern",
			includePattern: "",
			excludePattern: "[", // Invalid regex
			urlString:      "http://127.0.0.1:8085/page",
			expected:       false, // Should not skip due to regex error in exclude pattern
		},

		// Tests with both patterns
		{
			name:           "URL matches both include and exclude patterns",
			includePattern: "127.0.0.1:8085",
			excludePattern: "private",
			urlString:      "http://127.0.0.1:8085/private/page",
			expected:       true, // Should skip (exclude takes precedence)
		},
		{
			name:           "URL matches include but not exclude pattern",
			includePattern: "127.0.0.1:8085",
			excludePattern: "private",
			urlString:      "http://127.0.0.1:8085/public/page",
			expected:       false, // Should not skip
		},
		{
			name:           "URL matches neither include nor exclude pattern",
			includePattern: "127.0.0.1:8085",
			excludePattern: "private",
			urlString:      "https://different.com/page",
			expected:       true, // Should skip (doesn't match include)
		},

		// Edge cases
		{
			name:           "Empty URL",
			includePattern: "127.0.0.1:8085",
			excludePattern: "private",
			urlString:      "",
			expected:       true, // Should skip as it doesn't match include pattern
		},
		{
			name:           "No patterns specified",
			includePattern: "",
			excludePattern: "",
			urlString:      "http://127.0.0.1:8085/page",
			expected:       false, // Should not skip when no patterns are specified
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup patterns for this test case
			model.IncludePattern = tc.includePattern
			model.ExcludePattern = tc.excludePattern

			// Parse URL
			parsedURL, err := url.Parse(tc.urlString)
			if err != nil && tc.urlString != "" {
				t.Fatalf("Failed to parse URL %s: %v", tc.urlString, err)
			}

			// Run test
			result := shouldSkipURLBasedOnPattern(parsedURL)

			// Verify result
			if result != tc.expected {
				t.Errorf("shouldSkipURLBasedOnPattern() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

// Additional test to ensure thread safety when patterns are changed concurrently
func TestShouldSkipURLBasedOnPatternConcurrency(t *testing.T) {
	// Save original values to restore after test
	originalIncludePattern := model.IncludePattern
	originalExcludePattern := model.ExcludePattern

	// Restore after test
	defer func() {
		model.IncludePattern = originalIncludePattern
		model.ExcludePattern = originalExcludePattern
	}()

	// Test concurrent access
	done := make(chan bool)
	url1, _ := url.Parse("http://127.0.0.1:8085/installation.html")
	url2, _ := url.Parse("http://127.0.0.1:8085/tutoriel.html")

	for i := 0; i < 10; i++ {
		go func(i int) {
			// Alternate between different pattern combinations
			if i%2 == 0 {
				model.IncludePattern = "127.0.0.1:8085"
				model.ExcludePattern = ""
			} else {
				model.IncludePattern = ""
				model.ExcludePattern = "127.0.0.1:8085"
			}

			// Just ensure no panics occur
			_ = shouldSkipURLBasedOnPattern(url1)
			_ = shouldSkipURLBasedOnPattern(url2)

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
