package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/PuerkitoBio/goquery"
	// Import the main package
)

func Crawl(baseURL, currentURL string, currentDepth int) {
	// Stop if max depth reached
	if currentDepth > model.Depth {
		return
	}

	// Check if URL already visited
	_, alreadyVisited := model.VisitedURLs.LoadOrStore(currentURL, true)
	if alreadyVisited {
		return
	}

	fmt.Printf("Crawling: %s (depth %d)\n", currentURL, currentDepth)

	// Increment wait group
	model.Wg.Add(1)
	// Start a new goroutine for asynchronous crawling
	go func(url string, d int) {
		// Decrement the wait group when this goroutine completes
		defer model.Wg.Done()

		// Check the links on the current page
		links := CheckLinks(baseURL, url)

		// If the current depth is less than the maximum depth, continue crawling
		if d < model.Depth {
			// Iterate over each link found on the current page
			for _, link := range links {
				// Only recursively crawl internal links
				if !link.IsExternal {
					// Start a new goroutine for each internal link to crawl it
					go Crawl(baseURL, link.TargetURL, d+1)
				}
			}
		}
	}(currentURL, currentDepth)
}

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
	fmt.Printf("Found %d links on %s\n", len(pageLinks), pageURL)
	return pageLinks
}

func parseBaseURL(baseURL string) *url.URL {
	baseUrlParsed, err := url.Parse(baseURL)
	if err != nil {
		fmt.Printf("Error parsing base URL %s: %s\n", baseURL, err)
		return nil
	}
	if baseUrlParsed.Host == "" {
		fmt.Printf("Error parsing base URL %s: no host found\n", baseURL)
		return nil
	}
	return baseUrlParsed
}

func fetchAndParseDocument(pageURL string) *goquery.Document {
	client := &http.Client{
		Timeout: time.Duration(model.Timeout) * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		fmt.Printf("Error creating request for %s: %s\n", pageURL, err)
		return nil
	}

	req.Header.Set("User-Agent", model.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching %s: %s\n", pageURL, err)
		return nil
	}
	defer resp.Body.Close()

	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return nil
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing HTML from %s: %s\n", pageURL, err)
		return nil
	}

	return doc
}

func extractLinks(baseUrlParsed *url.URL, pageURL string, doc *goquery.Document) []model.LinkResult {
	pageLinks := []model.LinkResult{}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			return
		}

		linkURL := resolveAndFilterURL(baseUrlParsed, pageURL, href)
		if linkURL == nil {
			return
		}

		isExternal := baseUrlParsed.Hostname() != linkURL.Hostname()

		if (model.IgnoreExternal && isExternal) || (model.OnlyExternal && !isExternal) {
			return
		}

		if shouldSkipURLBasedOnPattern(linkURL) {
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
