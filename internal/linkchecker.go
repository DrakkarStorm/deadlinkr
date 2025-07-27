package internal

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// LinkCheckerService implements the LinkChecker interface
type LinkCheckerService struct {
	client      HTTPClient
	userAgent   string
	timeout     time.Duration
	rateLimiter *DomainRateLimiter
}

// NewLinkCheckerService creates a new LinkCheckerService
func NewLinkCheckerService(client HTTPClient, userAgent string, timeout time.Duration) *LinkCheckerService {
	// Default: 2 requests per second per domain, burst of 5
	rateLimiter := NewDomainRateLimiter(2.0, 5.0)
	
	return &LinkCheckerService{
		client:      client,
		userAgent:   userAgent,
		timeout:     timeout,
		rateLimiter: rateLimiter,
	}
}

// NewLinkCheckerServiceWithRateLimit creates a LinkCheckerService with custom rate limiting
func NewLinkCheckerServiceWithRateLimit(client HTTPClient, userAgent string, timeout time.Duration, requestsPerSecond, burst float64) *LinkCheckerService {
	rateLimiter := NewDomainRateLimiter(requestsPerSecond, burst)
	
	return &LinkCheckerService{
		client:      client,
		userAgent:   userAgent,
		timeout:     timeout,
		rateLimiter: rateLimiter,
	}
}

// CheckLink checks if a link is broken
func (lc *LinkCheckerService) CheckLink(linkURL string) (int, string) {
	resp, err := lc.FetchWithRetry(linkURL, 3)
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

// FetchWithRetry fetches a URL with retry logic
func (lc *LinkCheckerService) FetchWithRetry(url string, retry int) (*model.HTTPResponse, error) {
	// Apply rate limiting before each retry attempt
	if err := lc.rateLimiter.Wait(url); err != nil {
		logger.Errorf("Rate limiting error for %s: %s", url, err)
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", lc.userAgent)

	var resp *http.Response
	var errRequest error
	for i := 1; i <= retry; i++ {
		resp, errRequest = lc.client.Do(req)
		if errRequest == nil {
			return &model.HTTPResponse{
				Response: resp,
			}, nil
		}
		logger.Errorf("Attempt %d failed: %v, retrying in %d seconds...", i, errRequest, 5)
		time.Sleep(5 * time.Second)
	}
	return nil, errRequest
}

// SetDomainRateLimit sets a custom rate limit for a specific domain
func (lc *LinkCheckerService) SetDomainRateLimit(domain string, requestsPerSecond float64) {
	lc.rateLimiter.UpdateConfig(domain, requestsPerSecond)
}

// GetRateLimiterStats returns current rate limiter statistics
func (lc *LinkCheckerService) GetRateLimiterStats() map[string]RateLimiterStats {
	return lc.rateLimiter.GetStats()
}