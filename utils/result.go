package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/EnzoDechaene/deadlinkr/model"
)

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

func ExportResults(format string) {
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

func exportToCSV() {
	file, err := os.Create("deadlinkr-report.csv")
	if err != nil {
		fmt.Printf("Error creating CSV file: %s\n", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"Source URL", "Target URL", "Status", "Error", "Is External"}); err != nil {
		fmt.Printf("Error writing CSV header: %s\n", err)
		return
	}

	// Write data
	for _, result := range model.Results {
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
			fmt.Printf("Error writing CSV row: %s\n", err)
			return
		}

	}

	fmt.Println("Report exported to deadlinkr-report.csv")
}

func exportToJSON() {
	file, err := os.Create("deadlinkr-report.json")
	if err != nil {
		fmt.Printf("Error creating JSON file: %s\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(model.Results); err != nil {
		fmt.Printf("Error encoding JSON: %s\n", err)
		return
	}

	fmt.Println("Report exported to deadlinkr-report.json")
}

func exportToHTML() {
	file, err := os.Create("deadlinkr-report.html")
	if err != nil {
		fmt.Printf("Error creating HTML file: %s\n", err)
		return
	}
	defer file.Close()

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
		rowClass := "good"
		if result.Status >= 400 || result.Error != "" {
			rowClass = "error"
		} else if result.Status >= 300 {
			rowClass = "warning"
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
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("Report exported to deadlinkr-report.html")
}
