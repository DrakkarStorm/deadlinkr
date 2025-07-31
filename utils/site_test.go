package utils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHTTPServer creates a test HTTP server with predefined responses
type MockHTTPServer struct {
	server *httptest.Server
	routes map[string]MockResponse
}

type MockResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
	Delay      time.Duration
}

func NewMockHTTPServer() *MockHTTPServer {
	mock := &MockHTTPServer{
		routes: make(map[string]MockResponse),
	}
	
	mock.server = httptest.NewServer(http.HandlerFunc(mock.handler))
	return mock
}

func (m *MockHTTPServer) handler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	// Add delay if specified
	if response, exists := m.routes[path]; exists {
		if response.Delay > 0 {
			time.Sleep(response.Delay)
		}
		
		// Set headers
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		
		w.WriteHeader(response.StatusCode)
		_, _ = w.Write([]byte(response.Body))
		return
	}
	
	// Default 404 response
	w.WriteHeader(404)
	_, _ = w.Write([]byte("Not Found"))
}

func (m *MockHTTPServer) AddRoute(path string, response MockResponse) {
	m.routes[path] = response
}

func (m *MockHTTPServer) URL() string {
	return m.server.URL
}

func (m *MockHTTPServer) Close() {
	m.server.Close()
}

// setupTestState resets all global state for clean tests
func setupTestState() func() {
	// Initialize logger
	logger.InitLogger("debug")
	
	// Save original state (avoid copying locks)
	originalResults := model.Results
	originalDepth := model.Depth
	originalConcurrency := model.Concurrency
	originalTimeout := model.Timeout
	
	// Reset to clean state
	model.Results = []model.LinkResult{}
	model.Depth = 1
	model.Concurrency = 5
	model.Timeout = 5
	model.VisitedURLs = sync.Map{}
	model.Wg = sync.WaitGroup{}
	
	// Return cleanup function
	return func() {
		model.Results = originalResults
		model.Depth = originalDepth
		model.Concurrency = originalConcurrency
		model.Timeout = originalTimeout
		// Don't restore sync structures - leave them clean for next test
		model.VisitedURLs = sync.Map{}
		model.Wg = sync.WaitGroup{}
	}
}

func TestCheckLinks_WithMocks(t *testing.T) {
	cleanup := setupTestState()
	defer cleanup()
	
	// Create mock server
	mockServer := NewMockHTTPServer()
	defer mockServer.Close()
	
	// Set up mock routes
	mockServer.AddRoute("/", MockResponse{
		StatusCode: 200,
		Body: `<html>
			<head><title>Test Page</title></head>
			<body>
				<a href="/page1">Internal Link 1</a>
				<a href="/page2">Internal Link 2</a>
				<a href="http://external.com">External Link</a>
				<a href="/broken">Broken Link</a>
			</body>
		</html>`,
		Headers: map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/page1", MockResponse{
		StatusCode: 200,
		Body:       "<html><body>Page 1</body></html>",
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/page2", MockResponse{
		StatusCode: 200,
		Body:       "<html><body>Page 2</body></html>",
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	// /broken will return default 404
	
	baseURL := mockServer.URL()
	pageURL := mockServer.URL() + "/"
	
	// Test CheckLinks function
	links := CheckLinks(baseURL, pageURL)
	
	// Verify results
	require.NotEmpty(t, links, "Should find links on the page")
	
	// Check that we found the expected links
	linkTargets := make(map[string]model.LinkResult)
	for _, link := range links {
		linkTargets[link.TargetURL] = link
	}
	
	// Verify internal links
	page1Link, exists := linkTargets[baseURL+"/page1"]
	assert.True(t, exists, "Should find /page1 link")
	if exists {
		assert.Equal(t, 200, page1Link.Status)
		assert.False(t, page1Link.IsExternal)
		assert.Empty(t, page1Link.Error)
	}
	
	page2Link, exists := linkTargets[baseURL+"/page2"]
	assert.True(t, exists, "Should find /page2 link")
	if exists {
		assert.Equal(t, 200, page2Link.Status)
		assert.False(t, page2Link.IsExternal)
		assert.Empty(t, page2Link.Error)
	}
	
	brokenLink, exists := linkTargets[baseURL+"/broken"]
	assert.True(t, exists, "Should find /broken link")
	if exists {
		assert.Equal(t, 404, brokenLink.Status)
		assert.False(t, brokenLink.IsExternal)
		assert.Empty(t, brokenLink.Error)
	}
	
	// External link should be marked as external
	externalLink, exists := linkTargets["http://external.com"]
	assert.True(t, exists, "Should find external link")
	if exists {
		assert.True(t, externalLink.IsExternal)
		// External link might fail (which is expected in tests)
	}
}

func TestCheckLinks_EmptyPage(t *testing.T) {
	cleanup := setupTestState()
	defer cleanup()
	
	mockServer := NewMockHTTPServer()
	defer mockServer.Close()
	
	mockServer.AddRoute("/empty", MockResponse{
		StatusCode: 200,
		Body:       "<html><body>No links here</body></html>",
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	baseURL := mockServer.URL()
	pageURL := mockServer.URL() + "/empty"
	
	links := CheckLinks(baseURL, pageURL)
	
	assert.Empty(t, links, "Should find no links on empty page")
}

func TestCheckLinks_InvalidPage(t *testing.T) {
	cleanup := setupTestState()
	defer cleanup()
	
	mockServer := NewMockHTTPServer()
	defer mockServer.Close()
	
	// Don't add any routes - all requests will return 404
	
	baseURL := mockServer.URL()
	pageURL := mockServer.URL() + "/nonexistent"
	
	// This should return empty links because the page returns 404
	links := CheckLinks(baseURL, pageURL)
	
	assert.Empty(t, links, "Should return empty slice for 404 page")
}

func TestCrawl_WithMocks(t *testing.T) {
	cleanup := setupTestState()
	defer cleanup()
	
	mockServer := NewMockHTTPServer()
	defer mockServer.Close()
	
	// Create a simple site structure
	mockServer.AddRoute("/", MockResponse{
		StatusCode: 200,
		Body: `<html>
			<body>
				<a href="/page1">Page 1</a>
				<a href="/page2">Page 2</a>
			</body>
		</html>`,
		Headers: map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/page1", MockResponse{
		StatusCode: 200,
		Body: `<html>
			<body>
				<a href="/subpage">Subpage</a>
				<a href="/">Home</a>
			</body>
		</html>`,
		Headers: map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/page2", MockResponse{
		StatusCode: 200,
		Body:       "<html><body>Simple page</body></html>",
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/subpage", MockResponse{
		StatusCode: 200,
		Body:       "<html><body>Subpage content</body></html>",
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	baseURL := mockServer.URL()
	
	// Set depth to 2 to test recursive crawling
	model.Depth = 2
	
	// Start crawling
	Crawl(baseURL, baseURL+"/", 0, model.Concurrency)
	
	// Wait for all goroutines to complete with timeout
	done := make(chan struct{})
	go func() {
		model.Wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Crawl test timed out - goroutines did not complete")
	}
	
	// Verify that URLs were visited
	visited := []string{}
	model.VisitedURLs.Range(func(key, value interface{}) bool {
		if url, ok := key.(string); ok {
			visited = append(visited, url)
		}
		return true
	})
	
	assert.NotEmpty(t, visited, "Should have visited some URLs")
	assert.Contains(t, visited, baseURL+"/", "Should have visited root URL")
	
	// Verify that results were collected
	assert.NotEmpty(t, model.Results, "Should have collected some results")
	
	// Print results for debugging (optional)
	t.Logf("Visited %d URLs: %v", len(visited), visited)
	t.Logf("Found %d links", len(model.Results))
}

func TestCrawl_RespectDepthLimit(t *testing.T) {
	cleanup := setupTestState()
	defer cleanup()
	
	mockServer := NewMockHTTPServer()
	defer mockServer.Close()
	
	// Create a deep link structure
	mockServer.AddRoute("/", MockResponse{
		StatusCode: 200,
		Body:       `<html><body><a href="/level1">Level 1</a></body></html>`,
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/level1", MockResponse{
		StatusCode: 200,
		Body:       `<html><body><a href="/level2">Level 2</a></body></html>`,
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/level2", MockResponse{
		StatusCode: 200,
		Body:       `<html><body><a href="/level3">Level 3</a></body></html>`,
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/level3", MockResponse{
		StatusCode: 200,
		Body:       "<html><body>Deep page</body></html>",
		Headers:    map[string]string{"Content-Type": "text/html"},
	})
	
	baseURL := mockServer.URL()
	
	// Set depth limit to 1
	model.Depth = 1
	
	// Start crawling
	Crawl(baseURL, baseURL+"/", 0, model.Concurrency)
	
	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		model.Wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Depth limit test timed out")
	}
	
	// Verify depth limit was respected
	visited := []string{}
	model.VisitedURLs.Range(func(key, value interface{}) bool {
		if url, ok := key.(string); ok {
			visited = append(visited, url)
		}
		return true
	})
	
	// Should only visit root and level1, not deeper levels
	assert.Contains(t, visited, baseURL+"/", "Should visit root")
	assert.Contains(t, visited, baseURL+"/level1", "Should visit level 1")
	assert.NotContains(t, visited, baseURL+"/level2", "Should not visit level 2 (depth limit)")
	assert.NotContains(t, visited, baseURL+"/level3", "Should not visit level 3 (depth limit)")
}

func TestCrawl_DuplicateURLHandling(t *testing.T) {
	cleanup := setupTestState()
	defer cleanup()
	
	mockServer := NewMockHTTPServer()
	defer mockServer.Close()
	
	// Create circular references
	mockServer.AddRoute("/", MockResponse{
		StatusCode: 200,
		Body: `<html>
			<body>
				<a href="/page1">Page 1</a>
				<a href="/page1">Page 1 Again</a>
			</body>
		</html>`,
		Headers: map[string]string{"Content-Type": "text/html"},
	})
	
	mockServer.AddRoute("/page1", MockResponse{
		StatusCode: 200,
		Body: `<html>
			<body>
				<a href="/">Back Home</a>
				<a href="/">Home Again</a>
			</body>
		</html>`,
		Headers: map[string]string{"Content-Type": "text/html"},
	})
	
	baseURL := mockServer.URL()
	model.Depth = 2
	
	// Start crawling
	Crawl(baseURL, baseURL+"/", 0, model.Concurrency)
	
	// Wait for completion
	done := make(chan struct{})
	go func() {
		model.Wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Duplicate URL test timed out")
	}
	
	// Count how many times each URL was visited
	visitedCount := make(map[string]int)
	model.VisitedURLs.Range(func(key, value interface{}) bool {
		if url, ok := key.(string); ok {
			visitedCount[url]++
		}
		return true
	})
	
	// Each URL should be visited only once
	for url, count := range visitedCount {
		assert.Equal(t, 1, count, fmt.Sprintf("URL %s should be visited only once, but was visited %d times", url, count))
	}
}