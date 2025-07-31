package utils

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/PuerkitoBio/goquery"
)

// Crawl crawls the given URL and its links up to the specified depth.
// example: Crawl("https://example.com", "https://example.com", 0)
// example: Crawl("https://example.com", "https://example.com/page", 1)
func Crawl(baseURL, currentURL string, currentDepth int, concurrency int) {
	// Stop if max depth reached
	if currentDepth > model.Depth {
		return
	}

	// Check if URL already visited
	_, alreadyVisited := model.VisitedURLs.LoadOrStore(currentURL, true)
	if alreadyVisited {
		logger.Debugf("→ skip (already visited) : %s", currentURL)
		return
	}

	logger.Debugf("Crawling: %s (depth %d)", currentURL, currentDepth)

	// Increment wait group
	model.Wg.Add(1)
	// Start a new goroutine for asynchronous crawling
	go func(url string, d int) {
		// Decrement the wait group when this goroutine completes
		defer model.Wg.Done()

		// Check the links on the current page
		links := CheckLinks(baseURL, url)

		logger.Debugf("Found %d links on %s", len(links), url)

		// If the current depth is less than the maximum depth, continue crawling
		if d < model.Depth {
			// Iterate over each link found on the current page
			for _, link := range links {
				// Only recursively crawl internal links
				if !link.IsExternal {
					// Start a new goroutine for each internal link to crawl it
					Crawl(baseURL, link.TargetURL, d+1, model.Concurrency)
				}
			}
		}
	}(currentURL, currentDepth)
}

// CheckLinks checks all links on a page and returns a slice of LinkResult structs.
func CheckLinks(baseURL, pageURL string) []model.LinkResult {
	pageLinks := []model.LinkResult{}

	baseUrlParsed := parseBaseURL(baseURL)
	if baseUrlParsed == nil {
		return pageLinks
	}

	doc := fetchAndParseDocument(pageURL)
	if doc == nil {
		return pageLinks
	}

	pageLinks = extractLinks(baseUrlParsed, pageURL, doc)
	logger.Debugf("Found %d links on %s", len(pageLinks), pageURL)
	return pageLinks
}

func parseBaseURL(baseURL string) *url.URL {
	baseUrlParsed, err := url.Parse(baseURL)
	if err != nil {
		logger.Errorf("Error parsing base URL %s: %s", baseURL, err)
		return nil
	}
	if baseUrlParsed.Host == "" {
		logger.Errorf("Error parsing base URL %s: no host found", baseURL)
		return nil
	}
	return baseUrlParsed
}

func fetchAndParseDocument(pageURL string) *goquery.Document {
	retry := 3
	resp, err := FetchWithRetry(pageURL, retry)

	if err != nil {
		logger.Errorf("Failed to fetch %s after %d retries: %s", pageURL, retry, err)
		return nil
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Errorf("Error closing response body for %s: %s", pageURL, err)
		}
	}()

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Errorf("Error parsing HTML from %s: %s", pageURL, err)
		return nil
	}

	return doc
}

func extractLinks(baseUrlParsed *url.URL, pageURL string, doc *goquery.Document) []model.LinkResult {
	pageLinks := []model.LinkResult{}

	doc.Find("body a[href]").Not(model.ExcludeHtmlTags).Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			logger.Debugf("Skipping link due to missing href or #: %s", href)
			return
		}

		linkURL := resolveAndFilterURL(baseUrlParsed, pageURL, href)
		if linkURL == nil {
			logger.Debugf("Skipping link due to invalid URL resolution: %s", href)
			return
		}

		isExternal := baseUrlParsed.Hostname() != linkURL.Hostname()

		if model.OnlyInternal && isExternal {
			return
		}

		if shouldSkipURLBasedOnPattern(linkURL) {
			logger.Debugf("Skipping link due to pattern match: %s", href)
			return
		}

		status, errMsg := CheckLink(linkURL.String())

		linkResult := model.LinkResult{
			SourceURL:  pageURL,
			TargetURL:  linkURL.String(),
			Status:     status,
			Error:      errMsg,
			IsExternal: isExternal,
		}

		pageLinks = append(pageLinks, linkResult)
		addLinkResultToModel(linkResult)
	})

	return pageLinks
}

func resolveAndFilterURL(baseUrlParsed *url.URL, pageURL, href string) *url.URL {
	linkURL, err := ResolveURL(pageURL, href)
	if err != nil || ShouldSkipURL(baseUrlParsed, linkURL) {
		return nil
	}
	return linkURL
}

func shouldSkipURLBasedOnPattern(linkURL *url.URL) bool {
	if model.IncludePattern != "" {
		matched, err := regexp.MatchString(model.IncludePattern, linkURL.String())
		if err != nil || !matched {
			return true
		}
	}

	if model.ExcludePattern != "" {
		matched, err := regexp.MatchString(model.ExcludePattern, linkURL.String())
		if err == nil && matched {
			return true
		}
	}

	return false
}

func addLinkResultToModel(linkResult model.LinkResult) {
	model.ResultsMutex.Lock()
	model.Results = append(model.Results, linkResult)
	model.ResultsMutex.Unlock()
}
