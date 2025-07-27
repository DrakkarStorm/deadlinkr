package internal

import (
	"sync"
	"time"
)

// CacheEntry represents a cached link check result
type CacheEntry struct {
	Status    int
	Message   string
	Timestamp time.Time
	TTL       time.Duration
}

// IsExpired checks if the cache entry has expired
func (entry *CacheEntry) IsExpired() bool {
	return time.Since(entry.Timestamp) > entry.TTL
}

// LinkCache provides intelligent caching for link check results
type LinkCache struct {
	cache      map[string]*CacheEntry
	mutex      sync.RWMutex
	defaultTTL time.Duration
	maxSize    int
	
	// Statistics
	hits   int64
	misses int64
}

// NewLinkCache creates a new link cache
func NewLinkCache(defaultTTL time.Duration, maxSize int) *LinkCache {
	return &LinkCache{
		cache:      make(map[string]*CacheEntry),
		defaultTTL: defaultTTL,
		maxSize:    maxSize,
	}
}

// Get retrieves a cached result if it exists and is not expired
func (lc *LinkCache) Get(url string) (int, string, bool) {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()
	
	entry, exists := lc.cache[url]
	if !exists {
		lc.misses++
		return 0, "", false
	}
	
	if entry.IsExpired() {
		// Don't remove here to avoid write lock, cleanup will handle it
		lc.misses++
		return 0, "", false
	}
	
	lc.hits++
	return entry.Status, entry.Message, true
}

// Set stores a result in the cache
func (lc *LinkCache) Set(url string, status int, message string) {
	lc.SetWithTTL(url, status, message, lc.defaultTTL)
}

// SetWithTTL stores a result in the cache with custom TTL
func (lc *LinkCache) SetWithTTL(url string, status int, message string, ttl time.Duration) {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	
	// If cache is full, remove expired entries first
	if len(lc.cache) >= lc.maxSize {
		lc.removeExpired()
		
		// If still full after cleanup, remove oldest entry
		if len(lc.cache) >= lc.maxSize {
			lc.removeOldest()
		}
	}
	
	lc.cache[url] = &CacheEntry{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}

// removeExpired removes all expired entries (caller must hold write lock)
func (lc *LinkCache) removeExpired() {
	for url, entry := range lc.cache {
		if entry.IsExpired() {
			delete(lc.cache, url)
		}
	}
}

// removeOldest removes the oldest entry (caller must hold write lock)
func (lc *LinkCache) removeOldest() {
	var oldestURL string
	var oldestTime time.Time
	
	for url, entry := range lc.cache {
		if oldestURL == "" || entry.Timestamp.Before(oldestTime) {
			oldestURL = url
			oldestTime = entry.Timestamp
		}
	}
	
	if oldestURL != "" {
		delete(lc.cache, oldestURL)
	}
}

// Clear removes all entries from the cache
func (lc *LinkCache) Clear() {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	
	lc.cache = make(map[string]*CacheEntry)
	lc.hits = 0
	lc.misses = 0
}

// Cleanup removes expired entries and returns the number of entries removed
func (lc *LinkCache) Cleanup() int {
	lc.mutex.Lock()
	defer lc.mutex.Unlock()
	
	initialSize := len(lc.cache)
	lc.removeExpired()
	return initialSize - len(lc.cache)
}

// Stats returns cache statistics
func (lc *LinkCache) Stats() CacheStats {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()
	
	total := lc.hits + lc.misses
	var hitRate float64
	if total > 0 {
		hitRate = float64(lc.hits) / float64(total)
	}
	
	return CacheStats{
		Hits:     lc.hits,
		Misses:   lc.misses,
		Size:     len(lc.cache),
		HitRate:  hitRate,
		MaxSize:  lc.maxSize,
	}
}

// Size returns the current number of entries in the cache
func (lc *LinkCache) Size() int {
	lc.mutex.RLock()
	defer lc.mutex.RUnlock()
	return len(lc.cache)
}

// CacheStats holds cache performance statistics
type CacheStats struct {
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	Size    int     `json:"size"`
	HitRate float64 `json:"hit_rate"`
	MaxSize int     `json:"max_size"`
}

// IntelligentTTLStrategy determines TTL based on HTTP status code
func IntelligentTTLStrategy(status int, defaultTTL time.Duration) time.Duration {
	switch {
	case status >= 200 && status < 300:
		// Success responses can be cached longer
		return defaultTTL * 2
	case status == 404 || status == 410:
		// Not found responses can be cached for a while
		return defaultTTL
	case status >= 500 && status < 600:
		// Server errors should be cached for a short time
		return defaultTTL / 4
	case status == 429:
		// Rate limited, cache very briefly
		return defaultTTL / 10
	default:
		// Other client errors
		return defaultTTL / 2
	}
}