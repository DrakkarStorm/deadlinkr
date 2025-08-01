package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version will be set during build with ldflags
	Version = "dev"
	// Commit will be set during build with ldflags  
	Commit = "unknown"
	// BuildDate will be set during build with ldflags
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version, commit, build date and Go runtime information for deadlinkr.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("deadlinkr version %s\n", Version)
		fmt.Printf("  commit:     %s\n", Commit)
		fmt.Printf("  build date: %s\n", BuildDate)
		fmt.Printf("  go version: %s\n", runtime.Version())
		fmt.Printf("  platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}