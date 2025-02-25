package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/trum/noteo/internal/app/api"
	"gitlab.com/trum/noteo/internal/domain"
)

func TestNewContainer(t *testing.T) {
	// Create test config
	cfg := &Config{
		BotToken:  "test_token",
		Port:      8080,
		LogFormat: "text",
		LogLevel:  "info",
	}

	// Create container
	container, err := NewContainer(cfg)
	require.NoError(t, err)
	require.NotNil(t, container)

	// Test that all services can be resolved
	err = container.container.Invoke(func(
		apiService *api.Service,
		projectService *domain.ProjectService,
		subscriptionService *domain.SubscriptionService,
	) {
		assert.NotNil(t, apiService)
		assert.NotNil(t, projectService)
		assert.NotNil(t, subscriptionService)
	})
	require.NoError(t, err)
}

func TestNewContainer_InvalidConfig(t *testing.T) {
	// Test with nil config
	container, err := NewContainer(nil)
	require.NoError(t, err)
	require.NotNil(t, container)

	err = container.container.Invoke(func(
		apiService *api.Service,
	) {
		assert.Nil(t, apiService)
	})
	assert.Error(t, err)
}
