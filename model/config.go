package model

import "sync"

var (
	Depth          int
	Concurrency    int
	Timeout        int
	IgnoreExternal bool
	OnlyExternal   bool
	UserAgent      string
	IncludePattern string
	ExcludePattern string
	Format         string

	Results      []LinkResult
	VisitedURLs  sync.Map
	ResultsMutex sync.Mutex
	Wg           sync.WaitGroup
)
