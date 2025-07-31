package internal

import (
	"sync"
	"testing"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/stretchr/testify/assert"
)

func TestDomainRateLimiter(t *testing.T) {
	// Initialize logger for tests
	model.Quiet = true
	logger.InitLogger("debug")
	defer logger.CloseLogger()
	
	t.Run("Creates rate limiter", func(t *testing.T) {
		limiter := NewDomainRateLimiter(1.0, 2.0)
		assert.NotNil(t, limiter)
	})

	t.Run("Allows requests within rate limit", func(t *testing.T) {
		limiter := NewDomainRateLimiter(10.0, 5.0) // 10 req/s, burst 5
		
		// First few requests should be immediate (burst)
		start := time.Now()
		for i := 0; i < 3; i++ {
			err := limiter.Wait("https://example.com/test")
			assert.NoError(t, err)
		}
		elapsed := time.Since(start)
		
		// Should complete quickly due to burst capacity
		assert.Less(t, elapsed, 100*time.Millisecond)
	})

	t.Run("Rate limits requests", func(t *testing.T) {
		limiter := NewDomainRateLimiter(2.0, 1.0) // 2 req/s, burst 1
		
		// First request should be immediate
		start := time.Now()
		err := limiter.Wait("https://example.com/test")
		assert.NoError(t, err)
		
		// Second request should be rate limited
		err = limiter.Wait("https://example.com/test")
		assert.NoError(t, err)
		elapsed := time.Since(start)
		
		// Should take at least ~500ms (1/2 second for 2 req/s)
		assert.GreaterOrEqual(t, elapsed, 400*time.Millisecond)
	})

	t.Run("Handles different domains separately", func(t *testing.T) {
		limiter := NewDomainRateLimiter(1.0, 1.0) // 1 req/s, burst 1
		
		// Both domains should allow immediate first request
		start := time.Now()
		
		err1 := limiter.Wait("https://domain1.com/test")
		assert.NoError(t, err1)
		
		err2 := limiter.Wait("https://domain2.com/test")
		assert.NoError(t, err2)
		
		elapsed := time.Since(start)
		
		// Both should complete quickly since they're different domains
		assert.Less(t, elapsed, 100*time.Millisecond)
	})

	t.Run("Updates domain configuration", func(t *testing.T) {
		limiter := NewDomainRateLimiter(1.0, 1.0)
		
		// Set a high rate for example.com
		limiter.UpdateConfig("example.com", 100.0)
		
		// Multiple requests should be fast
		start := time.Now()
		for i := 0; i < 3; i++ {
			err := limiter.Wait("https://example.com/test")
			assert.NoError(t, err)
		}
		elapsed := time.Since(start)
		
		assert.Less(t, elapsed, 100*time.Millisecond)
	})

	t.Run("Returns statistics", func(t *testing.T) {
		limiter := NewDomainRateLimiter(2.0, 3.0)
		
		// Make a request to create a bucket
		err := limiter.Wait("https://example.com/test")
		assert.NoError(t, err)
		
		stats := limiter.GetStats()
		assert.Contains(t, stats, "example.com")
		
		domainStats := stats["example.com"]
		assert.Equal(t, "example.com", domainStats.Domain)
		assert.Equal(t, 2.0, domainStats.Rate)
		assert.Equal(t, 3.0, domainStats.MaxTokens)
	})

	t.Run("Handles concurrent requests safely", func(t *testing.T) {
		limiter := NewDomainRateLimiter(5.0, 2.0) // 5 req/s, burst 2
		
		var wg sync.WaitGroup
		errors := make(chan error, 10)
		
		// Launch multiple concurrent requests
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := limiter.Wait("https://example.com/test")
				errors <- err
			}()
		}
		
		wg.Wait()
		close(errors)
		
		// All should succeed
		for err := range errors {
			assert.NoError(t, err)
		}
	})
}

func TestTokenBucket(t *testing.T) {
	t.Run("Refills tokens over time", func(t *testing.T) {
		bucket := &TokenBucket{
			tokens:     0,
			maxTokens:  2,
			refillRate: 4.0, // 4 tokens per second
			lastRefill: time.Now().Add(-500 * time.Millisecond), // 0.5 seconds ago
		}
		
		// Should not block since bucket should have refilled ~2 tokens
		err := bucket.wait()
		assert.NoError(t, err)
	})
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		url      string
		expected string
		hasError bool
	}{
		{"https://example.com/path", "example.com", false},
		{"http://subdomain.example.com:8080/path", "subdomain.example.com:8080", false},
		{"ftp://example.com", "example.com", false},
		{"invalid-url", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			domain, err := extractDomain(tt.url)
			
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, domain)
			}
		})
	}
}