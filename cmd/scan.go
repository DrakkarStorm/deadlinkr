package cmd

import (
	"github.com/DrakkarStorm/deadlinkr/logger"
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

		logger.Debugf("Starting scan of %s with depth %d", baseURL, model.Depth)

		// Start crawling
		utils.Crawl(baseURL, baseURL, 0)

		// Wait for all crawling to complete
		model.Wg.Wait()

		logger.Infof("Scan complete. Found %d links, %d broken.\n", len(model.Results), utils.CountBrokenLinks())

		logger.Debugf("Exporting results to %s", model.Format)
		if model.Format != "" {
			utils.ExportResults(model.Format)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Define a flag for the export format
	scanCmd.Flags().StringVar(&model.Format, "format", "html", "Export format (csv, json, html)")
	scanCmd.PersistentFlags().IntVar(&model.Depth, "depth", 1, "Maximum crawl depth")

}