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

		// Use optimized crawler by default
		err := utils.CrawlWithOptimizedServices(baseURL, baseURL, 0)
		if err != nil {
			logger.Errorf("Error during scan: %s", err)
			return
		}

		logger.Infof("Scan complete. Found %d links, %d broken.\n", len(model.Results), utils.CountBrokenLinks())

		// Auto-detect format from output file if not specified
		format := model.Format
		if format == "" && model.Output != "" {
			format = utils.DetectFormatFromOutput(model.Output)
		}

		logger.Debugf("Exporting results with format: %s, output: %s", format, model.Output)
		if format != "" || model.Output != "" {
			utils.ExportResults(format)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.PersistentFlags().IntVarP(&model.Concurrency, "concurrency", "c", 20, "Number of concurrent requests")
	scanCmd.PersistentFlags().IntVarP(&model.Depth, "depth", "d", 1, "Maximum crawl depth")

}
