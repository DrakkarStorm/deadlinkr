package cmd

import (
	"fmt"

	"github.com/EnzoDechaene/deadlinkr/model"
	"github.com/EnzoDechaene/deadlinkr/utils"
	"github.com/spf13/cobra"
)

// deadlinkr report - Afficher les résultats du dernier scan
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Afficher les résultats du dernier scan",
	Run: func(cmd *cobra.Command, args []string) {
		if len(model.Results) == 0 {
			fmt.Println("No results available. Run a scan first.")
			return
		}

		utils.DisplayResults()
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
