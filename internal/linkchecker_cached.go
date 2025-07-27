package internal

import (
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// CachedLinkCheckerService wraps any LinkChecker with intelligent caching
type CachedLinkCheckerService struct {
	checker LinkChecker
	cache   *LinkCache
}

// NewCachedLinkCheckerService creates a new cached link checker
func NewCachedLinkCheckerService(checker LinkChecker, cacheSize int, defaultTTL time.Duration) *CachedLinkCheckerService {
	return &CachedLinkCheckerService{
		checker: checker,
		cache:   NewLinkCache(defaultTTL, cacheSize),
	}
}

// CheckLink checks a link with caching
func (clc *CachedLinkCheckerService) CheckLink(linkURL string) (int, string) {
	// Try to get from cache first
	if status, message, found := clc.cache.Get(linkURL); found {
		logger.Debugf("Cache hit for %s: %d", linkURL, status)
		return status, message
	}
	
	// Not in cache, check the link
	logger.Debugf("Cache miss for %s, checking link", linkURL)
	status, message := clc.checker.CheckLink(linkURL)
	
	// Store in cache with intelligent TTL
	ttl := IntelligentTTLStrategy(status, clc.cache.defaultTTL)
	clc.cache.SetWithTTL(linkURL, status, message, ttl)
	
	return status, message
}

// FetchWithRetry implements the LinkChecker interface
func (clc *CachedLinkCheckerService) FetchWithRetry(url string, retry int) (*model.HTTPResponse, error) {
	return clc.checker.FetchWithRetry(url, retry)
}

// GetCacheStats returns cache statistics
func (clc *CachedLinkCheckerService) GetCacheStats() CacheStats {
	return clc.cache.Stats()
}

// ClearCache clears all cached entries
func (clc *CachedLinkCheckerService) ClearCache() {
	clc.cache.Clear()
}

// CleanupCache removes expired entries
func (clc *CachedLinkCheckerService) CleanupCache() int {
	return clc.cache.Cleanup()
}

// CachedOptimizedLinkCheckerService wraps OptimizedLinkCheckerService with caching
type CachedOptimizedLinkCheckerService struct {
	*CachedLinkCheckerService
	optimizedChecker *OptimizedLinkCheckerService
}

// NewCachedOptimizedLinkCheckerService creates a cached optimized link checker
func NewCachedOptimizedLinkCheckerService(client HTTPClient, userAgent string, timeout time.Duration, requestsPerSecond, burst float64, cacheSize int, defaultTTL time.Duration) *CachedOptimizedLinkCheckerService {
	optimizedChecker := NewOptimizedLinkCheckerService(client, userAgent, timeout, requestsPerSecond, burst)
	cachedChecker := NewCachedLinkCheckerService(optimizedChecker, cacheSize, defaultTTL)
	
	return &CachedOptimizedLinkCheckerService{
		CachedLinkCheckerService: cachedChecker,
		optimizedChecker:         optimizedChecker,
	}
}

// GetOptimizationStats returns optimization statistics from the underlying checker
func (colc *CachedOptimizedLinkCheckerService) GetOptimizationStats() OptimizedLinkStats {
	return colc.optimizedChecker.GetOptimizationStats()
}

// SetDomainRateLimit sets rate limit for a specific domain
func (colc *CachedOptimizedLinkCheckerService) SetDomainRateLimit(domain string, requestsPerSecond float64) {
	colc.optimizedChecker.SetDomainRateLimit(domain, requestsPerSecond)
}

// GetRateLimiterStats returns rate limiter statistics
func (colc *CachedOptimizedLinkCheckerService) GetRateLimiterStats() map[string]RateLimiterStats {
	return colc.optimizedChecker.GetRateLimiterStats()
}

// GetCombinedStats returns both cache and optimization stats
func (colc *CachedOptimizedLinkCheckerService) GetCombinedStats() CombinedStats {
	return CombinedStats{
		Cache:        colc.GetCacheStats(),
		Optimization: colc.GetOptimizationStats(),
		RateLimit:    colc.GetRateLimiterStats(),
	}
}

// CombinedStats holds all performance statistics
type CombinedStats struct {
	Cache        CacheStats                      `json:"cache"`
	Optimization OptimizedLinkStats              `json:"optimization"`
	RateLimit    map[string]RateLimiterStats     `json:"rate_limit"`
}