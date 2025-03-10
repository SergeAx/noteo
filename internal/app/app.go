package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sergeax/noteo/internal/app/api"
	"github.com/sergeax/noteo/internal/app/bot"
	"github.com/sergeax/noteo/internal/app/queue"
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
		messageQueue *queue.Queue,
	) error {
		// Setup signal handling for graceful shutdown
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Channel to listen for OS signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Start message queue
		messageQueue.Start()

		// Start bot service
		go botService.Start()

		// Start API server in a goroutine
		go func() {
			if err := apiService.Start(ctx); err != nil {
				slog.Error("API server failed", "error", err)
				cancel()
			}
		}()

		// Wait for termination signal
		select {
		case <-sigChan:
			slog.Info("Received shutdown signal")
		case <-ctx.Done():
			slog.Info("Context canceled")
		}

		// Graceful shutdown
		slog.Info("Shutting down services")
		botService.Stop()
		if err := apiService.Stop(); err != nil {
			slog.Error("Error shutting down API service", "error", err)
		}
		messageQueue.Stop()
		slog.Info("Shutdown complete")

		return nil
	})
}
