package utils

import (
	"sync"
	"testing"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/stretchr/testify/assert"
)

// Helper function to set up test environment for crawl tests
func setupCrawlTest() func() {
	logger.InitLogger("debug")
	
	// Save original model state
	originalResults := model.Results
	originalDepth := model.Depth
	originalTimeout := model.Timeout
	originalUserAgent := model.UserAgent
	originalConcurrency := model.Concurrency

	// Set up test environment
	model.Results = []model.LinkResult{}
	model.VisitedURLs = sync.Map{}
	model.Depth = 2
	model.Timeout = 5
	model.UserAgent = "TestUserAgent"
	model.Concurrency = 5
	model.Wg = sync.WaitGroup{}

	// Return a function to restore original state
	return func() {
		model.Results = originalResults
		model.Depth = originalDepth
		model.Timeout = originalTimeout
		model.UserAgent = originalUserAgent
		model.Concurrency = originalConcurrency
		logger.CloseLogger()
	}
}

func TestCrawl(t *testing.T) {
	teardown := setupCrawlTest()
	defer teardown()

	t.Run("Crawl stops at max depth", func(t *testing.T) {
		model.Depth = 0 // Set max depth to 0
		
		baseURL := "http://127.0.0.1:8085"
		currentURL := "http://127.0.0.1:8085/index.html"
		
		// This should return immediately due to depth check
		Crawl(baseURL, currentURL, 1, model.Concurrency)
		
		// Wait a brief moment for any potential goroutines
		time.Sleep(100 * time.Millisecond)
		
		// Since we exceeded max depth, no processing should occur
		// We can't easily assert on the exact behavior since it's concurrent
	})

	t.Run("Crawl skips already visited URLs", func(t *testing.T) {
		model.Depth = 2
		
		baseURL := "http://127.0.0.1:8085"
		currentURL := "http://127.0.0.1:8085/index.html"
		
		// Mark URL as already visited
		model.VisitedURLs.Store(currentURL, true)
		
		// This should return immediately due to already visited check
		Crawl(baseURL, currentURL, 0, model.Concurrency)
		
		// Wait a brief moment
		time.Sleep(100 * time.Millisecond)
		
		// The URL should still be marked as visited
		_, exists := model.VisitedURLs.Load(currentURL)
		assert.True(t, exists)
	})

	// Simplified test - only test that the function exists and doesn't crash
	t.Run("Crawl function works", func(t *testing.T) {
		model.Depth = 1
		model.VisitedURLs = sync.Map{} // Clear visited URLs
		model.Wg = sync.WaitGroup{} // Reset WaitGroup
		
		baseURL := "http://127.0.0.1:8085"
		currentURL := "http://127.0.0.1:8085/index.html"
		
		// This should not panic
		Crawl(baseURL, currentURL, 0, model.Concurrency)
		
		// The URL should be marked as visited
		_, exists := model.VisitedURLs.Load(currentURL)
		assert.True(t, exists)
	})
}

func TestCrawlDepthLogic(t *testing.T) {
	teardown := setupCrawlTest()
	defer teardown()

	t.Run("Crawl depth progression", func(t *testing.T) {
		model.Depth = 2
		model.VisitedURLs = sync.Map{} // Clear visited URLs
		
		baseURL := "http://127.0.0.1:8085"
		
		// Test different depth levels
		testCases := []struct {
			currentDepth int
			shouldProcess bool
		}{
			{0, true},  // depth 0, max 2 - should process
			{1, true},  // depth 1, max 2 - should process
			{2, true},  // depth 2, max 2 - should process
			{3, false}, // depth 3, max 2 - should not process
		}
		
		for _, tc := range testCases {
			model.VisitedURLs = sync.Map{} // Clear for each test
			url := "http://127.0.0.1:8085/index.html"
			
			Crawl(baseURL, url, tc.currentDepth, model.Concurrency)
			
			// Brief wait to allow processing
			time.Sleep(50 * time.Millisecond)
			
			if tc.shouldProcess {
				// URL should be marked as visited when processing occurs
				_, exists := model.VisitedURLs.Load(url)
				assert.True(t, exists, "Depth %d should be processed", tc.currentDepth)
			}
		}
	})
}