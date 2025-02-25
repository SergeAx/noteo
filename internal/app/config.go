package app

import (
	"fmt"

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
	viper.SetDefault("NOTEO_PORT", 8080)
	viper.SetDefault("NOTEO_LOG_FORMAT", "JSON")
	viper.SetDefault("NOTEO_LOG_LEVEL", "INFO")

	// Setup environment variables
	viper.SetEnvPrefix("NOTEO")
	viper.AutomaticEnv()

	// Configure and read .env file
	viper.SetConfigName(".env")
	viper.SetConfigType("dotenv")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config file: %w", err)
		}
	}

	// Required configs
	if !viper.IsSet("NOTEO_BOT_TOKEN") {
		return nil, fmt.Errorf("NOTEO_BOT_TOKEN is required")
	}

	return &Config{
		BotToken:  viper.GetString("NOTEO_BOT_TOKEN"),
		Port:      viper.GetInt("NOTEO_PORT"),
		LogFormat: viper.GetString("NOTEO_LOG_FORMAT"),
		LogLevel:  viper.GetString("NOTEO_LOG_LEVEL"),
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
