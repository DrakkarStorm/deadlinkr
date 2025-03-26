package cmd

import (
	"fmt"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/DrakkarStorm/deadlinkr/utils"
	"github.com/spf13/cobra"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan [url]",
	Short: "Scan a website for broken links",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		baseURL := args[0]

		// Reset global state
		model.Results = []model.LinkResult{}

		fmt.Printf("Starting scan of %s with depth %d\n", baseURL, model.Depth)

		// Start crawling
		utils.Crawl(baseURL, baseURL, 0)

		// Wait for all crawling to complete
		model.Wg.Wait()

		fmt.Printf("Scan complete. Found %d links, %d broken.\n", len(model.Results), utils.CountBrokenLinks())

		fmt.Println("Exporting results to", model.Format)
		if model.Format != "" {
			utils.ExportResults(model.Format)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Define a flag for the export format
	scanCmd.Flags().StringVar(&model.Format, "format", "html", "Export format (csv, json, html)")
}