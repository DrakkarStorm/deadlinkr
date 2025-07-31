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

	rootCmd.PersistentFlags().IntVarP(&model.Timeout, "timeout", "t", 15, "Request timeout in seconds")

	rootCmd.PersistentFlags().BoolVar(&model.OnlyInternal, "only-internal", false, "Check only internal links")

	rootCmd.PersistentFlags().StringVar(&model.UserAgent, "user-agent", "DeadLinkr/1.0", "Custom user agent")

	rootCmd.PersistentFlags().StringVar(&model.IncludePattern, "include-pattern", "", "Only include URLs matching this regex")
	rootCmd.PersistentFlags().StringVar(&model.ExcludePattern, "exclude-pattern", "", "Exclude URLs matching this regex")

	rootCmd.PersistentFlags().StringVar(&model.ExcludeHtmlTags, "exclude-html-tags", "", "Exclude specific HTML tags separated by commas")

	rootCmd.PersistentFlags().BoolVar(&model.ShowAll, "show-all", false, "Show all links including working ones (default: only broken links)")
	rootCmd.PersistentFlags().BoolVar(&model.DisplayOnlyExternal, "only-external", false, "Show only external links")

	rootCmd.PersistentFlags().StringVarP(&model.Output, "output", "o", "", "Output file path (format auto-detected from extension: .csv, .json, .html)")
	rootCmd.PersistentFlags().StringVarP(&model.Format, "format", "f", "", "Export format (csv, json, html) - overrides auto-detection from output file")

	rootCmd.PersistentFlags().Float64Var(&model.RateLimitRequestsPerSecond, "rate-limit", 2.0, "Requests per second per domain")
	rootCmd.PersistentFlags().Float64Var(&model.RateLimitBurst, "rate-burst", 5.0, "Burst capacity for rate limiting")

	rootCmd.PersistentFlags().BoolVar(&model.OptimizeWithHeadRequests, "optimize-head", true, "Use HEAD requests when possible to reduce bandwidth")

	rootCmd.PersistentFlags().BoolVar(&model.CacheEnabled, "cache", true, "Enable intelligent caching of link check results")
	rootCmd.PersistentFlags().IntVar(&model.CacheSize, "cache-size", 1000, "Maximum number of entries in the cache")
	rootCmd.PersistentFlags().IntVar(&model.CacheTTLMinutes, "cache-ttl", 60, "Cache time-to-live in minutes")

	// Authentication flags
	rootCmd.PersistentFlags().StringVar(&model.AuthBasic, "auth-basic", "", "Basic authentication in 'user:password' format (or use DEADLINKR_AUTH_USER/DEADLINKR_AUTH_PASS env vars)")
	rootCmd.PersistentFlags().StringVar(&model.AuthBearer, "auth-bearer", "", "Bearer token authentication (or use DEADLINKR_AUTH_TOKEN env var)")
	rootCmd.PersistentFlags().StringArrayVar(&model.AuthHeaders, "auth-header", []string{}, "Custom authentication headers in 'Key: Value' format (can be used multiple times)")
	rootCmd.PersistentFlags().StringVar(&model.AuthCookies, "auth-cookies", "", "Cookie authentication string")
}
