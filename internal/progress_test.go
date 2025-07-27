package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgressTracker_BasicFunctionality(t *testing.T) {
	tracker := NewProgressTracker(true)
	
	// Test initial state
	assert.True(t, tracker.IsEnabled())
	stats := tracker.GetStats()
	assert.Equal(t, int64(0), stats.TotalTasks)
	assert.Equal(t, int64(0), stats.CompletedTasks)
	
	// Test setting total and incrementing completed
	tracker.SetTotal(10)
	tracker.IncrementCompleted()
	tracker.IncrementCompleted()
	
	stats = tracker.GetStats()
	assert.Equal(t, int64(10), stats.TotalTasks)
	assert.Equal(t, int64(2), stats.CompletedTasks)
	assert.Equal(t, float64(20), stats.ProgressPercent)
}

func TestProgressTracker_CacheStats(t *testing.T) {
	tracker := NewProgressTracker(true)
	
	// Update cache stats
	tracker.UpdateCacheStats(80, 20)
	
	stats := tracker.GetStats()
	assert.Equal(t, float64(80), stats.CacheHitRate) // 80/(80+20) * 100 = 80%
}

func TestProgressTracker_BandwidthStats(t *testing.T) {
	tracker := NewProgressTracker(true)
	
	// Update bandwidth stats
	tracker.UpdateBandwidthStats(1024*1024, 50, 30) // 1MB saved, 50 HEAD, 30 GET
	
	stats := tracker.GetStats()
	assert.Equal(t, int64(1024*1024), stats.BandwidthSaved)
	assert.Equal(t, int64(50), stats.HeadRequests)
	assert.Equal(t, int64(30), stats.GetRequests)
}

func TestProgressTracker_LinksPerSecond(t *testing.T) {
	tracker := NewProgressTracker(true)
	
	// Simulate some time passing and links being checked
	time.Sleep(100 * time.Millisecond)
	tracker.IncrementCompleted()
	tracker.IncrementCompleted()
	tracker.IncrementCompleted()
	
	stats := tracker.GetStats()
	assert.Greater(t, stats.LinksPerSecond, float64(0))
	assert.Equal(t, int64(3), stats.LinksChecked)
}

func TestProgressTracker_ActiveTasksManagement(t *testing.T) {
	tracker := NewProgressTracker(true)
	
	// Test active task management
	tracker.IncrementActive()
	tracker.IncrementActive()
	
	stats := tracker.GetStats()
	assert.Equal(t, int64(2), stats.ActiveTasks)
	
	tracker.DecrementActive()
	stats = tracker.GetStats()
	assert.Equal(t, int64(1), stats.ActiveTasks)
	
	// Test that decrement doesn't go below zero
	tracker.DecrementActive()
	tracker.DecrementActive()
	stats = tracker.GetStats()
	assert.Equal(t, int64(0), stats.ActiveTasks)
}

func TestProgressTracker_ErrorCount(t *testing.T) {
	tracker := NewProgressTracker(true)
	
	tracker.IncrementError()
	tracker.IncrementError()
	
	stats := tracker.GetStats()
	assert.Equal(t, int64(2), stats.ErrorCount)
}

func TestProgressTracker_Disabled(t *testing.T) {
	tracker := NewProgressTracker(false)
	
	assert.False(t, tracker.IsEnabled())
	assert.False(t, tracker.ShouldUpdate())
}

func TestProgressTracker_UpdateInterval(t *testing.T) {
	tracker := NewProgressTracker(true)
	
	// Should not update immediately (LastUpdate was just set)
	assert.False(t, tracker.ShouldUpdate())
	
	// Manually set LastUpdate to past to simulate interval passage
	tracker.mutex.Lock()
	tracker.LastUpdate = time.Now().Add(-600 * time.Millisecond)
	tracker.mutex.Unlock()
	
	// Should now be ready for update
	assert.True(t, tracker.ShouldUpdate())
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}
	
	for _, test := range tests {
		result := formatBytes(test.bytes)
		assert.Equal(t, test.expected, result, "Failed for %d bytes", test.bytes)
	}
}