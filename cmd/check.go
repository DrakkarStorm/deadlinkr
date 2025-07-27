package cmd

import (
	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/DrakkarStorm/deadlinkr/utils"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check [url]",
	Short: "Check a single page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pageURL := args[0]

		// Initialize
		model.Results = []model.LinkResult{}

		logger.Debugf("Checking links on %s", pageURL)

		// Check single page without recursion
		utils.CheckLinks(pageURL, pageURL)

		logger.Infof("Check complete. Found %d links, %d broken.", len(model.Results), utils.CountBrokenLinks())

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
	rootCmd.AddCommand(checkCmd)

}
