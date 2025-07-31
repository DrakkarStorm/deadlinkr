package internal

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServiceFactory(t *testing.T) {
	factory := NewServiceFactory()
	
	t.Run("Creates factory", func(t *testing.T) {
		assert.NotNil(t, factory)
	})

	t.Run("Creates config from parameters", func(t *testing.T) {
		config := factory.CreateCrawlConfigFromParams(
			2,     // maxDepth
			10,    // concurrency
			true,  // onlyInternal
			".*",  // includePattern
			"skip", // excludePattern
			"nav", // excludeHtmlTags
		)

		assert.Equal(t, 2, config.MaxDepth)
		assert.Equal(t, 10, config.Concurrency)
		assert.True(t, config.OnlyInternal)
		assert.Equal(t, ".*", config.IncludePattern)
		assert.Equal(t, "skip", config.ExcludePattern)
		assert.Equal(t, "nav", config.ExcludeHtmlTags)
	})

	t.Run("Creates crawler service", func(t *testing.T) {
		config := factory.CreateCrawlConfigFromParams(1, 5, false, "", "", "")
		httpClient := &http.Client{Timeout: 10 * time.Second}
		
		crawler := factory.CreateCrawlerService(config, "TestAgent", 5*time.Second, httpClient)
		
		assert.NotNil(t, crawler)
	})
}

func TestURLProcessor(t *testing.T) {
	processor := NewURLProcessorService("", "")

	t.Run("Resolves relative URL", func(t *testing.T) {
		resolved, err := processor.ResolveURL("https://example.com/page", "about")
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/about", resolved.String())
	})

	t.Run("Validates base URL", func(t *testing.T) {
		url, err := processor.ValidateURL("https://example.com")
		assert.NoError(t, err)
		assert.Equal(t, "example.com", url.Host)
	})

	t.Run("Rejects invalid schemes", func(t *testing.T) {
		base, _ := processor.ValidateURL("https://example.com")
		link, _ := processor.ResolveURL("https://example.com", "mailto:test@example.com")
		
		shouldSkip := processor.ShouldSkipURL(base, link)
		assert.True(t, shouldSkip)
	})
}

func TestResultCollector(t *testing.T) {
	collector := NewResultCollectorService()

	t.Run("Starts empty", func(t *testing.T) {
		results := collector.GetResults()
		assert.Empty(t, results)
		assert.Equal(t, 0, collector.CountBrokenLinks())
	})

	t.Run("Manages visited URLs", func(t *testing.T) {
		url := "https://example.com"
		
		assert.False(t, collector.IsVisited(url))
		collector.MarkVisited(url)
		assert.True(t, collector.IsVisited(url))
	})

	t.Run("Clears data", func(t *testing.T) {
		collector.MarkVisited("https://test.com")
		collector.Clear()
		
		assert.False(t, collector.IsVisited("https://test.com"))
		assert.Empty(t, collector.GetResults())
	})
}