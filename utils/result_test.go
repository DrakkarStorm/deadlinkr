package utils

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const SOURCE_URL = "http://127.0.0.1:8085"

// Helper function to set up test environment
func setupTest() func() {
	logger.InitLogger("debug")
	// Save original model.Results
	originalResults := model.Results
	// originalVisitedURLs := model.VisitedURLs
	originalDepth := model.Depth
	originalTimeout := model.Timeout
	originalUserAgent := model.UserAgent
	// originalWg := model.Wg

	// Set up test environment
	model.Results = []model.LinkResult{}
	model.VisitedURLs = sync.Map{}
	model.Depth = 2
	model.Timeout = 5
	model.UserAgent = "TestUserAgent"
	model.Wg = sync.WaitGroup{}

	// Return a function to restore original state
	return func() {
		model.Results = originalResults
		// model.VisitedURLs = originalVisitedURLs
		model.Depth = originalDepth
		model.Timeout = originalTimeout
		model.UserAgent = originalUserAgent
		// model.Wg = originalWg
		logger.CloseLogger()

		// Clean up test files
		_ = os.Remove("deadlinkr-report.csv")
		_ = os.Remove("deadlinkr-report.json")
		_ = os.Remove("deadlinkr-report.html")
		_ = os.Remove("deadlinkr.log")
	}
}

// TestDisplayResults tests the DisplayResults function
func TestDisplayResults(t *testing.T) {
	teardown := setupTest()
	defer teardown()

	testCases := []struct {
		name     string
		results  []model.LinkResult
		expected string
	}{
		{
			name:     "No broken links",
			results:  []model.LinkResult{},
			expected: "No broken links found!",
		},
		{
			name: "With error links",
			results: []model.LinkResult{
				{SourceURL: SOURCE_URL, TargetURL: "http://broken.com", Status: 0, Error: "connection error"},
			},
			expected: "Broken links:\n=============\n- http://broken.com (from http://127.0.0.1:8085): Error: connection error",
		},
		{
			name: "With status error links",
			results: []model.LinkResult{
				{SourceURL: SOURCE_URL, TargetURL: "http://broken.com", Status: 404, Error: ""},
			},
			expected: "Broken links:\n=============\n- http://broken.com (from http://127.0.0.1:8085): Status: 404",
		},
		{
			name: "Mixed links",
			results: []model.LinkResult{
				{SourceURL: SOURCE_URL, TargetURL: "http://127.0.0.1:8085", Status: 200, Error: ""},
				{SourceURL: SOURCE_URL, TargetURL: "http://broken.com", Status: 404, Error: ""},
				{SourceURL: SOURCE_URL, TargetURL: "http://error.com", Status: 0, Error: "timeout"},
			},
			expected: "Broken links:\n=============\n- http://broken.com (from http://127.0.0.1:8085): Status: 404\n- http://error.com (from http://127.0.0.1:8085): Error: timeout",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create buffer to capture output
			var buf bytes.Buffer
			origStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Set test data
			model.Results = tc.results

			// Call function
			DisplayResults()

			// Restore stdout and get output
			_ = w.Close()
			os.Stdout = origStdout
			_, err := io.Copy(&buf, r)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			output := strings.TrimSpace(buf.String())

			// Verify output
			assert.Contains(t, output, tc.expected)
		})
	}
}

// TestExportToCSV tests the CSV export functionality
func TestExportToCSV(t *testing.T) {
	teardown := setupTest()
	defer teardown()

	// Set up test data
	model.Results = []model.LinkResult{
		{SourceURL: SOURCE_URL, TargetURL: "http://127.0.0.1:8085", Status: 200, Error: "", IsExternal: false},
		{SourceURL: SOURCE_URL, TargetURL: "http://broken.com", Status: 404, Error: "", IsExternal: true},
		{SourceURL: SOURCE_URL, TargetURL: "http://error.com", Status: 0, Error: "timeout", IsExternal: true},
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call exportToCSV (we need to use ExportResults since exportToCSV is private)
	ExportResults("csv")

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout
	_, err := io.Copy(&buf, r)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat("deadlinkr-report.csv")
	require.NoError(t, err, "CSV file should be created")

	// Read and verify file content
	file, err := os.Open("deadlinkr-report.csv")
	require.NoError(t, err)
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("Error closing test file: %v", err)
		}
	}()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Verify header
	assert.Equal(t, []string{"Source URL", "Target URL", "Status", "Error", "Is External"}, records[0])

	// Verify data rows
	assert.Equal(t, SOURCE_URL, records[1][0])
	assert.Equal(t, "http://127.0.0.1:8085", records[1][1])
	assert.Equal(t, "200", records[1][2])
	assert.Equal(t, "", records[1][3])
	assert.Equal(t, "false", records[1][4])

	assert.Equal(t, SOURCE_URL, records[2][0])
	assert.Equal(t, "http://broken.com", records[2][1])
	assert.Equal(t, "404", records[2][2])
	assert.Equal(t, "", records[2][3])
	assert.Equal(t, "true", records[2][4])

	assert.Equal(t, SOURCE_URL, records[3][0])
	assert.Equal(t, "http://error.com", records[3][1])
	assert.Equal(t, "0", records[3][2])
	assert.Equal(t, "timeout", records[3][3])
	assert.Equal(t, "true", records[3][4])
}

// TestExportToJSON tests the JSON export functionality
func TestExportToJSON(t *testing.T) {
	teardown := setupTest()
	defer teardown()

	// Set up test data
	model.Results = []model.LinkResult{
		{SourceURL: SOURCE_URL, TargetURL: "http://127.0.0.1:8085", Status: 200, Error: "", IsExternal: false},
		{SourceURL: SOURCE_URL, TargetURL: "http://broken.com", Status: 404, Error: "", IsExternal: true},
	}

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call exportToJSON (we need to use ExportResults since exportToJSON is private)
	ExportResults("json")

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout
	_, err := io.Copy(&buf, r)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat("deadlinkr-report.json")
	require.NoError(t, err, "JSON file should be created")

	// Read and verify file content
	fileContent, err := os.ReadFile("deadlinkr-report.json")
	require.NoError(t, err)

	var results []model.LinkResult
	err = json.Unmarshal(fileContent, &results)
	require.NoError(t, err)

	// Verify content
	assert.Equal(t, 2, len(results))
	assert.Equal(t, SOURCE_URL, results[0].SourceURL)
	assert.Equal(t, "http://127.0.0.1:8085", results[0].TargetURL)
	assert.Equal(t, 200, results[0].Status)
	assert.Equal(t, "", results[0].Error)
	assert.Equal(t, false, results[0].IsExternal)

	assert.Equal(t, SOURCE_URL, results[1].SourceURL)
	assert.Equal(t, "http://broken.com", results[1].TargetURL)
	assert.Equal(t, 404, results[1].Status)
	assert.Equal(t, "", results[1].Error)
	assert.Equal(t, true, results[1].IsExternal)
}

// TestExportResults tests the ExportResults function
func TestExportResults(t *testing.T) {
	teardown := setupTest()
	defer teardown()

	// Prepare test output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test with unsupported format
	ExportResults("xml")

	// Restore stdout and read output
	_ = w.Close()
	os.Stdout = oldStdout
	_, err := io.Copy(&buf, r)
	require.NoError(t, err, "Failed to read output")
	output := buf.String()

	// Verify error message for unsupported format
	assert.Contains(t, output, "Unsupported format: xml. Use csv, json, or html")
}

// TestExportToHTML tests the HTML export functionality specifically
func TestExportToHTML(t *testing.T) {
	teardown := setupTest()
	defer teardown()

	// Set up test data with various scenarios
	model.Results = []model.LinkResult{
		{
			SourceURL:  SOURCE_URL,
			TargetURL:  "http://good.com",
			Status:     200,
			Error:      "",
			IsExternal: true,
		},
		{
			SourceURL:  SOURCE_URL,
			TargetURL:  "http://redirect.com",
			Status:     301,
			Error:      "",
			IsExternal: true,
		},
		{
			SourceURL:  SOURCE_URL,
			TargetURL:  "http://broken.com",
			Status:     404,
			Error:      "",
			IsExternal: true,
		},
		{
			SourceURL:  SOURCE_URL,
			TargetURL:  "http://error.com",
			Status:     0,
			Error:      "Connection failed",
			IsExternal: false,
		},
	}

	// Configure to show all links (not just errors)
	originalShowAll := model.ShowAll
	model.ShowAll = true
	defer func() { model.ShowAll = originalShowAll }()

	// Test HTML export
	ExportResults("html")

	// Verify file was created
	_, err := os.Stat("deadlinkr-report.html")
	require.NoError(t, err, "HTML file should be created")

	// Read and verify file content
	content, err := os.ReadFile("deadlinkr-report.html")
	require.NoError(t, err)

	htmlContent := string(content)

	// Verify HTML structure
	assert.Contains(t, htmlContent, "<!DOCTYPE html>")
	assert.Contains(t, htmlContent, "<title>DeadLinkr Report</title>")
	assert.Contains(t, htmlContent, "<h1>DeadLinkr Report</h1>")
	
	// Verify statistics
	assert.Contains(t, htmlContent, "Total links checked: 4")
	assert.Contains(t, htmlContent, "Broken links found: 2") // 404 + error

	// Verify table headers
	assert.Contains(t, htmlContent, "<th>Source URL</th>")
	assert.Contains(t, htmlContent, "<th>Target URL</th>")
	assert.Contains(t, htmlContent, "<th>Status</th>")
	assert.Contains(t, htmlContent, "<th>Error</th>")
	assert.Contains(t, htmlContent, "<th>Type</th>")

	// Verify link data is present
	assert.Contains(t, htmlContent, "http://good.com")
	assert.Contains(t, htmlContent, "http://broken.com")
	assert.Contains(t, htmlContent, "Connection failed")

	// Verify CSS classes are applied
	assert.Contains(t, htmlContent, "class=\"good\"")
	assert.Contains(t, htmlContent, "class=\"warning\"")
	assert.Contains(t, htmlContent, "class=\"error\"")

	// Verify link types
	assert.Contains(t, htmlContent, "External")
	assert.Contains(t, htmlContent, "Internal")
}

// TestExportHTMLWithFilters tests HTML export with display filters
func TestExportHTMLWithFilters(t *testing.T) {
	teardown := setupTest()
	defer teardown()

	// Set up test data
	model.Results = []model.LinkResult{
		{
			SourceURL:  SOURCE_URL,
			TargetURL:  "http://internal.com",
			Status:     200,
			Error:      "",
			IsExternal: false,
		},
		{
			SourceURL:  SOURCE_URL,
			TargetURL:  "http://external.com",
			Status:     404,
			Error:      "",
			IsExternal: true,
		},
	}

	t.Run("Display only external links", func(t *testing.T) {
		model.DisplayOnlyExternal = true
		model.ShowAll = true
		defer func() {
			model.DisplayOnlyExternal = false
			model.ShowAll = false
		}()

		ExportResults("html")

		content, err := os.ReadFile("deadlinkr-report.html")
		require.NoError(t, err)
		htmlContent := string(content)

		// Should contain external link
		assert.Contains(t, htmlContent, "http://external.com")
		// Should not contain internal link (filtered out)
		assert.NotContains(t, htmlContent, "http://internal.com")
	})

	t.Run("Display only errors", func(t *testing.T) {
		model.ShowAll = false
		model.DisplayOnlyExternal = false
		defer func() {
			model.ShowAll = false
		}()

		ExportResults("html")

		content, err := os.ReadFile("deadlinkr-report.html")
		require.NoError(t, err)
		htmlContent := string(content)

		// Should contain error link (404)
		assert.Contains(t, htmlContent, "http://external.com")
		// Should not contain good link (200)
		assert.NotContains(t, htmlContent, "http://internal.com")
	})
}

// TestExportHTMLWithCustomOutput tests HTML export with custom output path
func TestExportHTMLWithCustomOutput(t *testing.T) {
	teardown := setupTest()
	defer teardown()

	// Set custom output path
	model.Output = "custom-report.html"
	defer func() {
		model.Output = ""
		_ = os.Remove("custom-report.html")
	}()

	model.Results = []model.LinkResult{
		{
			SourceURL:  SOURCE_URL,
			TargetURL:  "http://test.com",
			Status:     200,
			Error:      "",
			IsExternal: false,
		},
	}

	// Configure to show all links (not just errors)
	originalShowAll := model.ShowAll
	model.ShowAll = true
	defer func() { model.ShowAll = originalShowAll }()

	ExportResults("html")

	// Verify custom file was created
	_, err := os.Stat("custom-report.html")
	require.NoError(t, err, "Custom HTML file should be created")

	// Verify content
	content, err := os.ReadFile("custom-report.html")
	require.NoError(t, err)
	assert.Contains(t, string(content), "http://test.com")
}
