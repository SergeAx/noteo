package app

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// InitLogger initializes the global logger based on configuration
func InitLogger(cfg *Config) error {
	// Parse and validate log level
	logLevel, err := parseLogLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	var handler slog.Handler
	logFormat := strings.ToUpper(cfg.LogFormat)

	switch logFormat {
	case "TEXT":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	case "JSON":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})
	default:
		return fmt.Errorf("invalid log format: %s (must be either 'JSON' or 'TEXT')", logFormat)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return nil
}

func parseLogLevel(level string) (slog.Level, error) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("must be one of: DEBUG, INFO, WARN, ERROR. Got: %s", level)
	}
}
