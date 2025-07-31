package internal

import (
	"net/http"
	"net/url"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/PuerkitoBio/goquery"
)

// HTTPClient interface to support both regular and authenticated clients
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// LinkChecker interface defines methods for checking individual links
type LinkChecker interface {
	CheckLink(linkURL string) (int, string)
	FetchWithRetry(url string, retry int) (*model.HTTPResponse, error)
}

// OptimizedLinkChecker extends LinkChecker with optimization features
type OptimizedLinkChecker interface {
	LinkChecker
	GetOptimizationStats() OptimizedLinkStats
	SetDomainRateLimit(domain string, requestsPerSecond float64)
	GetRateLimiterStats() map[string]RateLimiterStats
}

// PageParser interface defines methods for parsing web pages
type PageParser interface {
	ParsePage(pageURL string) (*goquery.Document, error)
	ExtractLinks(baseURL *url.URL, pageURL string, doc *goquery.Document) []model.LinkResult
}

// URLProcessor interface defines methods for URL processing
type URLProcessor interface {
	ResolveURL(pageURL, href string) (*url.URL, error)
	ShouldSkipURL(baseURL, linkURL *url.URL) bool
	ValidateURL(baseURL string) (*url.URL, error)
}

// Crawler interface defines methods for crawling websites
type Crawler interface {
	Crawl(baseURL, currentURL string, currentDepth int) error
	SetConfig(config *CrawlConfig)
}

// ResultCollector interface defines methods for collecting and managing results
type ResultCollector interface {
	AddResult(result model.LinkResult)
	GetResults() []model.LinkResult
	CountBrokenLinks() int
	IsVisited(url string) bool
	MarkVisited(url string)
	Clear()
}

// CrawlConfig holds configuration for crawling
type CrawlConfig struct {
	MaxDepth        int
	Concurrency     int
	OnlyInternal    bool
	IncludePattern  string
	ExcludePattern  string
	ExcludeHtmlTags string
}