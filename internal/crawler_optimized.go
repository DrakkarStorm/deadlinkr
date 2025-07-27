package internal

import (
	"fmt"
	"sync"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// OptimizedCrawlerService implements the Crawler interface with a worker pool
type OptimizedCrawlerService struct {
	pageParser       PageParser
	urlProcessor     URLProcessor
	resultCollector  ResultCollector
	config           *CrawlConfig
	workerPool       *WorkerPool
	activeJobs       sync.WaitGroup
	started          bool
	progressTracker  *ProgressTracker
	shutdownManager  *ShutdownManager
}

// NewOptimizedCrawlerService creates a new optimized crawler service
func NewOptimizedCrawlerService(pageParser PageParser, urlProcessor URLProcessor, resultCollector ResultCollector, config *CrawlConfig) *OptimizedCrawlerService {
	// Progress tracking enabled if not in quiet mode
	progressEnabled := !model.Quiet
	
	// Create shutdown manager
	shutdownManager := NewShutdownManager()
	
	crawler := &OptimizedCrawlerService{
		pageParser:       pageParser,
		urlProcessor:     urlProcessor,
		resultCollector:  resultCollector,
		config:           config,
		started:          false,
		progressTracker:  NewProgressTracker(progressEnabled),
		shutdownManager:  shutdownManager,
	}
	
	// Create worker pool - use concurrency setting as worker count
	crawler.workerPool = NewWorkerPool(config.Concurrency, &CrawlerService{
		pageParser:      pageParser,
		urlProcessor:    urlProcessor,
		resultCollector: resultCollector,
		config:          config,
	})
	
	// Connect progress tracker to worker pool
	crawler.workerPool.SetProgressTracker(crawler.progressTracker)
	
	// Register shutdown hooks
	crawler.registerShutdownHooks()
	
	return crawler
}

// Crawl starts crawling using the worker pool
func (c *OptimizedCrawlerService) Crawl(baseURL, currentURL string, currentDepth int) error {
	if !c.started {
		c.workerPool.Start()
		c.started = true
	}
	
	// Create initial job
	job := Job{
		BaseURL:      baseURL,
		TargetURL:    currentURL,
		CurrentDepth: currentDepth,
		Callback: func(results []model.LinkResult, err error) {
			c.activeJobs.Done() // Signal completion
		},
	}
	
	// Submit initial job
	c.activeJobs.Add(1)
	if !c.workerPool.Submit(job) {
		c.activeJobs.Done()
		logger.Errorf("Failed to submit initial job for %s", currentURL)
		return fmt.Errorf("failed to submit initial job for %s", currentURL)
	}
	
	return nil
}

// Wait waits for all crawling to complete
func (c *OptimizedCrawlerService) Wait() {
	// Start shutdown signal monitoring
	go c.shutdownManager.WaitForShutdown()
	
	// Start progress updates if enabled
	if c.progressTracker.IsEnabled() {
		go c.runProgressUpdates()
	}
	
	// Wait for all jobs to complete or shutdown signal
	done := make(chan bool, 1)
	go func() {
		c.activeJobs.Wait()
		
		// Wait for worker pool to become idle
		for !c.workerPool.IsIdle() {
			select {
			case <-c.shutdownManager.Context().Done():
				logger.Infof("Shutdown signal received, stopping crawler")
				c.forceStop()
				done <- true
				return
			case <-time.After(100 * time.Millisecond):
				c.updateProgressStats()
			}
		}
		done <- true
	}()
	
	// Wait for completion or shutdown
	select {
	case <-done:
		logger.Debugf("Crawling completed normally")
	case <-c.shutdownManager.Context().Done():
		logger.Infof("Crawling interrupted by shutdown signal")
		c.forceStop()
	}
	
	// Finish progress tracking
	c.progressTracker.Finish()
	
	// Stop the worker pool
	if c.started {
		c.workerPool.Stop()
		c.started = false
	}
	
	// Wait for shutdown to complete if it was initiated
	if c.shutdownManager.IsShuttingDown() {
		c.shutdownManager.WaitForCompletion()
	}
}

// SetConfig updates the crawler configuration
func (c *OptimizedCrawlerService) SetConfig(config *CrawlConfig) {
	c.config = config
}

// StartCrawl starts the initial crawl with proper management
func (c *OptimizedCrawlerService) StartCrawl(baseURL, currentURL string, currentDepth int) error {
	return c.Crawl(baseURL, currentURL, currentDepth)
}

// GetResults returns the collected results
func (c *OptimizedCrawlerService) GetResults() []model.LinkResult {
	return c.resultCollector.GetResults()
}

// CountBrokenLinks returns the count of broken links
func (c *OptimizedCrawlerService) CountBrokenLinks() int {
	return c.resultCollector.CountBrokenLinks()
}

// GetStats returns worker pool statistics
func (c *OptimizedCrawlerService) GetStats() PoolStats {
	return c.workerPool.GetStats()
}

// Stop gracefully stops the crawler
func (c *OptimizedCrawlerService) Stop() {
	if c.started {
		c.workerPool.Stop()
		c.started = false
	}
	
	// Cleanup shutdown manager
	c.shutdownManager.Cleanup()
}

// forceStop immediately stops the crawler (used during shutdown)
func (c *OptimizedCrawlerService) forceStop() {
	logger.Infof("Force stopping crawler...")
	
	if c.started {
		c.workerPool.ForceStop() // Will implement this method
		c.started = false
	}
}

// registerShutdownHooks registers cleanup functions for graceful shutdown
func (c *OptimizedCrawlerService) registerShutdownHooks() {
	// Register worker pool cleanup
	c.shutdownManager.AddShutdownHook(func() error {
		logger.Debugf("Shutdown hook: stopping worker pool")
		if c.started {
			c.workerPool.Stop()
			c.started = false
		}
		return nil
	})
	
	// Register progress tracker cleanup
	c.shutdownManager.AddShutdownHook(func() error {
		logger.Debugf("Shutdown hook: finishing progress tracker")
		if c.progressTracker.IsEnabled() {
			c.progressTracker.Finish()
		}
		return nil
	})
	
	// Register result collection finalization
	c.shutdownManager.AddShutdownHook(func() error {
		logger.Debugf("Shutdown hook: finalizing results")
		results := c.resultCollector.GetResults()
		logger.Infof("Final results: %d links checked, %d broken", 
			len(results), c.resultCollector.CountBrokenLinks())
		return nil
	})
}

// runProgressUpdates runs the progress update loop
func (c *OptimizedCrawlerService) runProgressUpdates() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if c.workerPool.IsIdle() && c.activeJobs == (sync.WaitGroup{}) {
				return // Crawling is done
			}
			c.updateProgressStats()
			c.progressTracker.RenderProgressBar()
		}
	}
}

// updateProgressStats updates progress tracker with current statistics
func (c *OptimizedCrawlerService) updateProgressStats() {
	// Get worker pool stats
	poolStats := c.workerPool.GetStats()
	
	// Update basic progress
	c.progressTracker.SetTotal(poolStats.JobsQueued)
	// Note: CompletedTasks will be updated by individual job completions
	
	// Update active tasks from pool
	c.progressTracker.mutex.Lock()
	c.progressTracker.ActiveTasks = poolStats.JobsActive
	c.progressTracker.mutex.Unlock()
	
	// Try to get cache and optimization stats if available
	c.updatePerformanceStats()
}

// updatePerformanceStats updates performance-related statistics
func (c *OptimizedCrawlerService) updatePerformanceStats() {
	// Try to extract stats from page parser if it supports optimization
	if optimizedChecker, ok := c.pageParser.(*PageParserService); ok {
		if linkChecker := optimizedChecker.LinkChecker; linkChecker != nil {
			// Check if it's a cached optimized checker
			if cachedChecker, ok := linkChecker.(*CachedOptimizedLinkCheckerService); ok {
				// Get cache stats
				cacheStats := cachedChecker.GetCacheStats()
				c.progressTracker.UpdateCacheStats(cacheStats.Hits, cacheStats.Misses)
				
				// Get optimization stats
				optimizationStats := cachedChecker.GetOptimizationStats()
				c.progressTracker.UpdateBandwidthStats(
					optimizationStats.BytesSaved,
					optimizationStats.HeadRequestsUsed,
					optimizationStats.GetRequestsUsed,
				)
			} else if optimizedChecker, ok := linkChecker.(*OptimizedLinkCheckerService); ok {
				// Get optimization stats only
				optimizationStats := optimizedChecker.GetOptimizationStats()
				c.progressTracker.UpdateBandwidthStats(
					optimizationStats.BytesSaved,
					optimizationStats.HeadRequestsUsed,
					optimizationStats.GetRequestsUsed,
				)
			}
		}
	}
}