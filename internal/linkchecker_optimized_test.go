package internal

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
)

func TestOptimizedLinkCheckerService_CheckLink(t *testing.T) {
	// Initialize logger for tests
	logger.InitLogger("error")
	defer logger.CloseLogger()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "HEAD":
			w.WriteHeader(http.StatusOK)
		case "GET":
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("<html><body>Test page</body></html>"))
		}
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	checker := NewOptimizedLinkCheckerService(client, "Test/1.0", 5*time.Second, 2.0, 5.0)

	status, errMsg := checker.CheckLink(server.URL)
	if status != 200 {
		t.Errorf("Expected status 200, got %d", status)
	}
	if errMsg != "" {
		t.Errorf("Expected no error message, got %s", errMsg)
	}

	// Check stats
	stats := checker.GetOptimizationStats()
	if stats.HeadRequestsUsed == 0 {
		t.Error("Expected HEAD request to be used")
	}
}

func TestOptimizedLinkCheckerService_HeadFallback(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()

	// Server that doesn't support HEAD
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Test page</body></html>"))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	checker := NewOptimizedLinkCheckerService(client, "Test/1.0", 5*time.Second, 2.0, 5.0)

	status, errMsg := checker.CheckLink(server.URL)
	if status != 200 {
		t.Errorf("Expected status 200, got %d", status)
	}
	if errMsg != "" {
		t.Errorf("Expected no error message, got %s", errMsg)
	}

	// Check that fallback occurred
	stats := checker.GetOptimizationStats()
	if stats.HeadFallbacks == 0 {
		t.Error("Expected HEAD fallback to occur")
	}
	if stats.GetRequestsUsed == 0 {
		t.Error("Expected GET request to be used after fallback")
	}

	// Second request to same domain should skip HEAD
	status2, errMsg2 := checker.CheckLink(server.URL + "/page2")
	if status2 != 200 {
		t.Errorf("Expected status 200, got %d", status2)
	}
	if errMsg2 != "" {
		t.Errorf("Expected no error message, got %s", errMsg2)
	}

	// Should have used GET directly this time
	stats2 := checker.GetOptimizationStats()
	if stats2.GetRequestsUsed != 2 {
		t.Errorf("Expected 2 GET requests, got %d", stats2.GetRequestsUsed)
	}
}

func TestOptimizedLinkCheckerService_RateLimit(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	checker := NewOptimizedLinkCheckerService(client, "Test/1.0", 5*time.Second, 0.5, 1.0) // Very low rate

	start := time.Now()
	
	// Make two requests - second should be rate limited
	checker.CheckLink(server.URL)
	checker.CheckLink(server.URL)
	
	elapsed := time.Since(start)
	
	// Should take at least 2 seconds due to rate limiting (1/0.5 = 2 seconds)
	if elapsed < 1*time.Second {
		t.Errorf("Expected rate limiting delay, but requests completed in %v", elapsed)
	}
}

func TestOptimizedLinkCheckerService_FileDetection(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()

	// Server serving different content types
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/image.jpg":
			w.Header().Set("Content-Type", "image/jpeg")
		case "/document.pdf":
			w.Header().Set("Content-Type", "application/pdf")
		case "/page.html":
			w.Header().Set("Content-Type", "text/html")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	checker := NewOptimizedLinkCheckerService(client, "Test/1.0", 5*time.Second, 2.0, 5.0)

	// Test image file
	status, errMsg := checker.CheckLink(server.URL + "/image.jpg")
	if status != 200 {
		t.Errorf("Expected status 200 for image, got %d", status)
	}
	if errMsg != "" {
		t.Errorf("Expected no error for image, got %s", errMsg)
	}

	// Test PDF file
	status, errMsg = checker.CheckLink(server.URL + "/document.pdf")
	if status != 200 {
		t.Errorf("Expected status 200 for PDF, got %d", status)
	}
	if errMsg != "" {
		t.Errorf("Expected no error for PDF, got %s", errMsg)
	}

	// Test HTML page
	status, errMsg = checker.CheckLink(server.URL + "/page.html")
	if status != 200 {
		t.Errorf("Expected status 200 for HTML, got %d", status)
	}
	if errMsg != "" {
		t.Errorf("Expected no error for HTML, got %s", errMsg)
	}
}

func TestOptimizedLinkCheckerService_DomainRateLimit(t *testing.T) {
	logger.InitLogger("error")
	defer logger.CloseLogger()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}
	checker := NewOptimizedLinkCheckerService(client, "Test/1.0", 5*time.Second, 2.0, 5.0)

	// Test that rate limiter stats are available after setting domain rate limit
	domain, _ := extractDomain(server.URL)
	checker.SetDomainRateLimit(domain, 1.0)

	// Make one request to initialize the bucket
	checker.CheckLink(server.URL)

	// Check rate limiter stats
	stats := checker.GetRateLimiterStats()
	if _, exists := stats[domain]; !exists {
		t.Error("Expected rate limiter stats for domain")
	}
	
	if stat, exists := stats[domain]; exists {
		if stat.Rate != 1.0 {
			t.Errorf("Expected rate 1.0, got %f", stat.Rate)
		}
	}
}

func TestOptimizedLinkStats(t *testing.T) {
	stats := &OptimizedLinkStats{}

	stats.incrementHeadRequests()
	stats.incrementGetRequests()
	stats.incrementHeadFallbacks()
	stats.addBytesSaved(1024)
	stats.addTimeSaved(100 * time.Millisecond)

	if stats.HeadRequestsUsed != 1 {
		t.Errorf("Expected 1 HEAD request, got %d", stats.HeadRequestsUsed)
	}
	if stats.GetRequestsUsed != 1 {
		t.Errorf("Expected 1 GET request, got %d", stats.GetRequestsUsed)
	}
	if stats.HeadFallbacks != 1 {
		t.Errorf("Expected 1 HEAD fallback, got %d", stats.HeadFallbacks)
	}
	if stats.BytesSaved != 1024 {
		t.Errorf("Expected 1024 bytes saved, got %d", stats.BytesSaved)
	}
	if stats.TimeSaved != 100*time.Millisecond {
		t.Errorf("Expected 100ms time saved, got %v", stats.TimeSaved)
	}
}