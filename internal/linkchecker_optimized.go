package internal

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// OptimizedLinkCheckerService implements smart HEAD/GET request strategy
type OptimizedLinkCheckerService struct {
	client           HTTPClient
	userAgent        string
	timeout          time.Duration
	rateLimiter      *DomainRateLimiter
	headSupport      map[string]bool // Track which domains support HEAD
	headSupportMutex sync.RWMutex
	stats            *OptimizedLinkStats
}

// OptimizedLinkStats tracks optimization statistics
type OptimizedLinkStats struct {
	HeadRequestsUsed     int64
	GetRequestsUsed      int64
	HeadFallbacks        int64
	BytesSaved           int64
	TimeSaved            time.Duration
	mutex                sync.RWMutex
}

// NewOptimizedLinkCheckerService creates an optimized link checker
func NewOptimizedLinkCheckerService(client HTTPClient, userAgent string, timeout time.Duration, requestsPerSecond, burst float64) *OptimizedLinkCheckerService {
	rateLimiter := NewDomainRateLimiter(requestsPerSecond, burst)
	
	return &OptimizedLinkCheckerService{
		client:      client,
		userAgent:   userAgent,
		timeout:     timeout,
		rateLimiter: rateLimiter,
		headSupport: make(map[string]bool),
		stats:       &OptimizedLinkStats{},
	}
}

// CheckLink uses optimized HEAD/GET strategy
func (olc *OptimizedLinkCheckerService) CheckLink(linkURL string) (int, string) {
	domain, err := extractDomain(linkURL)
	if err != nil {
		return 0, "Invalid URL: " + err.Error()
	}

	// Check if we know this domain supports HEAD
	useHead := olc.shouldTryHead(domain)
	
	if useHead {
		// Try HEAD first
		status, errMsg, success := olc.tryHeadRequest(linkURL)
		if success {
			olc.recordHeadSuccess(domain)
			olc.stats.incrementHeadRequests()
			return status, errMsg
		} else {
			// HEAD failed, mark domain and fallback to GET
			olc.recordHeadFailure(domain)
			olc.stats.incrementHeadFallbacks()
			logger.Debugf("HEAD failed for %s, falling back to GET", linkURL)
		}
	}

	// Use GET request
	status, errMsg := olc.getRequest(linkURL)
	olc.stats.incrementGetRequests()
	return status, errMsg
}

// tryHeadRequest attempts a HEAD request
func (olc *OptimizedLinkCheckerService) tryHeadRequest(linkURL string) (int, string, bool) {
	start := time.Now()
	
	resp, err := olc.FetchWithRetryMethod(linkURL, 1, "HEAD") // Only 1 retry for HEAD
	if err != nil {
		return 0, err.Error(), false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Errorf("Error closing HEAD response body for %s: %s", linkURL, err)
		}
	}()

	// Check if HEAD is actually supported
	if resp.StatusCode == 405 || resp.StatusCode == 501 { // Method Not Allowed / Not Implemented
		return 0, "", false
	}

	duration := time.Since(start)
	olc.stats.addTimeSaved(duration)

	// For HEAD requests, we only care about the status code
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return resp.StatusCode, "", true
	}
	
	return resp.StatusCode, "", true
}

// getRequest performs a standard GET request
func (olc *OptimizedLinkCheckerService) getRequest(linkURL string) (int, string) {
	resp, err := olc.FetchWithRetryMethod(linkURL, 3, "GET")
	if err != nil {
		return 0, err.Error()
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Errorf("Error closing GET response body for %s: %s", linkURL, err)
		}
	}()

	// Analyze the MIME type to detect files
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/") ||
		strings.Contains(contentType, "image/") ||
		strings.Contains(contentType, "video/") ||
		strings.Contains(contentType, "audio/") ||
		strings.Contains(contentType, "font/") ||
		strings.Contains(contentType, "text/plain") {
		logger.Debugf("The URL appears to point to a file (MIME type: %s)", contentType)
		return resp.StatusCode, ""
	}

	// For HTML content, do a minimal read to check if it's valid
	if strings.Contains(contentType, "text/html") {
		// Read only a small portion to verify it's not empty
		limitedReader := io.LimitReader(resp.Body, 1024) // Read max 1KB
		body, err := io.ReadAll(limitedReader)
		if err != nil {
			return resp.StatusCode, "Error reading response body: " + err.Error()
		}

		if len(body) == 0 {
			return resp.StatusCode, "The response body is empty"
		}
		
		// Record bytes saved vs full download
		olc.stats.addBytesSaved(int64(len(body)))
	}

	return resp.StatusCode, ""
}

// FetchWithRetryMethod performs HTTP request with specified method
func (olc *OptimizedLinkCheckerService) FetchWithRetryMethod(url string, retry int, method string) (*model.HTTPResponse, error) {
	// Apply rate limiting
	if err := olc.rateLimiter.Wait(url); err != nil {
		logger.Errorf("Rate limiting error for %s: %s", url, err)
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", olc.userAgent)
	
	// For HEAD requests, we want to minimize data transfer
	if method == "HEAD" {
		req.Header.Set("Accept", "*/*")
	}

	var resp *http.Response
	var errRequest error
	for i := 1; i <= retry; i++ {
		resp, errRequest = olc.client.Do(req)
		if errRequest == nil {
			return &model.HTTPResponse{Response: resp}, nil
		}
		
		if i < retry {
			logger.Errorf("Attempt %d failed: %v, retrying in %d seconds...", i, errRequest, 2)
			time.Sleep(2 * time.Second)
		}
	}
	return nil, errRequest
}

// FetchWithRetry implements the LinkChecker interface (defaults to GET)
func (olc *OptimizedLinkCheckerService) FetchWithRetry(url string, retry int) (*model.HTTPResponse, error) {
	return olc.FetchWithRetryMethod(url, retry, "GET")
}

// Head support tracking methods
func (olc *OptimizedLinkCheckerService) shouldTryHead(domain string) bool {
	olc.headSupportMutex.RLock()
	defer olc.headSupportMutex.RUnlock()
	
	// If we haven't tested this domain, try HEAD
	supported, exists := olc.headSupport[domain]
	return !exists || supported
}

func (olc *OptimizedLinkCheckerService) recordHeadSuccess(domain string) {
	olc.headSupportMutex.Lock()
	defer olc.headSupportMutex.Unlock()
	olc.headSupport[domain] = true
}

func (olc *OptimizedLinkCheckerService) recordHeadFailure(domain string) {
	olc.headSupportMutex.Lock()
	defer olc.headSupportMutex.Unlock()
	olc.headSupport[domain] = false
}

// Statistics methods
func (olc *OptimizedLinkCheckerService) GetOptimizationStats() OptimizedLinkStats {
	olc.stats.mutex.RLock()
	defer olc.stats.mutex.RUnlock()
	return OptimizedLinkStats{
		HeadRequestsUsed: olc.stats.HeadRequestsUsed,
		GetRequestsUsed:  olc.stats.GetRequestsUsed,
		HeadFallbacks:    olc.stats.HeadFallbacks,
		BytesSaved:       olc.stats.BytesSaved,
		TimeSaved:        olc.stats.TimeSaved,
	}
}

func (stats *OptimizedLinkStats) incrementHeadRequests() {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	stats.HeadRequestsUsed++
}

func (stats *OptimizedLinkStats) incrementGetRequests() {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	stats.GetRequestsUsed++
}

func (stats *OptimizedLinkStats) incrementHeadFallbacks() {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	stats.HeadFallbacks++
}

func (stats *OptimizedLinkStats) addBytesSaved(bytes int64) {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	stats.BytesSaved += bytes
}

func (stats *OptimizedLinkStats) addTimeSaved(duration time.Duration) {
	stats.mutex.Lock()
	defer stats.mutex.Unlock()
	stats.TimeSaved += duration
}

// Implement other required methods for compatibility
func (olc *OptimizedLinkCheckerService) SetDomainRateLimit(domain string, requestsPerSecond float64) {
	olc.rateLimiter.UpdateConfig(domain, requestsPerSecond)
}

func (olc *OptimizedLinkCheckerService) GetRateLimiterStats() map[string]RateLimiterStats {
	return olc.rateLimiter.GetStats()
}