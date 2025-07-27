package internal

import (
	"net/http"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// ServiceFactory creates and wires up all services
type ServiceFactory struct{}

// NewServiceFactory creates a new ServiceFactory
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{}
}

// CreateCrawlerService creates a fully configured crawler service
func (sf *ServiceFactory) CreateCrawlerService(config *CrawlConfig, userAgent string, timeout time.Duration, httpClient *http.Client) *CrawlerService {
	// Wrap HTTP client with authentication if configured
	authClient := sf.createAuthenticatedClient(httpClient)
	
	// Create services
	linkChecker := NewLinkCheckerService(authClient, userAgent, timeout)
	urlProcessor := NewURLProcessorService(config.IncludePattern, config.ExcludePattern)
	resultCollector := NewResultCollectorService()
	pageParser := NewPageParserService(linkChecker, urlProcessor, config.ExcludeHtmlTags, config.OnlyInternal)

	// Create crawler
	crawler := NewCrawlerService(pageParser, urlProcessor, resultCollector, config)

	return crawler
}

// CreateOptimizedCrawlerService creates an optimized crawler with worker pool
func (sf *ServiceFactory) CreateOptimizedCrawlerService(config *CrawlConfig, userAgent string, timeout time.Duration, httpClient *http.Client) *OptimizedCrawlerService {
	// Wrap HTTP client with authentication if configured
	authClient := sf.createAuthenticatedClient(httpClient)
	
	// Create services
	linkChecker := NewLinkCheckerService(authClient, userAgent, timeout)
	urlProcessor := NewURLProcessorService(config.IncludePattern, config.ExcludePattern)
	resultCollector := NewResultCollectorService()
	pageParser := NewPageParserService(linkChecker, urlProcessor, config.ExcludeHtmlTags, config.OnlyInternal)

	// Create optimized crawler
	crawler := NewOptimizedCrawlerService(pageParser, urlProcessor, resultCollector, config)

	return crawler
}

// CreateOptimizedCrawlerServiceWithRateLimit creates an optimized crawler with custom rate limiting
func (sf *ServiceFactory) CreateOptimizedCrawlerServiceWithRateLimit(config *CrawlConfig, userAgent string, timeout time.Duration, httpClient *http.Client, rateLimit, burst float64) *OptimizedCrawlerService {
	// Wrap HTTP client with authentication if configured
	authClient := sf.createAuthenticatedClient(httpClient)
	
	// Create services with custom rate limiting
	linkChecker := NewLinkCheckerServiceWithRateLimit(authClient, userAgent, timeout, rateLimit, burst)
	urlProcessor := NewURLProcessorService(config.IncludePattern, config.ExcludePattern)
	resultCollector := NewResultCollectorService()
	pageParser := NewPageParserService(linkChecker, urlProcessor, config.ExcludeHtmlTags, config.OnlyInternal)

	// Create optimized crawler
	crawler := NewOptimizedCrawlerService(pageParser, urlProcessor, resultCollector, config)

	return crawler
}

// CreateOptimizedCrawlerServiceWithHeadOptimization creates an optimized crawler with HEAD request optimization
func (sf *ServiceFactory) CreateOptimizedCrawlerServiceWithHeadOptimization(config *CrawlConfig, userAgent string, timeout time.Duration, httpClient *http.Client, rateLimit, burst float64) *OptimizedCrawlerService {
	// Wrap HTTP client with authentication if configured
	authClient := sf.createAuthenticatedClient(httpClient)
	
	// Create optimized link checker with HEAD requests
	linkChecker := NewOptimizedLinkCheckerService(authClient, userAgent, timeout, rateLimit, burst)
	urlProcessor := NewURLProcessorService(config.IncludePattern, config.ExcludePattern)
	resultCollector := NewResultCollectorService()
	pageParser := NewPageParserService(linkChecker, urlProcessor, config.ExcludeHtmlTags, config.OnlyInternal)

	// Create optimized crawler
	crawler := NewOptimizedCrawlerService(pageParser, urlProcessor, resultCollector, config)

	return crawler
}

// CreateCachedOptimizedCrawlerService creates an optimized crawler with caching and HEAD optimization
func (sf *ServiceFactory) CreateCachedOptimizedCrawlerService(config *CrawlConfig, userAgent string, timeout time.Duration, httpClient *http.Client, rateLimit, burst float64, cacheSize int, cacheTTL time.Duration) *OptimizedCrawlerService {
	// Wrap HTTP client with authentication if configured
	authClient := sf.createAuthenticatedClient(httpClient)
	
	// Create cached optimized link checker with HEAD requests
	linkChecker := NewCachedOptimizedLinkCheckerService(authClient, userAgent, timeout, rateLimit, burst, cacheSize, cacheTTL)
	urlProcessor := NewURLProcessorService(config.IncludePattern, config.ExcludePattern)
	resultCollector := NewResultCollectorService()
	pageParser := NewPageParserService(linkChecker, urlProcessor, config.ExcludeHtmlTags, config.OnlyInternal)

	// Create optimized crawler
	crawler := NewOptimizedCrawlerService(pageParser, urlProcessor, resultCollector, config)

	return crawler
}

// CreateCrawlConfig creates a CrawlConfig from the global model
func (sf *ServiceFactory) CreateCrawlConfig() *CrawlConfig {
	// Import from model package to avoid circular dependency issues
	// For now, we'll accept these as parameters or create a separate config struct
	return &CrawlConfig{
		MaxDepth:        2, // default
		Concurrency:     20, // default
		OnlyInternal:    false,
		IncludePattern:  "",
		ExcludePattern:  "",
		ExcludeHtmlTags: "",
	}
}

// CreateCrawlConfigFromParams creates a CrawlConfig from parameters
func (sf *ServiceFactory) CreateCrawlConfigFromParams(maxDepth, concurrency int, onlyInternal bool, includePattern, excludePattern, excludeHtmlTags string) *CrawlConfig {
	return &CrawlConfig{
		MaxDepth:        maxDepth,
		Concurrency:     concurrency,
		OnlyInternal:    onlyInternal,
		IncludePattern:  includePattern,
		ExcludePattern:  excludePattern,
		ExcludeHtmlTags: excludeHtmlTags,
	}
}

// createAuthenticatedClient wraps an HTTP client with authentication capabilities
func (sf *ServiceFactory) createAuthenticatedClient(httpClient *http.Client) *AuthenticatedHTTPClient {
	// Create authentication config
	config := NewAuthConfig()
	
	// Configure Basic Authentication
	if model.AuthBasic != "" {
		user, pass, err := ParseBasicAuthFromString(model.AuthBasic)
		if err != nil {
			logger.Errorf("Invalid basic auth format: %v", err)
		} else {
			config.BasicUser = user
			config.BasicPassword = pass
			config.BasicEnabled = true
			logger.Infof("Configured basic authentication for user: %s", user)
		}
	} else {
		// Check environment variables
		if user, pass, found := ParseBasicAuthFromEnv(); found {
			config.BasicUser = user
			config.BasicPassword = pass
			config.BasicEnabled = true
			logger.Infof("Configured basic authentication from environment for user: %s", user)
		}
	}
	
	// Configure Bearer Token
	if model.AuthBearer != "" {
		config.BearerToken = model.AuthBearer
		config.BearerEnabled = true
		logger.Infof("Configured bearer token authentication")
	} else {
		// Check environment variable
		if token, found := ParseBearerTokenFromEnv(); found {
			config.BearerToken = token
			config.BearerEnabled = true
			logger.Infof("Configured bearer token authentication from environment")
		}
	}
	
	// Configure Custom Headers
	if len(model.AuthHeaders) > 0 {
		config.CustomHeaders = make(map[string]string)
		for _, header := range model.AuthHeaders {
			key, value, err := ParseCustomHeaderFromString(header)
			if err != nil {
				logger.Errorf("Invalid header format '%s': %v", header, err)
				continue
			}
			config.CustomHeaders[key] = value
		}
		if len(config.CustomHeaders) > 0 {
			config.HeadersEnabled = true
			logger.Infof("Configured %d custom authentication headers", len(config.CustomHeaders))
		}
	} else {
		// Check environment variable
		if envHeaders := ParseCustomHeadersFromEnv(); len(envHeaders) > 0 {
			config.CustomHeaders = envHeaders
			config.HeadersEnabled = true
			logger.Infof("Configured %d custom authentication headers from environment", len(envHeaders))
		}
	}
	
	// Configure Cookies
	if model.AuthCookies != "" {
		config.Cookies = model.AuthCookies
		config.CookiesEnabled = true
		logger.Infof("Configured cookie authentication")
	}
	
	// Create authenticated client
	authClient := NewAuthenticatedHTTPClient(httpClient, config)
	
	// Log authentication summary
	if config.BasicEnabled || config.BearerEnabled || config.HeadersEnabled || config.CookiesEnabled {
		logger.Infof("Authentication configured: %s", authClient.GetAuthSummary())
	}
	
	return authClient
}