package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/sergeax/noteo/internal/app/queue"
	"github.com/sergeax/noteo/internal/domain"
)

type Service struct {
	config              *Config
	messageQueue        *queue.Queue
	projectService      *domain.ProjectService
	subscriptionService *domain.SubscriptionService
	server              *http.Server
}

func NewService(cfg *Config, messageQueue *queue.Queue, projectService *domain.ProjectService, subscriptionService *domain.SubscriptionService) *Service {
	return &Service{
		config:              cfg,
		messageQueue:        messageQueue,
		projectService:      projectService,
		subscriptionService: subscriptionService,
	}
}

func (s *Service) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/notify", s.handleNotify)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
		Handler:      mux,
	}

	// Start server in a goroutine so it doesn't block
	errCh := make(chan error, 1)
	go func() {
		slog.Info("Starting API server", "port", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		return s.Stop()
	case err := <-errCh:
		return err
	}
}

// Stop gracefully shuts down the API server
func (s *Service) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("Shutting down API server")
	return s.server.Shutdown(ctx)
}

func (s *Service) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		http.Error(w, "Missing bearer token", http.StatusUnauthorized)
		return
	}

	var notification struct {
		Body string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate token
	project, err := s.projectService.GetByToken(token)
	if err != nil {
		slog.Error("Failed to get project by token", "error", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Get active subscriptions for the project
	subscriptions, err := s.subscriptionService.GetProjectSubscriptions(project.ID)
	if err != nil {
		slog.Error("Failed to get active subscriptions", "error", err, "projectId", project.ID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Send notification to all subscribers
	for _, sub := range subscriptions {
		if sub.Paused() {
			continue
		}
		if err := s.messageQueue.Put(domain.Message{
			UserID: sub.UserID,
			Text:   notification.Body,
			Muted:  sub.Muted,
		}); err != nil {
			if err == queue.ErrQueueFull {
				slog.Error("Message queue is full", "projectId", project.ID)
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				return
			}
			slog.Error("Failed to send notification", "error", err, "chatId", sub.UserID)
		}
	}

	w.WriteHeader(http.StatusOK)
}
