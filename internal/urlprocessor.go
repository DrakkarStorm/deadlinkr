package internal

import (
	"net/url"
	"regexp"

	"github.com/DrakkarStorm/deadlinkr/logger"
)

// URLProcessorService implements the URLProcessor interface
type URLProcessorService struct {
	includePattern string
	excludePattern string
}

// NewURLProcessorService creates a new URLProcessorService
func NewURLProcessorService(includePattern, excludePattern string) *URLProcessorService {
	return &URLProcessorService{
		includePattern: includePattern,
		excludePattern: excludePattern,
	}
}

// ResolveURL resolves a relative URL to an absolute URL
func (up *URLProcessorService) ResolveURL(pageURL, href string) (*url.URL, error) {
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

// ShouldSkipURL checks if a URL should be skipped
func (up *URLProcessorService) ShouldSkipURL(baseURL, linkURL *url.URL) bool {
	// Skip mailto, tel, javascript, etc.
	if linkURL.Scheme != "http" && linkURL.Scheme != "https" {
		return true
	}

	// Check patterns
	return up.shouldSkipURLBasedOnPattern(linkURL)
}

// ValidateURL validates and parses a base URL
func (up *URLProcessorService) ValidateURL(baseURL string) (*url.URL, error) {
	baseUrlParsed, err := url.Parse(baseURL)
	if err != nil {
		logger.Errorf("Error parsing base URL %s: %s", baseURL, err)
		return nil, err
	}
	if baseUrlParsed.Host == "" {
		logger.Errorf("Error parsing base URL %s: no host found", baseURL)
		return nil, err
	}
	return baseUrlParsed, nil
}

// shouldSkipURLBasedOnPattern checks if URL should be skipped based on patterns
func (up *URLProcessorService) shouldSkipURLBasedOnPattern(linkURL *url.URL) bool {
	if up.includePattern != "" {
		matched, err := regexp.MatchString(up.includePattern, linkURL.String())
		if err != nil || !matched {
			return true
		}
	}

	if up.excludePattern != "" {
		matched, err := regexp.MatchString(up.excludePattern, linkURL.String())
		if err == nil && matched {
			return true
		}
	}

	return false
}

// SetPatterns updates the include and exclude patterns
func (up *URLProcessorService) SetPatterns(includePattern, excludePattern string) {
	up.includePattern = includePattern
	up.excludePattern = excludePattern
}