package model

import "sync"

// Depth is the maximum depth for crawling
var Depth int

// Concurrency is the number of concurrent requests
var Concurrency int

// Timeout is the timeout for each request
var Timeout int

// IgnoreExternal indicates whether to ignore external links
var IgnoreExternal bool

// OnlyExternal indicates whether to check only external links
var OnlyExternal bool

// UserAgent is the user agent string for requests
var UserAgent string

// IncludePattern is the regex pattern for including URLs
var IncludePattern string

// ExcludePattern is the regex pattern for excluding URLs
var ExcludePattern string

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