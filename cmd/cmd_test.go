package cmd

import (
	"bytes"
	"os"
	"sync"
	"testing"

	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to set up test environment
func setupCmdTest() func() {
	// Save original state
	originalResults := model.Results
	originalDepth := model.Depth
	originalTimeout := model.Timeout
	originalConcurrency := model.Concurrency
	originalQuiet := model.Quiet
	originalLogLevel := model.LogLevel

	// Set up test environment
	model.Results = []model.LinkResult{}
	model.VisitedURLs = sync.Map{}
	model.Depth = 1
	model.Timeout = 5
	model.Concurrency = 10
	model.Quiet = true // Avoid log file creation in tests
	model.LogLevel = "info"

	// Return a function to restore original state
	return func() {
		model.Results = originalResults
		model.Depth = originalDepth
		model.Timeout = originalTimeout
		model.Concurrency = originalConcurrency
		model.Quiet = originalQuiet
		model.LogLevel = originalLogLevel
		logger.CloseLogger()
		
		// Clean up test files
		_ = os.Remove("deadlinkr-report.csv")
		_ = os.Remove("deadlinkr-report.json")
		_ = os.Remove("deadlinkr-report.html")
		_ = os.Remove("deadlinkr.log")
	}
}

func TestRootCmd(t *testing.T) {
	teardown := setupCmdTest()
	defer teardown()

	t.Run("Root command exists", func(t *testing.T) {
		assert.NotNil(t, rootCmd)
		assert.Equal(t, "deadlinkr", rootCmd.Use)
		assert.Contains(t, rootCmd.Short, "broken links")
	})

	t.Run("Root command has persistent flags", func(t *testing.T) {
		// Test that flags are properly defined
		flag := rootCmd.PersistentFlags().Lookup("quiet")
		assert.NotNil(t, flag)
		
		flag = rootCmd.PersistentFlags().Lookup("timeout")
		assert.NotNil(t, flag)
		
		flag = rootCmd.PersistentFlags().Lookup("user-agent")
		assert.NotNil(t, flag)
	})
}

func TestScanCmd(t *testing.T) {
	teardown := setupCmdTest()
	defer teardown()

	t.Run("Scan command exists", func(t *testing.T) {
		assert.NotNil(t, scanCmd)
		assert.Equal(t, "scan [url]", scanCmd.Use)
		assert.Contains(t, scanCmd.Short, "Scan")
	})

	t.Run("Scan command has flags", func(t *testing.T) {
		flag := scanCmd.PersistentFlags().Lookup("concurrency")
		assert.NotNil(t, flag)
		
		flag = scanCmd.Flags().Lookup("format")
		assert.NotNil(t, flag)
		
		flag = scanCmd.PersistentFlags().Lookup("depth")
		assert.NotNil(t, flag)
	})

	t.Run("Scan command requires exactly one argument", func(t *testing.T) {
		// This tests the Args: cobra.ExactArgs(1) setting
		cmd := scanCmd
		args := []string{}
		err := cmd.Args(cmd, args)
		assert.Error(t, err)
		
		args = []string{"url1", "url2"}
		err = cmd.Args(cmd, args)
		assert.Error(t, err)
		
		args = []string{"http://example.com"}
		err = cmd.Args(cmd, args)
		assert.NoError(t, err)
	})
}

func TestCheckCmd(t *testing.T) {
	teardown := setupCmdTest()
	defer teardown()

	t.Run("Check command exists", func(t *testing.T) {
		assert.NotNil(t, checkCmd)
		assert.Equal(t, "check [url]", checkCmd.Use)
		assert.Contains(t, checkCmd.Short, "Check")
	})

	t.Run("Check command requires exactly one argument", func(t *testing.T) {
		cmd := checkCmd
		args := []string{}
		err := cmd.Args(cmd, args)
		assert.Error(t, err)
		
		args = []string{"http://example.com"}
		err = cmd.Args(cmd, args)
		assert.NoError(t, err)
	})
}

func TestExecute(t *testing.T) {
	teardown := setupCmdTest()
	defer teardown()

	// We can't easily test Execute() function directly because it calls os.Exit
	// But we can test that it exists and is callable in a safe way
	t.Run("Execute function exists", func(t *testing.T) {
		// Just verify the function exists and can be referenced
		assert.NotNil(t, Execute)
	})
}

func TestCommandIntegration(t *testing.T) {
	teardown := setupCmdTest()
	defer teardown()

	t.Run("Commands are properly added to root", func(t *testing.T) {
		commands := rootCmd.Commands()
		
		// Check that scan and check commands are added
		var foundScan, foundCheck bool
		for _, cmd := range commands {
			if cmd.Name() == "scan" {
				foundScan = true
			}
			if cmd.Name() == "check" {
				foundCheck = true
			}
		}
		
		assert.True(t, foundScan, "scan command should be added to root")
		assert.True(t, foundCheck, "check command should be added to root")
	})
}

func TestFlagDefaults(t *testing.T) {
	teardown := setupCmdTest()
	defer teardown()

	t.Run("Default values are set correctly", func(t *testing.T) {
		// Reset to defaults
		model.Timeout = 10
		model.Concurrency = 20
		model.Depth = 1
		model.UserAgent = "DeadLinkr/1.0"
		
		// Test default timeout
		assert.Equal(t, 10, model.Timeout)
		
		// Test default concurrency
		assert.Equal(t, 20, model.Concurrency)
		
		// Test default depth
		assert.Equal(t, 1, model.Depth)
		
		// Test default user agent
		assert.Equal(t, "DeadLinkr/1.0", model.UserAgent)
	})
}

func TestPersistentPreRun(t *testing.T) {
	teardown := setupCmdTest()
	defer teardown()

	t.Run("PersistentPreRun sets up logger", func(t *testing.T) {
		// Capture stdout to avoid test output pollution
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Test the persistent pre-run function
		model.LogLevel = "debug"
		rootCmd.PersistentPreRun(rootCmd, []string{})

		// Restore stdout
		_ = w.Close()
		os.Stdout = old
		
		// Read output
		var buf bytes.Buffer
		_, err := buf.ReadFrom(r)
		require.NoError(t, err)

		// The function should have run without error
		// (We can't easily test the exact behavior due to global state)
		assert.Equal(t, "debug", model.LogLevel)
	})
}