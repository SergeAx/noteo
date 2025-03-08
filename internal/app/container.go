package app

import (
	"fmt"

	"go.uber.org/dig"

	"gitlab.com/trum/noteo/internal/app/api"
	"gitlab.com/trum/noteo/internal/app/bot"
	"gitlab.com/trum/noteo/internal/app/db"
	"gitlab.com/trum/noteo/internal/app/queue"
	"gitlab.com/trum/noteo/internal/domain"
)

type Container struct {
	container *dig.Container
	err       error
}

func (c *Container) provide(constructor interface{}, description string, iface ...interface{}) {
	if c.err != nil {
		return
	}
	var opts []dig.ProvideOption
	if len(iface) > 0 {
		opts = append(opts, dig.As(iface...))
	}

	if err := c.container.Provide(constructor, opts...); err != nil {
		c.err = fmt.Errorf("failed to provide %s: %w", description, err)
	}
}

func (c *Container) Invoke(function interface{}) error {
	return c.container.Invoke(function)
}

// NewContainer creates and configures the DI container
func NewContainer(cfg *Config) (*Container, error) {
	c := &Container{
		container: dig.New(),
	}

	// Provide app config
	c.provide(func() (*Config, error) {
		if cfg == nil {
			return nil, fmt.Errorf("config is nil")
		}
		return cfg, nil
	}, "app config")

	// Provide service configs
	c.provide(NewBotConfig, "bot config")
	c.provide(NewAPIConfig, "api config")
	c.provide(NewDBConfig, "db config")
	c.provide(NewQueueConfig, "queue config")

	// Database
	c.provide(db.NewDB, "database")

	// Repositories
	c.provide(db.NewProjectRepository, "project repository", new(domain.ProjectRepository))
	c.provide(db.NewSubscriptionRepository, "subscription repository", new(domain.SubscriptionRepository))

	// Domain services
	c.provide(domain.NewProjectService, "project service")
	c.provide(domain.NewSubscriptionService, "subscription service")

	// App services
	c.provide(api.NewService, "api service")

	c.provide(bot.NewStateManager, "state manager")
	c.provide(bot.NewService, "bot service")

	if c.err != nil {
		return nil, c.err
	}

	return c, nil
}
