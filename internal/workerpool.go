package internal

import (
	"context"
	"sync"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// Job represents a work item for the worker pool
type Job struct {
	BaseURL      string
	TargetURL    string
	CurrentDepth int
	Callback     func([]model.LinkResult, error)
}

// WorkerPool manages a pool of workers for concurrent link checking
type WorkerPool struct {
	workers         int
	jobQueue        chan Job
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	crawler         *CrawlerService
	stats           *PoolStats
	progressTracker *ProgressTracker // Optional progress tracker
}

// PoolStats tracks worker pool statistics
type PoolStats struct {
	JobsQueued    int64
	JobsCompleted int64
	JobsActive    int64
	mutex         sync.RWMutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, crawler *CrawlerService) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		workers:         workers,
		jobQueue:        make(chan Job, workers*2), // Buffer for smooth flow
		ctx:             ctx,
		cancel:          cancel,
		crawler:         crawler,
		stats:           &PoolStats{},
		progressTracker: nil, // Will be set by crawler if needed
	}
}

// Start initializes and starts the worker pool
func (wp *WorkerPool) Start() {
	logger.Debugf("Starting worker pool with %d workers", wp.workers)
	
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop gracefully shuts down the worker pool
func (wp *WorkerPool) Stop() {
	logger.Debugf("Stopping worker pool...")
	close(wp.jobQueue)
	wp.wg.Wait()
	wp.cancel()
	logger.Debugf("Worker pool stopped")
}

// ForceStop immediately stops the worker pool without waiting
func (wp *WorkerPool) ForceStop() {
	logger.Debugf("Force stopping worker pool...")
	
	// Cancel context immediately
	wp.cancel()
	
	// Try to close job queue safely
	select {
	case <-wp.jobQueue:
		// Queue already closed or empty
	default:
		close(wp.jobQueue)
	}
	
	// Don't wait for workers - they should stop quickly due to cancelled context
	logger.Debugf("Worker pool force stopped")
}

// Submit adds a job to the worker pool
func (wp *WorkerPool) Submit(job Job) bool {
	select {
	case wp.jobQueue <- job:
		wp.incrementJobsQueued()
		return true
	case <-wp.ctx.Done():
		return false
	default:
		// Queue is full, could implement backpressure here
		logger.Warnf("Worker pool queue is full, dropping job for %s", job.TargetURL)
		return false
	}
}

// worker is the main worker goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	logger.Debugf("Worker %d started", id)
	
	for {
		select {
		case job, ok := <-wp.jobQueue:
			if !ok {
				logger.Debugf("Worker %d stopping", id)
				return
			}
			
			wp.incrementJobsActive()
			wp.processJob(id, job)
			wp.decrementJobsActive()
			wp.incrementJobsCompleted()
			
			// Update progress tracker if available
			if wp.progressTracker != nil {
				wp.progressTracker.IncrementCompleted()
			}
			
		case <-wp.ctx.Done():
			logger.Debugf("Worker %d cancelled", id)
			return
		}
	}
}

// processJob processes a single job
func (wp *WorkerPool) processJob(workerID int, job Job) {
	start := time.Now()
	logger.Debugf("Worker %d processing %s (depth %d)", workerID, job.TargetURL, job.CurrentDepth)
	
	// Check if we should skip this URL (already visited)
	if wp.crawler.resultCollector.IsVisited(job.TargetURL) {
		logger.Debugf("Worker %d skipping already visited: %s", workerID, job.TargetURL)
		if job.Callback != nil {
			job.Callback(nil, nil)
		}
		return
	}
	
	// Mark as visited first
	wp.crawler.resultCollector.MarkVisited(job.TargetURL)
	
	// Validate base URL
	baseUrlParsed, err := wp.crawler.urlProcessor.ValidateURL(job.BaseURL)
	if err != nil {
		logger.Errorf("Worker %d: Invalid base URL %s: %s", workerID, job.BaseURL, err)
		if job.Callback != nil {
			job.Callback(nil, err)
		}
		return
	}
	
	// Parse the page and extract links
	doc, err := wp.crawler.pageParser.ParsePage(job.TargetURL)
	if err != nil {
		logger.Errorf("Worker %d: Error parsing %s: %s", workerID, job.TargetURL, err)
		if job.Callback != nil {
			job.Callback(nil, err)
		}
		return
	}
	
	if doc == nil {
		logger.Debugf("Worker %d: No HTML content for %s", workerID, job.TargetURL)
		if job.Callback != nil {
			job.Callback(nil, nil)
		}
		return
	}
	
	// Extract links from the page
	links := wp.crawler.pageParser.ExtractLinks(baseUrlParsed, job.TargetURL, doc)
	
	// Add results to collector
	for _, link := range links {
		wp.crawler.resultCollector.AddResult(link)
	}
	
	duration := time.Since(start)
	logger.Debugf("Worker %d completed %s in %v (found %d links)", 
		workerID, job.TargetURL, duration, len(links))
	
	// Schedule internal links for further crawling if within depth limit
	// Note: For now, we'll handle this in the caller to avoid infinite recursion
	// TODO: Implement proper job scheduling with depth tracking
	
	// Call callback with results
	if job.Callback != nil {
		job.Callback(links, nil)
	}
}

// GetStats returns current pool statistics
func (wp *WorkerPool) GetStats() PoolStats {
	wp.stats.mutex.RLock()
	defer wp.stats.mutex.RUnlock()
	return PoolStats{
		JobsQueued:    wp.stats.JobsQueued,
		JobsCompleted: wp.stats.JobsCompleted,
		JobsActive:    wp.stats.JobsActive,
	}
}

// SetProgressTracker sets the progress tracker for this worker pool
func (wp *WorkerPool) SetProgressTracker(tracker *ProgressTracker) {
	wp.progressTracker = tracker
}

// Helper methods for stats
func (wp *WorkerPool) incrementJobsQueued() {
	wp.stats.mutex.Lock()
	wp.stats.JobsQueued++
	wp.stats.mutex.Unlock()
}

func (wp *WorkerPool) incrementJobsCompleted() {
	wp.stats.mutex.Lock()
	wp.stats.JobsCompleted++
	wp.stats.mutex.Unlock()
}

func (wp *WorkerPool) incrementJobsActive() {
	wp.stats.mutex.Lock()
	wp.stats.JobsActive++
	wp.stats.mutex.Unlock()
}

func (wp *WorkerPool) decrementJobsActive() {
	wp.stats.mutex.Lock()
	wp.stats.JobsActive--
	wp.stats.mutex.Unlock()
}

// IsIdle returns true if no jobs are active or queued
func (wp *WorkerPool) IsIdle() bool {
	stats := wp.GetStats()
	return stats.JobsActive == 0 && len(wp.jobQueue) == 0
}