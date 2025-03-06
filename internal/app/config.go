package app

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

	"gitlab.com/trum/noteo/internal/app/api"
	"gitlab.com/trum/noteo/internal/app/bot"
)

type Config struct {
	BotToken  string
	Port      int
	LogFormat string
	LogLevel  string
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

	// Get port and validate
	port := viper.GetInt("PORT")
	if port <= 0 || port > 65535 {
		return nil, fmt.Errorf("invalid port value: %d (must be between 1 and 65535)", port)
	}

	// Get log format and validate
	logFormat := strings.ToLower(viper.GetString("LOG_FORMAT"))
	if logFormat != "json" && logFormat != "text" {
		return nil, fmt.Errorf("invalid log format: %s (must be 'json' or 'text')", logFormat)
	}

	// Get log level and validate
	logLevel := strings.ToLower(viper.GetString("LOG_LEVEL"))
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLogLevels[logLevel] {
		return nil, fmt.Errorf("invalid log level: %s (must be one of debug, info, warn, error, fatal)", logLevel)
	}

	return &Config{
		BotToken:  viper.GetString("BOT_TOKEN"),
		Port:      port,
		LogFormat: logFormat,
		LogLevel:  logLevel,
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
		Port: cfg.Port,
	}
}
