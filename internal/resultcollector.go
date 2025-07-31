package internal

import (
	"sync"

	"github.com/DrakkarStorm/deadlinkr/model"
)

// ResultCollectorService implements the ResultCollector interface
type ResultCollectorService struct {
	results     []model.LinkResult
	visitedURLs sync.Map
	mutex       sync.Mutex
}

// NewResultCollectorService creates a new ResultCollectorService
func NewResultCollectorService() *ResultCollectorService {
	return &ResultCollectorService{
		results:     make([]model.LinkResult, 0),
		visitedURLs: sync.Map{},
	}
}

// AddResult adds a result to the collection
func (rc *ResultCollectorService) AddResult(result model.LinkResult) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	rc.results = append(rc.results, result)
}

// GetResults returns all collected results
func (rc *ResultCollectorService) GetResults() []model.LinkResult {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	// Return a copy to avoid race conditions
	resultsCopy := make([]model.LinkResult, len(rc.results))
	copy(resultsCopy, rc.results)
	return resultsCopy
}

// CountBrokenLinks counts the number of broken links
func (rc *ResultCollectorService) CountBrokenLinks() int {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	count := 0
	for _, result := range rc.results {
		if result.Status >= 400 || result.Error != "" {
			count++
		}
	}
	return count
}

// IsVisited checks if a URL has been visited
func (rc *ResultCollectorService) IsVisited(url string) bool {
	_, exists := rc.visitedURLs.Load(url)
	return exists
}

// MarkVisited marks a URL as visited
func (rc *ResultCollectorService) MarkVisited(url string) {
	rc.visitedURLs.Store(url, true)
}

// Clear clears all results and visited URLs
func (rc *ResultCollectorService) Clear() {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()
	
	rc.results = make([]model.LinkResult, 0)
	rc.visitedURLs = sync.Map{}
}