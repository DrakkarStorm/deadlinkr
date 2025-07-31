package logger

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/DrakkarStorm/deadlinkr/model"
	"github.com/stretchr/testify/assert"
)

func TestStringToLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"debug", DebugLevel},
		{"info", InfoLevel},
		{"warn", WarnLevel},
		{"error", ErrorLevel},
		{"fatal", FatalLevel},
		{"unknown", InfoLevel}, // default
		{"", InfoLevel},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := StringToLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestInitLogger(t *testing.T) {
	// Save original state
	originalQuiet := model.Quiet
	defer func() {
		model.Quiet = originalQuiet
		CloseLogger()
		// Clean up test log files
		_ = os.Remove("deadlinkr.log")
		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			_ = os.Remove("/var/log/deadlinkr.log")
		}
		if runtime.GOOS == "windows" {
			logPath := filepath.Join(os.Getenv("LOCALAPPDATA"), "deadlinkr", "deadlinkr.log")
			_ = os.Remove(logPath)
		}
	}()

	t.Run("Logger initialized when not quiet", func(t *testing.T) {
		model.Quiet = false
		
		InitLogger("debug")
		
		assert.NotNil(t, logger)
		assert.Equal(t, DebugLevel, logLevel)
	})

	t.Run("Logger not initialized when quiet", func(t *testing.T) {
		model.Quiet = true
		logger = nil // Reset
		
		InitLogger("info")
		
		// When quiet, logger should remain nil
		assert.Nil(t, logger)
	})
}

func TestLogLevels(t *testing.T) {
	// Save original state
	originalQuiet := model.Quiet
	defer func() {
		model.Quiet = originalQuiet
		CloseLogger()
		_ = os.Remove("deadlinkr.log")
	}()

	model.Quiet = false
	InitLogger("warn") // Set level to warn

	// These should be safe to call even if they don't actually log
	// (since we can't easily capture file output in tests)
	t.Run("All log functions work", func(t *testing.T) {
		Debugf("debug message")   // Should not log (below warn level)
		Infof("info message")     // Should not log (below warn level)
		Warnf("warn message")     // Should log
		Errorf("error message")   // Should log
		// Note: Not testing Fatalf as it calls os.Exit(1)
	})
}

func TestDurationf(t *testing.T) {
	// Save original state
	originalQuiet := model.Quiet
	originalTimeExecution := model.TimeExecution
	defer func() {
		model.Quiet = originalQuiet
		model.TimeExecution = originalTimeExecution
		CloseLogger()
		_ = os.Remove("deadlinkr.log")
	}()

	model.Quiet = false
	InitLogger("info")
	
	// This should not panic
	Durationf("test duration")
}

func TestCloseLogger(t *testing.T) {
	// Save original state
	originalQuiet := model.Quiet
	defer func() {
		model.Quiet = originalQuiet
	}()

	t.Run("Close logger when file exists", func(t *testing.T) {
		model.Quiet = false
		InitLogger("info")
		
		// This should not panic
		CloseLogger()
	})

	t.Run("Close logger when file is nil", func(t *testing.T) {
		logFile = nil
		
		// This should not panic
		CloseLogger()
	})
}

func TestLogFormat(t *testing.T) {
	// Test that log level names are correct
	assert.Equal(t, "DEBUG", logLevels[DebugLevel])
	assert.Equal(t, "INFO", logLevels[InfoLevel])
	assert.Equal(t, "WARN", logLevels[WarnLevel])
	assert.Equal(t, "ERROR", logLevels[ErrorLevel])
	assert.Equal(t, "FATAL", logLevels[FatalLevel])
}

func TestLogFilePath(t *testing.T) {
	// Save original state
	originalQuiet := model.Quiet
	defer func() {
		model.Quiet = originalQuiet
		CloseLogger()
		_ = os.Remove("deadlinkr.log")
	}()

	model.Quiet = false
	
	// Test that logger initializes without crashing
	// (actual file path testing is complex due to permissions)
	InitLogger("info")
	
	assert.NotNil(t, logger)
}

func TestLogWithQuietMode(t *testing.T) {
	// Save original state
	originalQuiet := model.Quiet
	defer func() {
		model.Quiet = originalQuiet
	}()

	t.Run("No logging when quiet mode is on", func(t *testing.T) {
		model.Quiet = true
		logger = nil
		
		// These should not panic even when logger is nil
		Log(InfoLevel, "test message")
		Debugf("debug message")
		Infof("info message")
		Warnf("warn message")
		Errorf("error message")
		Durationf("duration message")
	})
}