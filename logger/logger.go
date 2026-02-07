// logger/logger.go
package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/DrakkarStorm/deadlinkr/model"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

var logLevels = map[LogLevel]string{
	DebugLevel: "DEBUG",
	InfoLevel:  "INFO",
	WarnLevel:  "WARN",
	ErrorLevel: "ERROR",
	FatalLevel: "FATAL",
}

var logger *log.Logger
var logLevel LogLevel
var logFile *os.File

func InitLogger(level string) {
	if !model.Quiet {
		// Close any previously opened log file to avoid resource leaks
		CloseLogger()

		logLevel = StringToLogLevel(level)

		var logFilePath string
		var logDir string

		switch runtime.GOOS {
		case "linux", "darwin":
			logDir = "/var/log"
			logFilePath = "deadlinkr.log"
		case "windows":
			logDir = filepath.Join(os.Getenv("LOCALAPPDATA"), "deadlinkr")
			logFilePath = "deadlinkr.log"
		default:
			// For other operating systems, use a default directory
			logDir = os.TempDir()
			logFilePath = "deadlinkr.log"
		}

		// Try to create log file using os.Root for the specified directory
		root, err := os.OpenRoot(logDir)
		if err == nil {
			logFile, err = root.Create(logFilePath)
			if err == nil {
				fmt.Println("Logging to:", filepath.Join(logDir, logFilePath))
				logger = log.New(logFile, "", log.LstdFlags)
				_ = root.Close()
				return
			}
			_ = root.Close()
		}

		// If we can't create the log file in the specified directory, fall back to the current directory
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get working directory: %v", err)
		}

		root, err = os.OpenRoot(cwd)
		if err != nil {
			log.Fatalf("Failed to create root scope: %v", err)
		}
		defer func() {
			if err := root.Close(); err != nil {
				log.Printf("Error closing root scope: %v", err)
			}
		}()

		logFile, err = root.Create("deadlinkr.log")
		if err != nil {
			log.Fatalf("Failed to create log file: %v", err)
		}
		fmt.Println("Logging to:", filepath.Join(cwd, "deadlinkr.log"))

		logger = log.New(logFile, "", log.LstdFlags)
	}
}

func StringToLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	default:
		return InfoLevel
	}
}

func CloseLogger() {
	if logFile != nil {
		if err := logFile.Close(); err != nil {
			log.Printf("Error closing log file: %v", err)
		}
		logFile = nil
	}
}

func Log(level LogLevel, format string, a ...any) {
	if !model.Quiet {
		if level >= logLevel {
			logger.Printf("%s - %s", logLevels[level], fmt.Sprintf(format, a...))
		}
	}
}

func Debugf(format string, a ...any) {
	Log(DebugLevel, format, a...)
}

func Infof(format string, a ...any) {
	Log(InfoLevel, format, a...)
}

func Warnf(format string, a ...any) {
	Log(WarnLevel, format, a...)
}

func Errorf(format string, a ...any) {
	Log(ErrorLevel, format, a...)
}

func Fatalf(format string, a ...any) {
	Log(FatalLevel, format, a...)
	os.Exit(1)
}

func Durationf(format string, a ...any) {
	duration := time.Since(model.TimeExecution)
	Infof("%s - %s (durée : %v)", "Program Execution", fmt.Sprintf(format, a...), duration)
}
