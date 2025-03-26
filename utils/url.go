package utils

import (
	"net/http"
	"net/url"
	"time"

	"github.com/DrakkarStorm/deadlinkr/model"
)

// ResolveURL resolves a relative URL to an absolute URL.
// example: ResolveURL("https://example.com", "/about") -> "https://example.com/about"
func ResolveURL(pageURL, href string) (*url.URL, error) {
	hrefURL, err := url.Parse(href)
	if err != nil {
		return nil, err
	}

	pageUrlParsed, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}

	// Resolve relative URLs
	resolvedURL := pageUrlParsed.ResolveReference(hrefURL)

	return resolvedURL, nil
}

// ShouldSkipURL checks if a URL should be skipped.
// example: ShouldSkipURL("https://example.com", "mailto:example@example.com") -> true
func ShouldSkipURL(baseURL, linkURL *url.URL) bool {
	// Skip mailto, tel, javascript, etc.
	if linkURL.Scheme != "http" && linkURL.Scheme != "https" {
		return true
	}

	return false
}

// CheckLink checks if a link is broken.
// example: CheckLink("https://example.com") -> 200, ""
// example: CheckLink("https://example.com/404") -> 404, ""
func CheckLink(linkURL string) (int, string) {
	client := &http.Client{
		Timeout: time.Duration(model.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Return nil to follow redirects
			return nil
		},
	}

	req, err := http.NewRequest("HEAD", linkURL, nil)
	if err != nil {
		return 0, err.Error()
	}

	req.Header.Set("User-Agent", model.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer resp.Body.Close()

	return resp.StatusCode, ""
}

// CountBrokenLinks counts the number of broken links.
func CountBrokenLinks() int {
	count := 0
	for _, result := range model.Results {
		if result.Status >= 400 || result.Error != "" {
			count++
		}
	}
	return count
}
