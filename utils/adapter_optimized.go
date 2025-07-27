package utils

import (
	"time"

	"github.com/DrakkarStorm/deadlinkr/internal"
	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// CrawlWithOptimizedServices is the optimized implementation using worker pools
func CrawlWithOptimizedServices(baseURL, currentURL string, currentDepth int) error {
	factory := internal.NewServiceFactory()
	
	// Create config from global model state
	config := factory.CreateCrawlConfigFromParams(
		model.Depth,
		model.Concurrency,
		model.OnlyInternal,
		model.IncludePattern,
		model.ExcludePattern,
		model.ExcludeHtmlTags,
	)

	// Create optimized crawler service with rate limiting, HEAD optimization, and caching
	var crawler *internal.OptimizedCrawlerService
	if model.CacheEnabled && model.OptimizeWithHeadRequests {
		crawler = factory.CreateCachedOptimizedCrawlerService(
			config,
			model.UserAgent,
			time.Duration(model.Timeout)*time.Second,
			ClientHTTP, // Pass the existing HTTP client
			model.RateLimitRequestsPerSecond,
			model.RateLimitBurst,
			model.CacheSize,
			time.Duration(model.CacheTTLMinutes)*time.Minute,
		)
	} else if model.OptimizeWithHeadRequests {
		crawler = factory.CreateOptimizedCrawlerServiceWithHeadOptimization(
			config,
			model.UserAgent,
			time.Duration(model.Timeout)*time.Second,
			ClientHTTP, // Pass the existing HTTP client
			model.RateLimitRequestsPerSecond,
			model.RateLimitBurst,
		)
	} else {
		crawler = factory.CreateOptimizedCrawlerServiceWithRateLimit(
			config,
			model.UserAgent,
			time.Duration(model.Timeout)*time.Second,
			ClientHTTP, // Pass the existing HTTP client
			model.RateLimitRequestsPerSecond,
			model.RateLimitBurst,
		)
	}

	// Ensure cleanup
	defer crawler.Stop()

	// Start crawling
	err := crawler.StartCrawl(baseURL, currentURL, currentDepth)
	if err != nil {
		return err
	}

	// Wait for completion
	crawler.Wait()

	// Log some stats
	stats := crawler.GetStats()
	logger.Infof("Worker pool stats - Queued: %d, Completed: %d, Active: %d", 
		stats.JobsQueued, stats.JobsCompleted, stats.JobsActive)

	// Update global results for backward compatibility
	results := crawler.GetResults()
	model.ResultsMutex.Lock()
	model.Results = append(model.Results, results...)
	model.ResultsMutex.Unlock()

	return nil
}

// CheckLinksWithOptimizedServices checks links on a page using the optimized architecture
func CheckLinksWithOptimizedServices(baseURL, pageURL string) ([]model.LinkResult, error) {
	factory := internal.NewServiceFactory()
	
	config := factory.CreateCrawlConfigFromParams(
		1, // depth 1 for single page
		model.Concurrency,
		model.OnlyInternal,
		model.IncludePattern,
		model.ExcludePattern,
		model.ExcludeHtmlTags,
	)

	var crawler *internal.OptimizedCrawlerService
	if model.CacheEnabled && model.OptimizeWithHeadRequests {
		crawler = factory.CreateCachedOptimizedCrawlerService(
			config,
			model.UserAgent,
			time.Duration(model.Timeout)*time.Second,
			ClientHTTP, // Pass the existing HTTP client
			model.RateLimitRequestsPerSecond,
			model.RateLimitBurst,
			model.CacheSize,
			time.Duration(model.CacheTTLMinutes)*time.Minute,
		)
	} else if model.OptimizeWithHeadRequests {
		crawler = factory.CreateOptimizedCrawlerServiceWithHeadOptimization(
			config,
			model.UserAgent,
			time.Duration(model.Timeout)*time.Second,
			ClientHTTP, // Pass the existing HTTP client
			model.RateLimitRequestsPerSecond,
			model.RateLimitBurst,
		)
	} else {
		crawler = factory.CreateOptimizedCrawlerServiceWithRateLimit(
			config,
			model.UserAgent,
			time.Duration(model.Timeout)*time.Second,
			ClientHTTP, // Pass the existing HTTP client
			model.RateLimitRequestsPerSecond,
			model.RateLimitBurst,
		)
	}

	// Ensure cleanup
	defer crawler.Stop()

	err := crawler.StartCrawl(baseURL, pageURL, 0)
	if err != nil {
		return nil, err
	}

	crawler.Wait()

	return crawler.GetResults(), nil
}