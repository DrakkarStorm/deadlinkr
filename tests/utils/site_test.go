package utils_test

// TestCrawl tests the Crawl function
// func TestCrawl(t *testing.T) {
// 	// This is a more complex test as it involves concurrency, network requests, etc.
// 	// We'll set up a test server with specific paths to test crawling behavior

// 	mux := http.NewServeMux()

// 	// Root page with links
// 	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		if r.URL.Path != "/" {
// 			http.NotFound(w, r)
// 			return
// 		}
// 		fmt.Fprintf(w, `
// 			<html>
// 				<body>
// 					<a href="/page1">Page 1</a>
// 					<a href="/page2">Page 2</a>
// 					<a href="http://external.example.com">External</a>
// 					<a href="/nonexistent">Broken Link</a>
// 				</body>
// 			</html>
// 		`)
// 	})

// 	// Page 1 with additional links
// 	mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, `
// 			<html>
// 				<body>
// 					<a href="/">Home</a>
// 					<a href="/page3">Page 3</a>
// 				</body>
// 			</html>
// 		`)
// 	})

// 	// Page 2 (empty page)
// 	mux.HandleFunc("/page2", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, "<html><body>Page 2</body></html>")
// 	})

// 	// Page 3 (only available from page 1)
// 	mux.HandleFunc("/page3", func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, "<html><body>Page 3</body></html>")
// 	})

// 	server := httptest.NewServer(mux)
// 	defer server.Close()

// 	teardown := setupTest()
// 	defer teardown()

// 	// Set crawl depth
// 	model.Depth = 2

// 	// Start crawl
// 	utils.Crawl(server.URL, server.URL, 0)

// 	// Wait for all crawl goroutines to finish
// 	model.Wg.Wait()

// 	// Verify results
// 	// We should have visited at least the internal pages
// 	visitedCount := 0
// 	model.VisitedURLs.Range(func(key, value interface{}) bool {
// 		visitedCount++
// 		return true
// 	})

// 	// We should have visited at least 4 URLs (/, /page1, /page2, /page3)
// 	assert.GreaterOrEqual(t, visitedCount, 4)

// 	// We should have at least one broken link (/nonexistent)
// 	brokenCount := utils.CountBrokenLinks()
// 	assert.GreaterOrEqual(t, brokenCount, 1)

// 	// Verify we didn't go beyond max depth
// 	for _, result := range model.Results {
// 		// All results should be from the test server except external links
// 		if !strings.Contains(result.TargetURL, "external") {
// 			assert.True(t, strings.HasPrefix(result.TargetURL, server.URL))
// 		}
// 	}
// }
