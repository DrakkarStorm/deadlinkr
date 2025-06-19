package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deadlinkr",
	Short: "Deadlinkr is a tool to check for broken links in a website",
	Long:  `Deadlinkr is a tool to check for broken links in a website. It can be used to check a single page or a whole website. It supports concurrent requests and can export the results in various formats.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This function is executed before each command, after the flags have been parsed

		model.TimeExecution = time.Now()

        if model.LogLevel == "" {
            model.LogLevel = "info"
        }
        fmt.Println("Initializing logger with level:", model.LogLevel)
        logger.InitLogger(model.LogLevel)
    },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	defer logger.CloseLogger()

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&model.Quiet, "quiet", false, "Disable output")

    rootCmd.PersistentFlags().StringVar(&model.LogLevel, "log-level", "info", "Log level (debug, info, warn, error, fatal)")

	rootCmd.PersistentFlags().IntVar(&model.Timeout, "timeout", 10, "Request timeout in seconds")

	rootCmd.PersistentFlags().BoolVar(&model.OnlyInternal, "only-internal", false, "Check only internal links")

	rootCmd.PersistentFlags().StringVar(&model.UserAgent, "user-agent", "DeadLinkr/1.0", "Custom user agent")

	rootCmd.PersistentFlags().StringVar(&model.IncludePattern, "include-pattern", "", "Only include URLs matching this regex")
	rootCmd.PersistentFlags().StringVar(&model.ExcludePattern, "exclude-pattern", "", "Exclude URLs matching this regex")

	// for sre-docs you can use "div.md-sidebar__scrollwrap a[href]" to skip the menu on left
	rootCmd.PersistentFlags().StringVar(&model.ExcludeHtmlTags, "exclude-html-tags", "", "Exclude specific HTML tags separated by commas")

	rootCmd.PersistentFlags().BoolVar(&model.DisplayOnlyError, "display-only-error", true, "Display only error")
	rootCmd.PersistentFlags().BoolVar(&model.DisplayOnlyExternal, "display-only-external", false, "Display only external")

	rootCmd.PersistentFlags().StringVar(&model.Output, "output", "", "The path of output file (csv or json or html)")
}
