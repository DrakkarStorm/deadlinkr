package cmd

import (
	"fmt"

	"github.com/EnzoDechaene/deadlinkr/model"
	"github.com/EnzoDechaene/deadlinkr/utils"
	"github.com/spf13/cobra"
)

// deadlinkr export --format=csv/json/html - Exporter les résultats
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Exporter les résultats",
	Run: func(cmd *cobra.Command, args []string) {
		if len(model.Results) == 0 {
			fmt.Println("No results available. Run a scan first.")
			return
		}

		utils.ExportResults(model.Format)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().String("format", "csv", "Export format (csv, json, html)")
}
