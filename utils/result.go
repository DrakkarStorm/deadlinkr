package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// DisplayResults displays the results of the link check.
// example: DisplayResults() -> "Broken links: 2"
func DisplayResults() {
	brokenLinks := []model.LinkResult{}

	for _, result := range model.Results {
		if result.Status >= 400 || result.Error != "" {
			brokenLinks = append(brokenLinks, result)
		}
	}

	if len(brokenLinks) == 0 {
		fmt.Println("No broken links found!")
		return
	}

	fmt.Println("\nBroken links:")
	fmt.Println("=============")

	for _, link := range brokenLinks {
		if link.Error != "" {
			fmt.Printf("- %s (from %s): Error: %s\n", link.TargetURL, link.SourceURL, link.Error)
		} else {
			fmt.Printf("- %s (from %s): Status: %d\n", link.TargetURL, link.SourceURL, link.Status)
		}
	}
}

// DetectFormatFromOutput detects the format from the output file extension
func DetectFormatFromOutput(outputPath string) string {
	if outputPath == "" {
		return ""
	}
	
	ext := strings.ToLower(filepath.Ext(outputPath))
	switch ext {
	case ".csv":
		return "csv"
	case ".json":
		return "json"
	case ".html", ".htm":
		return "html"
	default:
		return ""
	}
}

// ExportResults exports the results of the link check to a file.
// Auto-detects format from output file extension if format is empty
func ExportResults(format string) {
	// Auto-detect format from output file extension if not specified
	if format == "" && model.Output != "" {
		format = DetectFormatFromOutput(model.Output)
		if format != "" {
			logger.Debugf("Auto-detected format '%s' from output file extension", format)
		}
	}
	
	// If still no format, default to displaying results
	if format == "" {
		DisplayResults()
		return
	}
	
	switch strings.ToLower(format) {
	case "csv":
		exportToCSV()
	case "json":
		exportToJSON()
	case "html":
		exportToHTML()
	default:
		fmt.Printf("Unsupported format: %s. Use csv, json, or html.\n", format)
	}
}

// exportToCSV exports the results to a CSV file.
func exportToCSV() {
	filename := "deadlinkr-report.csv"
	if model.Output != "" {
		filename = model.Output
	}

	// Create a root scoped to current working directory to prevent directory traversal
	cwd, err := os.Getwd()
	if err != nil {
		logger.Errorf("Error getting working directory: %s\n", err)
		return
	}

	root, err := os.OpenRoot(cwd)
	if err != nil {
		logger.Errorf("Error creating root scope: %s\n", err)
		return
	}
	defer func() {
		if err := root.Close(); err != nil {
			logger.Errorf("Error closing root scope: %s\n", err)
		}
	}()

	file, err := root.Create(filename)
	if err != nil {
		logger.Errorf("Error creating CSV file: %s\n", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Errorf("Error closing CSV file: %s\n", err)
		}
	}()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Source URL", "Target URL", "Status", "Error", "Is External"}); err != nil {
		logger.Errorf("Error writing CSV header: %s\n", err)
		return
	}

	// Write data
	for _, result := range model.Results {
		if (model.OnlyInternal && result.IsExternal || model.DisplayOnlyExternal && !result.IsExternal) {
			continue
		}

		isExternalStr := "false"
		if result.IsExternal {
			isExternalStr = "true"
		}

		if err := writer.Write([]string{
			result.SourceURL,
			result.TargetURL,
			fmt.Sprintf("%d", result.Status),
			result.Error,
			isExternalStr,
		}); err != nil {
			logger.Errorf("Error writing CSV row: %s\n", err)
			return
		}

	}

	logger.Debugf("Report exported to deadlinkr-report.csv")
}

// exportToJSON exports the results to a JSON file.
func exportToJSON() {
	filename := "deadlinkr-report.json"
	if model.Output != "" {
		filename = model.Output
	}

	// Create a root scoped to current working directory to prevent directory traversal
	cwd, err := os.Getwd()
	if err != nil {
		logger.Errorf("Error getting working directory: %s\n", err)
		return
	}

	root, err := os.OpenRoot(cwd)
	if err != nil {
		logger.Errorf("Error creating root scope: %s\n", err)
		return
	}
	defer func() {
		if err := root.Close(); err != nil {
			logger.Errorf("Error closing root scope: %s\n", err)
		}
	}()

	file, err := root.Create(filename)
	if err != nil {
		logger.Errorf("Error creating JSON file: %s\n", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Errorf("Error closing JSON file: %s\n", err)
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(model.Results); err != nil {
		logger.Errorf("Error encoding JSON: %s\n", err)
		return
	}

	logger.Debugf("Report exported to deadlinkr-report.json")
}

// exportToHTML exports the results to an HTML file.
func exportToHTML() {
	filename := "deadlinkr-report.html"
	if model.Output != "" {
		filename = model.Output
	}

	// Create a root scoped to current working directory to prevent directory traversal
	cwd, err := os.Getwd()
	if err != nil {
		logger.Errorf("Error getting working directory: %s\n", err)
		return
	}

	root, err := os.OpenRoot(cwd)
	if err != nil {
		logger.Errorf("Error creating root scope: %s\n", err)
		return
	}
	defer func() {
		if err := root.Close(); err != nil {
			logger.Errorf("Error closing root scope: %s\n", err)
		}
	}()

	file, err := root.Create(filename)
	if err != nil {
		logger.Errorf("Error creating HTML file: %s\n", err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Errorf("Error closing HTML file: %s\n", err)
		}
	}()

	// Create simple HTML report
	html := `<!DOCTYPE html>
<html>
<head>
    <title>DeadLinkr Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .error { background-color: #ffecec; }
        .warning { background-color: #fffaec; }
        .good { background-color: #efffec; }
    </style>
</head>
<body>
    <h1>DeadLinkr Report</h1>
    <p>Total links checked: ` + fmt.Sprintf("%d", len(model.Results)) + `</p>
    <p>Broken links found: ` + fmt.Sprintf("%d", CountBrokenLinks()) + `</p>

    <table>
        <tr>
            <th>Source URL</th>
            <th>Target URL</th>
            <th>Status</th>
            <th>Error</th>
            <th>Type</th>
        </tr>
`

	for _, result := range model.Results {
		if (model.OnlyInternal && result.IsExternal || model.DisplayOnlyExternal && !result.IsExternal) {
			continue
		}

		rowClass := "good"
		if result.Status >= 400 || result.Error != "" {
			rowClass = "error"
		} else if result.Status >= 300 {
			rowClass = "warning"
		}

		if rowClass == "good" {
			if !model.ShowAll {
				continue
			}
		}

		linkType := "Internal"
		if result.IsExternal {
			linkType = "External"
		}

		statusStr := fmt.Sprintf("%d", result.Status)
		if result.Status == 0 {
			statusStr = "Error"
		}

		html += `        <tr class="` + rowClass + `">
            <td>` + result.SourceURL + `</td>
            <td>` + result.TargetURL + `</td>
            <td>` + statusStr + `</td>
            <td>` + result.Error + `</td>
            <td>` + linkType + `</td>
        </tr>
`
	}

	html += `    </table>
</body>
</html>`

	_, err = file.WriteString(html)
	if err != nil {
		logger.Errorf("Error writing to file: %s", err.Error())
		return
	}
	logger.Debugf("Report exported to deadlinkr-report.html")
}
