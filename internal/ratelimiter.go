package internal

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
)

// RateLimiter manages rate limiting per domain
type RateLimiter interface {
	Wait(domain string) error
	UpdateConfig(domain string, requestsPerSecond float64)
}

// TokenBucket implements a token bucket rate limiter for a single domain
type TokenBucket struct {
	tokens         float64
	maxTokens      float64
	refillRate     float64  // tokens per second
	lastRefill     time.Time
	mutex          sync.Mutex
}

// DomainRateLimiter manages rate limiting across multiple domains
type DomainRateLimiter struct {
	buckets           map[string]*TokenBucket
	defaultRate       float64 // requests per second
	maxBurst          float64 // max tokens in bucket
	domainConfigs     map[string]float64 // custom rates per domain
	mutex             sync.RWMutex
}

// NewDomainRateLimiter creates a new domain-based rate limiter
func NewDomainRateLimiter(defaultRequestsPerSecond float64, maxBurst float64) *DomainRateLimiter {
	return &DomainRateLimiter{
		buckets:       make(map[string]*TokenBucket),
		defaultRate:   defaultRequestsPerSecond,
		maxBurst:      maxBurst,
		domainConfigs: make(map[string]float64),
	}
}

// Wait blocks until a request can be made to the given domain
func (drl *DomainRateLimiter) Wait(targetURL string) error {
	domain, err := extractDomain(targetURL)
	if err != nil {
		return err
	}

	bucket := drl.getBucket(domain)
	return bucket.wait()
}

// UpdateConfig sets a custom rate limit for a specific domain
func (drl *DomainRateLimiter) UpdateConfig(domain string, requestsPerSecond float64) {
	drl.mutex.Lock()
	defer drl.mutex.Unlock()
	
	drl.domainConfigs[domain] = requestsPerSecond
	
	// Update existing bucket if it exists
	if bucket, exists := drl.buckets[domain]; exists {
		bucket.mutex.Lock()
		bucket.refillRate = requestsPerSecond
		bucket.mutex.Unlock()
		logger.Debugf("Updated rate limit for %s to %.2f req/s", domain, requestsPerSecond)
	}
}

// getBucket gets or creates a token bucket for a domain
func (drl *DomainRateLimiter) getBucket(domain string) *TokenBucket {
	drl.mutex.Lock()
	defer drl.mutex.Unlock()
	
	bucket, exists := drl.buckets[domain]
	if !exists {
		rate := drl.defaultRate
		if customRate, hasCustom := drl.domainConfigs[domain]; hasCustom {
			rate = customRate
		}
		
		bucket = &TokenBucket{
			tokens:     drl.maxBurst, // Start with full bucket
			maxTokens:  drl.maxBurst,
			refillRate: rate,
			lastRefill: time.Now(),
		}
		drl.buckets[domain] = bucket
		logger.Debugf("Created rate limiter for %s: %.2f req/s, %.0f burst", domain, rate, drl.maxBurst)
	}
	
	return bucket
}

// wait blocks until a token is available
func (tb *TokenBucket) wait() error {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate
	
	// Cap at max tokens
	if tb.tokens > tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	
	tb.lastRefill = now
	
	// If we have tokens, consume one and proceed
	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return nil
	}
	
	// Calculate wait time for next token
	waitTime := time.Duration((1.0 - tb.tokens) / tb.refillRate * float64(time.Second))
	tb.mutex.Unlock() // Release lock while waiting
	
	logger.Debugf("Rate limiting: waiting %v for token", waitTime)
	time.Sleep(waitTime)
	
	tb.mutex.Lock() // Re-acquire for final token consumption
	tb.tokens = 0 // Consume the token we waited for
	
	return nil
}

// extractDomain extracts the domain from a URL
func extractDomain(targetURL string) (string, error) {
	if targetURL == "" {
		return "", fmt.Errorf("empty URL")
	}
	
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	
	if parsed.Host == "" {
		return "", fmt.Errorf("no host in URL")
	}
	
	return parsed.Host, nil
}

// GetStats returns statistics about the rate limiter
func (drl *DomainRateLimiter) GetStats() map[string]RateLimiterStats {
	drl.mutex.RLock()
	defer drl.mutex.RUnlock()
	
	stats := make(map[string]RateLimiterStats)
	for domain, bucket := range drl.buckets {
		bucket.mutex.Lock()
		stats[domain] = RateLimiterStats{
			Domain:      domain,
			Rate:        bucket.refillRate,
			MaxTokens:   bucket.maxTokens,
			CurrentTokens: bucket.tokens,
			LastRefill:  bucket.lastRefill,
		}
		bucket.mutex.Unlock()
	}
	
	return stats
}

// RateLimiterStats provides statistics about a domain's rate limiter
type RateLimiterStats struct {
	Domain        string
	Rate          float64
	MaxTokens     float64
	CurrentTokens float64
	LastRefill    time.Time
}

// ClearStats resets all rate limiter buckets (useful for testing)
func (drl *DomainRateLimiter) ClearStats() {
	drl.mutex.Lock()
	defer drl.mutex.Unlock()
	
	drl.buckets = make(map[string]*TokenBucket)
	logger.Debugf("Cleared all rate limiter buckets")
}