package cmd

import (
	"fmt"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/DrakkarStorm/deadlinkr/utils"
	"github.com/spf13/cobra"
)

// deadlinkr scan [url] - Scanner un site web complet
var scanCmd = &cobra.Command{
	Use:   "scan [url]",
	Short: "Scanner un site web complet",
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

		if model.Format != "" {
			fmt.Println("Exporting results to", model.Format)
			utils.ExportResults(model.Format)
		} else {
			utils.DisplayResults()
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVar(&model.Format, "format", "html", "Export format (csv, json, html)")
}
