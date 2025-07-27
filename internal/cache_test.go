package internal

import (
	"testing"
	"time"
)

func TestLinkCache_SetAndGet(t *testing.T) {
	cache := NewLinkCache(1*time.Hour, 10)
	
	// Test set and get
	cache.Set("http://example.com", 200, "OK")
	status, message, found := cache.Get("http://example.com")
	
	if !found {
		t.Error("Expected to find cached entry")
	}
	if status != 200 {
		t.Errorf("Expected status 200, got %d", status)
	}
	if message != "OK" {
		t.Errorf("Expected message 'OK', got %s", message)
	}
}

func TestLinkCache_Expiration(t *testing.T) {
	cache := NewLinkCache(50*time.Millisecond, 10)
	
	// Set with short TTL
	cache.SetWithTTL("http://example.com", 200, "OK", 50*time.Millisecond)
	
	// Should be found immediately
	_, _, found := cache.Get("http://example.com")
	if !found {
		t.Error("Expected to find cached entry")
	}
	
	// Wait for expiration
	time.Sleep(100 * time.Millisecond)
	
	// Should not be found after expiration
	_, _, found = cache.Get("http://example.com")
	if found {
		t.Error("Expected cached entry to be expired")
	}
}

func TestLinkCache_MaxSize(t *testing.T) {
	cache := NewLinkCache(1*time.Hour, 3)
	
	// Fill cache to capacity
	cache.Set("http://example1.com", 200, "OK")
	cache.Set("http://example2.com", 200, "OK")
	cache.Set("http://example3.com", 200, "OK")
	
	if cache.Size() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Size())
	}
	
	// Add one more - should evict oldest
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	cache.Set("http://example4.com", 200, "OK")
	
	if cache.Size() != 3 {
		t.Errorf("Expected cache size to remain 3, got %d", cache.Size())
	}
	
	// First entry should be evicted
	_, _, found := cache.Get("http://example1.com")
	if found {
		t.Error("Expected first entry to be evicted")
	}
	
	// Last entry should be present
	_, _, found = cache.Get("http://example4.com")
	if !found {
		t.Error("Expected last entry to be present")
	}
}

func TestLinkCache_Stats(t *testing.T) {
	cache := NewLinkCache(1*time.Hour, 10)
	
	// Initial stats
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Expected initial stats to be zero")
	}
	
	// Add entry and test hit
	cache.Set("http://example.com", 200, "OK")
	cache.Get("http://example.com")
	
	stats = cache.Stats()
	if stats.Hits != 1 {
		t.Errorf("Expected 1 hit, got %d", stats.Hits)
	}
	
	// Test miss
	cache.Get("http://notfound.com")
	
	stats = cache.Stats()
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.HitRate != 0.5 {
		t.Errorf("Expected hit rate 0.5, got %f", stats.HitRate)
	}
}

func TestLinkCache_Cleanup(t *testing.T) {
	cache := NewLinkCache(1*time.Hour, 10)
	
	// Add entries with different TTLs
	cache.SetWithTTL("http://example1.com", 200, "OK", 50*time.Millisecond)
	cache.SetWithTTL("http://example2.com", 200, "OK", 1*time.Hour)
	
	if cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Size())
	}
	
	// Wait for first entry to expire
	time.Sleep(100 * time.Millisecond)
	
	// Cleanup should remove expired entry
	removed := cache.Cleanup()
	if removed != 1 {
		t.Errorf("Expected 1 entry to be removed, got %d", removed)
	}
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1 after cleanup, got %d", cache.Size())
	}
}

func TestLinkCache_Clear(t *testing.T) {
	cache := NewLinkCache(1*time.Hour, 10)
	
	// Add entries
	cache.Set("http://example1.com", 200, "OK")
	cache.Set("http://example2.com", 404, "Not Found")
	
	if cache.Size() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Size())
	}
	
	// Clear cache
	cache.Clear()
	
	if cache.Size() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Size())
	}
	
	// Stats should also be reset
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Error("Expected stats to be reset after clear")
	}
}

func TestIntelligentTTLStrategy(t *testing.T) {
	baseTTL := 1 * time.Hour
	
	tests := []struct {
		status   int
		expected time.Duration
	}{
		{200, baseTTL * 2},    // Success
		{404, baseTTL},        // Not found
		{500, baseTTL / 4},    // Server error
		{429, baseTTL / 10},   // Rate limited
		{403, baseTTL / 2},    // Other client error
	}
	
	for _, test := range tests {
		result := IntelligentTTLStrategy(test.status, baseTTL)
		if result != test.expected {
			t.Errorf("For status %d, expected TTL %v, got %v", test.status, test.expected, result)
		}
	}
}