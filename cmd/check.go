package cmd

import (
	"fmt"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/DrakkarStorm/deadlinkr/utils"
	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check [url]",
	Short: "Vérifier une seule page",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pageURL := args[0]

		// Initialize
		model.Results = []model.LinkResult{}

		fmt.Printf("Checking links on %s\n", pageURL)

		// Check single page without recursion
		utils.CheckLinks(pageURL, pageURL)

		fmt.Printf("Check complete. Found %d links, %d broken.\n", len(model.Results), utils.CountBrokenLinks())

		if model.Format != "" {
			utils.ExportResults(model.Format)
		} else {
			utils.DisplayResults()
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)

	// Define a flag for the export format
	checkCmd.Flags().String("format", "", "Export format (csv, json, html)")
}
