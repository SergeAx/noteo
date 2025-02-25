package app

import (
	"gitlab.com/trum/noteo/internal/app/api"
	"gitlab.com/trum/noteo/internal/app/bot"
)

type App struct {
	cfg       *Config
	container *Container
}

func New() *App {
	return &App{}
}

func (a *App) Init() error {
	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	a.cfg = cfg

	// Initialize logger
	if err := InitLogger(a.cfg); err != nil {
		return err
	}

	// Create and configure DI container
	container, err := NewContainer(a.cfg)
	if err != nil {
		return err
	}
	a.container = container

	return nil
}

func (a *App) Run() error {
	return a.container.Invoke(func(
		apiService *api.Service,
		botService *bot.Service,
	) error {
		go botService.Start()
		apiService.Start()
		return nil
	})
}
