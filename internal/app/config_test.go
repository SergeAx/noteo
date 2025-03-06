package app

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Save original environment variables
	envVars := []string{"NOTEO_BOT_TOKEN", "NOTEO_PORT", "NOTEO_LOG_FORMAT", "NOTEO_LOG_LEVEL"}
	oldEnvVars := make(map[string]string)
	for _, env := range envVars {
		oldEnvVars[env] = os.Getenv(env)
	}

	// Restore original environment variables after test
	defer func() {
		for env, val := range oldEnvVars {
			if val == "" {
				os.Unsetenv(env)
			} else {
				os.Setenv(env, val)
			}
		}
	}()

	t.Run("Test with required environment variables", func(t *testing.T) {
		// Clear all environment variables for testing
		for _, env := range envVars {
			os.Unsetenv(env)
		}

		// Set environment variables for this test
		os.Setenv("NOTEO_BOT_TOKEN", "test-token")
		os.Setenv("NOTEO_PORT", "9090")
		os.Setenv("NOTEO_LOG_FORMAT", "text")
		os.Setenv("NOTEO_LOG_LEVEL", "debug")
		
		// Reset Viper to ensure a clean state
		viper.Reset()
		
		// Load config
		config, err := LoadConfig()
		
		// Verify
		require.NoError(t, err)
		assert.Equal(t, "test-token", config.BotToken)
		assert.Equal(t, 9090, config.Port)
		assert.Equal(t, "text", config.LogFormat)
		assert.Equal(t, "debug", config.LogLevel)
	})

	t.Run("Test with missing required BOT_TOKEN", func(t *testing.T) {
		// Clear all environment variables for testing
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		
		// Reset Viper to ensure a clean state
		viper.Reset()
		
		// Load config
		config, err := LoadConfig()
		
		// Verify
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "NOTEO_BOT_TOKEN is required")
	})

	t.Run("Test with default values", func(t *testing.T) {
		// Clear all environment variables for testing
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		
		// Set only the required BOT_TOKEN
		os.Setenv("NOTEO_BOT_TOKEN", "test-token")
		
		// Reset Viper to ensure a clean state
		viper.Reset()
		
		// Load config
		config, err := LoadConfig()
		
		// Verify
		require.NoError(t, err)
		assert.Equal(t, "test-token", config.BotToken)
		assert.Equal(t, 8080, config.Port) // Default value
		assert.Equal(t, "json", config.LogFormat) // Default value
		assert.Equal(t, "info", config.LogLevel) // Default value
	})
	
	t.Run("Test with invalid PORT value", func(t *testing.T) {
		// Clear all environment variables for testing
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		
		// Set required BOT_TOKEN and invalid PORT (0 is invalid)
		os.Setenv("NOTEO_BOT_TOKEN", "test-token")
		os.Setenv("NOTEO_PORT", "0")
		
		// Reset Viper to ensure a clean state
		viper.Reset()
		
		// Load config
		config, err := LoadConfig()
		
		// Verify
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "invalid port value")
	})
	
	t.Run("Test with non-numeric PORT value", func(t *testing.T) {
		// Clear all environment variables for testing
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		
		// Set required BOT_TOKEN and non-numeric PORT
		os.Setenv("NOTEO_BOT_TOKEN", "test-token")
		os.Setenv("NOTEO_PORT", "invalid-port")
		
		// Reset Viper to ensure a clean state
		viper.Reset()
		
		// Load config
		config, err := LoadConfig()
		
		// Verify
		require.Error(t, err)
		assert.Nil(t, config)
		// Viper will convert non-numeric to 0, which will then fail our port validation
		assert.Contains(t, err.Error(), "invalid port value")
	})
	
	t.Run("Test with invalid LOG_FORMAT value", func(t *testing.T) {
		// Clear all environment variables for testing
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		
		// Set required BOT_TOKEN and invalid LOG_FORMAT
		os.Setenv("NOTEO_BOT_TOKEN", "test-token")
		os.Setenv("NOTEO_LOG_FORMAT", "invalid-format")
		
		// Reset Viper to ensure a clean state
		viper.Reset()
		
		// Load config
		config, err := LoadConfig()
		
		// Verify
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "invalid log format")
	})
	
	t.Run("Test with invalid LOG_LEVEL value", func(t *testing.T) {
		// Clear all environment variables for testing
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		
		// Set required BOT_TOKEN and invalid LOG_LEVEL
		os.Setenv("NOTEO_BOT_TOKEN", "test-token")
		os.Setenv("NOTEO_LOG_LEVEL", "invalid-level")
		
		// Reset Viper to ensure a clean state
		viper.Reset()
		
		// Load config
		config, err := LoadConfig()
		
		// Verify
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "invalid log level")
	})
	
	t.Run("Test case insensitivity for LOG_FORMAT and LOG_LEVEL", func(t *testing.T) {
		// Clear all environment variables for testing
		for _, env := range envVars {
			os.Unsetenv(env)
		}
		
		// Set environment variables with mixed case
		os.Setenv("NOTEO_BOT_TOKEN", "test-token")
		os.Setenv("NOTEO_LOG_FORMAT", "JSON")
		os.Setenv("NOTEO_LOG_LEVEL", "DEBUG")
		
		// Reset Viper to ensure a clean state
		viper.Reset()
		
		// Load config
		config, err := LoadConfig()
		
		// Verify
		require.NoError(t, err)
		assert.Equal(t, "json", config.LogFormat) // Should be converted to lowercase
		assert.Equal(t, "debug", config.LogLevel) // Should be converted to lowercase
	})
}
