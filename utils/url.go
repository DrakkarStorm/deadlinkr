package utils

import (
	"net/http"
	"net/url"
	"time"

	"github.com/DrakkarStorm/deadlinkr/model"
)

func resolveURL(pageURL, href string) (*url.URL, error) {
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

func shouldSkipURL(baseURL, linkURL *url.URL) bool {
	// Skip mailto, tel, javascript, etc.
	if linkURL.Scheme != "http" && linkURL.Scheme != "https" {
		return true
	}

	return false
}

func checkLink(linkURL string) (int, string) {
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

func CountBrokenLinks() int {
	count := 0
	for _, result := range model.Results {
		if result.Status >= 400 || result.Error != "" {
			count++
		}
	}
	return count
}
