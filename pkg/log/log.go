// Package log provides a simple logging facility with support for different log levels
// and multiple output destinations (stdout/stderr and file).
//
// Log Levels:
//   - DEBUG: Detailed information for debugging (controlled by RUNSHELL_DEBUG env var)
//   - INFO: General operational information
//   - ERROR: Error conditions that should be addressed
//
// Environment Variables:
//   - RUNSHELL_DEBUG: Enable debug logging when set to "1" or "true"
//   - RUNSHELL_LOG_FILE: Path to log file (e.g., "/var/log/runshell/app.log")
//
// Basic Usage:
//
//	log.Info("Server starting on port %d", 8080)
//	log.Error("Failed to connect: %v", err)
//	if err := doSomething(); err != nil {
//	    log.Error("Operation failed: %v", err)
//	}
//
// Debug Logging:
//
//	// Only prints if RUNSHELL_DEBUG is enabled
//	log.Debug("Processing request: %+v", req)
//
// File Output:
//
// Method 1 - Using environment variable:
//
//	export RUNSHELL_LOG_FILE=/var/log/runshell/app.log
//
// Method 2 - Programmatically:
//
//	file, err := os.OpenFile("/var/log/runshell/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
//	if err != nil {
//	    log.Error("Failed to open log file: %v", err)
//	    return err
//	}
//	defer file.Close()
//	log.SetWriter(file)
//
// When both console and file output are enabled (by either method),
// logs will be written to both destinations simultaneously.
package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	// debugEnabled controls whether debug logs are printed
	debugEnabled bool

	// loggers for different levels
	debugLogger *log.Logger
	infoLogger  *log.Logger
	errorLogger *log.Logger

	// default writers
	defaultStdout = os.Stdout
	defaultStderr = os.Stderr

	// custom writer for additional output (e.g., file)
	customWriter io.Writer
)

func init() {
	// Check RUNSHELL_DEBUG environment variable
	debug := os.Getenv("RUNSHELL_DEBUG")
	debugEnabled = debug == "1" || debug == "true"

	// Check RUNSHELL_LOG_FILE environment variable
	if logFile := os.Getenv("RUNSHELL_LOG_FILE"); logFile != "" {
		if err := setupFileLogger(logFile); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to setup log file: %v\n", err)
		}
	}

	// Initialize loggers with default writers
	setupLoggers()
}

// setupFileLogger creates log directory and opens log file
func setupFileLogger(logPath string) error {
	// Create log directory if it doesn't exist
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file with append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	SetWriter(file)
	return nil
}

// setupLoggers initializes or reinitializes the loggers
func setupLoggers() {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds

	// If custom writer is set, use multi-writer
	var debugOut, infoOut, errorOut io.Writer

	if customWriter != nil {
		debugOut = io.MultiWriter(defaultStdout, customWriter)
		infoOut = io.MultiWriter(defaultStdout, customWriter)
		errorOut = io.MultiWriter(defaultStderr, customWriter)
	} else {
		debugOut = defaultStdout
		infoOut = defaultStdout
		errorOut = defaultStderr
	}

	debugLogger = log.New(debugOut, "[DEBUG] ", flags)
	infoLogger = log.New(infoOut, "[INFO] ", flags)
	errorLogger = log.New(errorOut, "[ERROR] ", flags)
}

// SetWriter sets a custom writer for logs (e.g., a file)
// This will be used in addition to standard output/error.
// The writer will receive all log levels (debug, info, and error).
// Note: The caller is responsible for closing the writer if needed.
func SetWriter(w io.Writer) {
	customWriter = w
	setupLoggers()
}

// Debug prints a debug message if RUNSHELL_DEBUG environment variable is set.
// The message format follows fmt.Printf conventions.
func Debug(format string, args ...interface{}) {
	if !debugEnabled {
		return
	}
	debugLogger.Printf(format, args...)
}

// Info prints an info message.
// The message format follows fmt.Printf conventions.
func Info(format string, args ...interface{}) {
	infoLogger.Printf(format, args...)
}

// Error prints an error message.
// The message format follows fmt.Printf conventions.
func Error(format string, args ...interface{}) {
	errorLogger.Printf(format, args...)
}
