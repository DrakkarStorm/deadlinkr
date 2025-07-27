package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleWorkerPoolCreation(t *testing.T) {
	t.Run("Can create worker pool components", func(t *testing.T) {
		// Just test that we can create the basic structures
		config := &CrawlConfig{
			MaxDepth:     1,
			Concurrency:  2,
			OnlyInternal: false,
		}
		
		assert.NotNil(t, config)
		assert.Equal(t, 1, config.MaxDepth)
		assert.Equal(t, 2, config.Concurrency)
	})
}