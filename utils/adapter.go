package utils

import (
	"time"

	"github.com/DrakkarStorm/deadlinkr/internal"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// CrawlWithServices is the new implementation using the service architecture
func CrawlWithServices(baseURL, currentURL string, currentDepth int) error {
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

	// Create crawler service
	crawler := factory.CreateCrawlerService(
		config,
		model.UserAgent,
		time.Duration(model.Timeout)*time.Second,
		ClientHTTP, // Pass the existing HTTP client
	)

	// Start crawling
	err := crawler.StartCrawl(baseURL, currentURL, currentDepth)
	if err != nil {
		return err
	}

	// Wait for completion
	crawler.Wait()

	// Update global results for backward compatibility
	results := crawler.GetResults()
	model.ResultsMutex.Lock()
	model.Results = append(model.Results, results...)
	model.ResultsMutex.Unlock()

	return nil
}

// CheckLinksWithServices checks links on a page using the new service architecture
func CheckLinksWithServices(baseURL, pageURL string) ([]model.LinkResult, error) {
	factory := internal.NewServiceFactory()
	
	config := factory.CreateCrawlConfigFromParams(
		1, // depth 1 for single page
		model.Concurrency,
		model.OnlyInternal,
		model.IncludePattern,
		model.ExcludePattern,
		model.ExcludeHtmlTags,
	)

	crawler := factory.CreateCrawlerService(
		config,
		model.UserAgent,
		time.Duration(model.Timeout)*time.Second,
		ClientHTTP, // Pass the existing HTTP client
	)

	err := crawler.Crawl(baseURL, pageURL, 0)
	if err != nil {
		return nil, err
	}

	return crawler.GetResults(), nil
}