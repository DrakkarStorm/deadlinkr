package model

import (
	"sync"
	"time"
)

// Depth is the maximum depth for crawling
var Depth int

// Quiet indicates whether to disable output
var Quiet bool

// LogLevel is the log level for the application
var LogLevel string

// Concurrency is the number of concurrent requests
var Concurrency int

// Timeout is the timeout for each request
var Timeout int

// OnlyInternal indicates whether to check only internal links
var OnlyInternal bool

// UserAgent is the user agent string for requests
var UserAgent string

// IncludePattern is the regex pattern for including URLs
var IncludePattern string

// ExcludePattern is the regex pattern for excluding URLs
var ExcludePattern string

// ExcludeHtmlTags is the list of HTML tags
var ExcludeHtmlTags string

// DisplayOnlyError indicates whether to display only error (legacy - inverted logic)
var DisplayOnlyError bool = true

// ShowAll indicates whether to show all links including working ones
var ShowAll bool

// DisplayOnlyExternal indicates whether to display only external links
var DisplayOnlyExternal bool

// Format is the export format for results
var Format string

// Results is a slice of LinkResult containing the results of the scan
var Results []LinkResult

// VisitedURLs is a sync.Map to keep track of visited URLs
var VisitedURLs sync.Map

// ResultsMutex is a sync.Mutex to protect concurrent access to Results
var ResultsMutex sync.Mutex

// Wg is a sync.WaitGroup to wait for all goroutines to finish
var Wg sync.WaitGroup

// timeExecution is the start time of the execution of the program
var TimeExecution time.Time

// Output is the output file for the results
var Output string

// RateLimitRequestsPerSecond is the default rate limit per domain
var RateLimitRequestsPerSecond float64 = 2.0

// RateLimitBurst is the burst capacity for rate limiting
var RateLimitBurst float64 = 5.0

// OptimizeWithHeadRequests enables HEAD request optimization
var OptimizeWithHeadRequests bool = true

// CacheEnabled enables link result caching
var CacheEnabled bool = true

// CacheSize is the maximum number of entries in the cache
var CacheSize int = 1000

// CacheTTL is the default cache time-to-live in minutes
var CacheTTLMinutes int = 60

// Authentication settings
var AuthBasic string       // Basic auth in "user:password" format
var AuthBearer string      // Bearer token
var AuthHeaders []string   // Custom headers in "Key: Value" format
var AuthCookies string     // Cookies string