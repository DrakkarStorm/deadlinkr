package cmd

import (
	"fmt"
	"os"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deadlinkr",
	Short: "Deadlinkr is a tool to check for broken links in a website",
	Long:  `Deadlinkr is a tool to check for broken links in a website. It can be used to check a single page or a whole website. It supports concurrent requests and can export the results in various formats.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVar(&model.Depth, "depth", 1, "Maximum crawl depth")
	rootCmd.PersistentFlags().IntVar(&model.Concurrency, "concurrency", 20, "Number of concurrent requests")
	rootCmd.PersistentFlags().IntVar(&model.Timeout, "timeout", 30, "Request timeout in seconds")
	rootCmd.PersistentFlags().BoolVar(&model.IgnoreExternal, "ignore-external", true, "Ignore external links")
	rootCmd.PersistentFlags().BoolVar(&model.OnlyExternal, "only-external", false, "Check only external links")
	rootCmd.PersistentFlags().StringVar(&model.UserAgent, "user-agent", "DeadLinkr/1.0", "Custom user agent")
	rootCmd.PersistentFlags().StringVar(&model.IncludePattern, "include-pattern", "", "Only include URLs matching this regex")
	rootCmd.PersistentFlags().StringVar(&model.ExcludePattern, "exclude-pattern", "", "Exclude URLs matching this regex")
}
