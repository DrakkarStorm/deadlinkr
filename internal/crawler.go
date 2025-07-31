package internal

import (
	"sync"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// CrawlerService implements the Crawler interface
type CrawlerService struct {
	pageParser      PageParser
	urlProcessor    URLProcessor
	resultCollector ResultCollector
	config          *CrawlConfig
	wg              *sync.WaitGroup
}

// NewCrawlerService creates a new CrawlerService
func NewCrawlerService(pageParser PageParser, urlProcessor URLProcessor, resultCollector ResultCollector, config *CrawlConfig) *CrawlerService {
	return &CrawlerService{
		pageParser:      pageParser,
		urlProcessor:    urlProcessor,
		resultCollector: resultCollector,
		config:          config,
		wg:              &sync.WaitGroup{},
	}
}

// Crawl crawls the given URL and its links up to the specified depth
func (c *CrawlerService) Crawl(baseURL, currentURL string, currentDepth int) error {
	// Stop if max depth reached
	if currentDepth > c.config.MaxDepth {
		return nil
	}

	// Check if URL already visited
	if c.resultCollector.IsVisited(currentURL) {
		logger.Debugf("→ skip (already visited) : %s", currentURL)
		return nil
	}

	// Mark as visited
	c.resultCollector.MarkVisited(currentURL)

	logger.Debugf("Crawling: %s (depth %d)", currentURL, currentDepth)

	// Validate base URL
	baseUrlParsed, err := c.urlProcessor.ValidateURL(baseURL)
	if err != nil {
		return err
	}

	// Parse the page and extract links
	doc, err := c.pageParser.ParsePage(currentURL)
	if err != nil {
		return err
	}
	if doc == nil {
		return nil
	}

	links := c.pageParser.ExtractLinks(baseUrlParsed, currentURL, doc)
	logger.Debugf("Found %d links on %s", len(links), currentURL)

	// Add results to collector
	for _, link := range links {
		c.resultCollector.AddResult(link)
	}

	// If the current depth is less than the maximum depth, continue crawling
	if currentDepth < c.config.MaxDepth {
		// Iterate over each link found on the current page
		for _, link := range links {
			// Only recursively crawl internal links
			if !link.IsExternal {
				// Start a new goroutine for each internal link to crawl it
				c.wg.Add(1)
				go func(targetURL string) {
					defer c.wg.Done()
					if err := c.Crawl(baseURL, targetURL, currentDepth+1); err != nil {
						logger.Errorf("Error crawling %s: %s", targetURL, err)
					}
				}(link.TargetURL)
			}
		}
	}

	return nil
}

// SetConfig updates the crawler configuration
func (c *CrawlerService) SetConfig(config *CrawlConfig) {
	c.config = config
}

// Wait waits for all crawling goroutines to complete
func (c *CrawlerService) Wait() {
	c.wg.Wait()
}

// StartCrawl starts the initial crawl with proper waitgroup management
func (c *CrawlerService) StartCrawl(baseURL, currentURL string, currentDepth int) error {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.Crawl(baseURL, currentURL, currentDepth); err != nil {
			logger.Errorf("Error in crawl: %s", err)
		}
	}()
	return nil
}

// GetResults returns the collected results
func (c *CrawlerService) GetResults() []model.LinkResult {
	return c.resultCollector.GetResults()
}

// CountBrokenLinks returns the count of broken links
func (c *CrawlerService) CountBrokenLinks() int {
	return c.resultCollector.CountBrokenLinks()
}