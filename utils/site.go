package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/EnzoDechaene/deadlinkr/model"
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
	go func(url string, d int) {
		defer model.Wg.Done()

		links := CheckLinks(baseURL, url)

		// Recursive crawling for internal links
		if d < model.Depth {
			for _, link := range links {
				if !link.IsExternal {
					go Crawl(baseURL, link.TargetURL, d+1)
				}
			}
		}
	}(currentURL, currentDepth)
}

func CheckLinks(baseURL, pageURL string) []model.LinkResult {
	pageLinks := []model.LinkResult{}

	baseUrlParsed, err := url.Parse(baseURL)
	if err != nil {
		fmt.Printf("Error parsing base URL %s: %s\n", baseURL, err)
		return pageLinks
	}

	client := &http.Client{
		Timeout: time.Duration(model.Timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		fmt.Printf("Error creating request for %s: %s\n", pageURL, err)
		return pageLinks
	}

	req.Header.Set("User-Agent", model.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error fetching %s: %s\n", pageURL, err)
		return pageLinks
	}
	defer resp.Body.Close()

	// Only parse HTML responses
	if !strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		return pageLinks
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing HTML from %s: %s\n", pageURL, err)
		return pageLinks
	}

	// Find all links
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" || strings.HasPrefix(href, "#") {
			return
		}

		linkURL, err := resolveURL(baseUrlParsed, pageURL, href)
		if err != nil {
			return
		}

		// Apply filters
		if shouldSkipURL(baseUrlParsed, linkURL) {
			return
		}

		isExternal := baseUrlParsed.Hostname() != linkURL.Hostname()

		if (model.IgnoreExternal && isExternal) || (model.OnlyExternal && !isExternal) {
			return
		}

		// Apply regex patterns if specified
		if model.IncludePattern != "" {
			matched, err := regexp.MatchString(model.IncludePattern, linkURL.String())
			if err != nil || !matched {
				return
			}
		}

		if model.ExcludePattern != "" {
			matched, err := regexp.MatchString(model.ExcludePattern, linkURL.String())
			if err == nil && matched {
				return
			}
		}

		// Check the link
		status, errMsg := checkLink(linkURL.String())

		linkResult := model.LinkResult{
			SourceURL:  pageURL,
			TargetURL:  linkURL.String(),
			Status:     status,
			Error:      errMsg,
			IsExternal: isExternal,
		}

		pageLinks = append(pageLinks, linkResult)

		model.ResultsMutex.Lock()
		model.Results = append(model.Results, linkResult)
		model.ResultsMutex.Unlock()
	})

	return pageLinks
}
