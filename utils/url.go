package utils

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
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
	resp, err := FetchWithRetry(linkURL, 3)
	if err != nil {
		return 0, err.Error()
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Errorf("Error closing response body for %s: %s", linkURL, err)
		}
	}()

	// Analyse the MIME type to detect files
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/") ||
		strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "font/") ||
		strings.Contains(contentType, "text/plain") {
		logger.Debugf("The URL appears to point to a file (MIME type: %s)\n", contentType)
		return resp.StatusCode, ""
	}

	// If it's a webpage, check for errors
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "Error reading response body: " + err.Error()
	}

	if len(body) == 0 {
		return resp.StatusCode, "The response body is empty"
	}

	return resp.StatusCode, ""
}

func FetchWithRetry(url string, retry int) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", model.UserAgent)

	var resp *http.Response
	var errRequest error
	for i := 1; i <= retry; i++ {
		resp, errRequest = ClientHTTP.Do(req)
		if errRequest == nil {
			return resp, nil
		}
		logger.Errorf("Attempt %d failed: %v, retrying in %d seconds...", i, errRequest, 5)
		time.Sleep(5 * time.Second)
	}
	return nil, errRequest
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
