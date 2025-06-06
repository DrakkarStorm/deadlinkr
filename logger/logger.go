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

func InitLogger(level LogLevel) {
	logLevel = level
	var logFilePath string

	switch runtime.GOOS {
	case "linux", "darwin":
		logFilePath = "/var/log/deadlinkr.log"
	case "windows":
		logFilePath = filepath.Join(os.Getenv("LOCALAPPDATA"), "deadlinkr", "deadlinkr.log")
	default:
		// For other operating systems, use a default directory
		logFilePath = filepath.Join(os.TempDir(), "deadlinkr.log")
	}

	logFile, err := os.Create(logFilePath)
	if err != nil {
		// If we can't create the log file in the specified directory, fall back to the current directory.
		logFilePath = "deadlinkr.log"
		logFile, err = os.Create(logFilePath)
		if err != nil {
			log.Fatalf("Failed to create log file: %v", err)
		}
	}
	fmt.Println("Logging to:", logFilePath)

	logger = log.New(logFile, "", log.LstdFlags)
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func Log(level LogLevel, format string, a ...any) {
	if level >= logLevel {
		logger.Printf("%s - %s", logLevels[level], fmt.Sprintf(format, a...))
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
