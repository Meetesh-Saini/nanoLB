package log

import (
	"fmt"
	"io"
	"nanoLB/internal/config"
	"os"

	"github.com/sirupsen/logrus"
)

// Global logger instance
var Logger *logrus.Logger

func Init() error {
	cfg := config.GetConfig()

	Logger = logrus.New()

	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level %s: %v", cfg.LogLevel, err)
	}
	Logger.SetLevel(level)

	// Set the log format based on the config
	switch cfg.LogFormat {
	case "json":
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		Logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	default:
		return fmt.Errorf("invalid log format %s, use 'json' or 'text'", cfg.LogFormat)
	}

	// Handle the output based on the config.LogOutput setting
	switch cfg.LogOutput {
	case "none":
		// Disable logging
		Logger.SetOutput(io.Discard) // Discards any log messages
	case "stdout":
		// Only log to stdout
		Logger.SetOutput(os.Stdout)
	case "file":
		// Only log to the file
		logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %v", cfg.LogFile, err)
		}
		Logger.SetOutput(logFile)
	case "both":
		// Log to both stdout and file
		logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %v", cfg.LogFile, err)
		}
		logOutput := io.MultiWriter(os.Stdout, logFile)
		Logger.SetOutput(logOutput)
	default:
		return fmt.Errorf("invalid log output type %s, use 'stdout', 'file', 'both', or 'none'", cfg.LogOutput)
	}

	return nil
}
