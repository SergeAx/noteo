package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"

	"gitlab.com/trum/noteo/internal/app/api"
	"gitlab.com/trum/noteo/internal/app/bot"
	"gitlab.com/trum/noteo/internal/app/db"
	"gitlab.com/trum/noteo/internal/app/queue"
)

type Config struct {
	BotToken  string
	Port      int
	LogFormat string
	LogLevel  string
	DBDSN     string
}

// LoadConfig initializes and returns the application configuration
func LoadConfig() (*Config, error) {
	// Set defaults before anything else
	viper.SetDefault("PORT", 8080)
	viper.SetDefault("LOG_FORMAT", "json")
	viper.SetDefault("LOG_LEVEL", "info")

	// Setup environment variables
	viper.SetEnvPrefix("NOTEO")
	viper.AutomaticEnv()

	// Required configs
	if !viper.IsSet("BOT_TOKEN") {
		return nil, fmt.Errorf("NOTEO_BOT_TOKEN is required")
	}

	if !viper.IsSet("DB_DSN") {
		return nil, fmt.Errorf("NOTEO_DB_DSN is required")
	}

	// Get port and validate
	port := viper.GetInt("PORT")
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port value: %d (must be between 1 and 65535)", port)
	}

	// Get log format and validate
	logFormat := strings.ToLower(strings.TrimSpace(viper.GetString("LOG_FORMAT")))
	if logFormat != "json" && logFormat != "text" {
		return nil, fmt.Errorf("invalid log format: %s (must be 'json' or 'text')", logFormat)
	}

	// Get log level and validate
	logLevel := strings.ToLower(strings.TrimSpace(viper.GetString("LOG_LEVEL")))
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[logLevel] {
		return nil, fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error)", logLevel)
	}

	return &Config{
		BotToken:  strings.TrimSpace(viper.GetString("BOT_TOKEN")),
		Port:      port,
		LogFormat: logFormat,
		LogLevel:  logLevel,
		DBDSN:     strings.TrimSpace(viper.GetString("DB_DSN")),
	}, nil
}

// NewBotConfig creates bot-specific configuration
func NewBotConfig(cfg *Config) *bot.Config {
	return &bot.Config{
		Token: cfg.BotToken,
	}
}

// NewAPIConfig creates API-specific configuration
func NewAPIConfig(cfg *Config) *api.Config {
	return &api.Config{
		Port:         cfg.Port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

// NewDBConfig creates database-specific configuration
func NewDBConfig(cfg *Config) *db.Config {
	return &db.Config{
		DSN: cfg.DBDSN,
	}
}

// NewQueueConfig creates a new queue configuration
func NewQueueConfig(cfg *Config) *queue.Config {
	return &queue.Config{
		Capacity:          1000,
		InitialRetryDelay: 1 * time.Second,
		MaxRetryDelay:     1 * time.Minute,
		MaxRetries:        10,
	}
}
