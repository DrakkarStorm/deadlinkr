package internal

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressTracker manages progress reporting and real-time statistics
type ProgressTracker struct {
	// Core progress
	TotalTasks      int64
	CompletedTasks  int64
	ActiveTasks     int64
	ErrorCount      int64
	
	// Performance stats
	LinksChecked    int64
	CacheHits       int64
	CacheMisses     int64
	BandwidthSaved  int64 // bytes
	HeadRequests    int64
	GetRequests     int64
	
	// Timing
	StartTime       time.Time
	LastUpdate      time.Time
	
	// Control
	mutex           sync.RWMutex
	enabled         bool
	updateInterval  time.Duration
	
	// Display state
	lastLineLength  int
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(enabled bool) *ProgressTracker {
	return &ProgressTracker{
		StartTime:      time.Now(),
		LastUpdate:     time.Now(),
		enabled:        enabled,
		updateInterval: 500 * time.Millisecond, // Update every 500ms
	}
}

// SetTotal sets the total number of tasks expected
func (pt *ProgressTracker) SetTotal(total int64) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.TotalTasks = total
}

// IncrementCompleted increments completed tasks counter
func (pt *ProgressTracker) IncrementCompleted() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.CompletedTasks++
	pt.LinksChecked++
	pt.LastUpdate = time.Now()
}

// IncrementActive increments active tasks counter
func (pt *ProgressTracker) IncrementActive() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.ActiveTasks++
}

// DecrementActive decrements active tasks counter
func (pt *ProgressTracker) DecrementActive() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	if pt.ActiveTasks > 0 {
		pt.ActiveTasks--
	}
}

// IncrementError increments error counter
func (pt *ProgressTracker) IncrementError() {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.ErrorCount++
}

// UpdateCacheStats updates cache statistics
func (pt *ProgressTracker) UpdateCacheStats(hits, misses int64) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.CacheHits = hits
	pt.CacheMisses = misses
}

// UpdateBandwidthStats updates bandwidth and request type statistics
func (pt *ProgressTracker) UpdateBandwidthStats(bandwidthSaved, headReqs, getReqs int64) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()
	pt.BandwidthSaved = bandwidthSaved
	pt.HeadRequests = headReqs
	pt.GetRequests = getReqs
}

// GetStats returns current statistics (thread-safe)
func (pt *ProgressTracker) GetStats() ProgressStats {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	
	elapsed := time.Since(pt.StartTime)
	
	var linksPerSecond float64
	if elapsed.Seconds() > 0 {
		linksPerSecond = float64(pt.LinksChecked) / elapsed.Seconds()
	}
	
	var cacheHitRate float64
	totalCacheOps := pt.CacheHits + pt.CacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(pt.CacheHits) / float64(totalCacheOps) * 100
	}
	
	var progressPercent float64
	if pt.TotalTasks > 0 {
		progressPercent = float64(pt.CompletedTasks) / float64(pt.TotalTasks) * 100
	}
	
	return ProgressStats{
		TotalTasks:      pt.TotalTasks,
		CompletedTasks:  pt.CompletedTasks,
		ActiveTasks:     pt.ActiveTasks,
		ErrorCount:      pt.ErrorCount,
		ProgressPercent: progressPercent,
		LinksPerSecond:  linksPerSecond,
		LinksChecked:    pt.LinksChecked,
		CacheHitRate:    cacheHitRate,
		CacheHits:       pt.CacheHits,
		CacheMisses:     pt.CacheMisses,
		BandwidthSaved:  pt.BandwidthSaved,
		HeadRequests:    pt.HeadRequests,
		GetRequests:     pt.GetRequests,
		ElapsedTime:     elapsed,
	}
}

// ShouldUpdate checks if progress should be updated based on interval
func (pt *ProgressTracker) ShouldUpdate() bool {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return pt.enabled && time.Since(pt.LastUpdate) >= pt.updateInterval
}

// IsEnabled returns whether progress tracking is enabled
func (pt *ProgressTracker) IsEnabled() bool {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return pt.enabled
}

// RenderProgressBar renders a terminal progress bar
func (pt *ProgressTracker) RenderProgressBar() {
	if !pt.IsEnabled() {
		return
	}
	
	stats := pt.GetStats()
	
	// Clear previous line
	if pt.lastLineLength > 0 {
		fmt.Printf("\r%s\r", strings.Repeat(" ", pt.lastLineLength))
	}
	
	// Build progress bar
	barWidth := 40
	filled := int(stats.ProgressPercent / 100.0 * float64(barWidth))
	if filled > barWidth {
		filled = barWidth
	}
	
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	
	// Format bandwidth
	bandwidthStr := formatBytes(stats.BandwidthSaved)
	
	// Build status line
	line := fmt.Sprintf("[%s] %.1f%% | %d/%d links | %.1f/s | Cache: %.1f%% | Saved: %s | Active: %d | Errors: %d",
		bar,
		stats.ProgressPercent,
		stats.CompletedTasks,
		stats.TotalTasks,
		stats.LinksPerSecond,
		stats.CacheHitRate,
		bandwidthStr,
		stats.ActiveTasks,
		stats.ErrorCount,
	)
	
	fmt.Print(line)
	pt.lastLineLength = len(line)
}

// Finish completes the progress bar and prints final stats
func (pt *ProgressTracker) Finish() {
	if !pt.IsEnabled() {
		return
	}
	
	stats := pt.GetStats()
	
	// Clear progress bar
	if pt.lastLineLength > 0 {
		fmt.Printf("\r%s\r", strings.Repeat(" ", pt.lastLineLength))
	}
	
	// Print final summary
	fmt.Printf("✓ Scan completed in %v\n", stats.ElapsedTime.Round(time.Millisecond))
	fmt.Printf("  Links checked: %d (%.1f/s)\n", stats.CompletedTasks, stats.LinksPerSecond)
	
	if stats.CacheHits+stats.CacheMisses > 0 {
		fmt.Printf("  Cache efficiency: %.1f%% (%d hits, %d misses)\n", 
			stats.CacheHitRate, stats.CacheHits, stats.CacheMisses)
	}
	
	if stats.BandwidthSaved > 0 {
		fmt.Printf("  Bandwidth saved: %s\n", formatBytes(stats.BandwidthSaved))
		fmt.Printf("  Requests: %d HEAD, %d GET\n", stats.HeadRequests, stats.GetRequests)
	}
	
	if stats.ErrorCount > 0 {
		fmt.Printf("  Errors encountered: %d\n", stats.ErrorCount)
	}
	
	fmt.Println()
}

// ProgressStats holds progress statistics
type ProgressStats struct {
	TotalTasks      int64
	CompletedTasks  int64
	ActiveTasks     int64
	ErrorCount      int64
	ProgressPercent float64
	LinksPerSecond  float64
	LinksChecked    int64
	CacheHitRate    float64
	CacheHits       int64
	CacheMisses     int64
	BandwidthSaved  int64
	HeadRequests    int64
	GetRequests     int64
	ElapsedTime     time.Duration
}

// formatBytes formats byte count as human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}