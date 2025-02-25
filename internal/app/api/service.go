package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type Service struct {
	config *Config
}

func NewService(cfg *Config) *Service {
	return &Service{
		config: cfg,
	}
}

func (s *Service) Start() {
	http.HandleFunc("/api/notify", s.handleNotify)

	slog.Info("Starting API server", "port", s.config.Port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), nil); err != nil {
		slog.Error("Failed to start API server", "error", err)
		os.Exit(1)
	}
}

func (s *Service) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		http.Error(w, "Missing authorization token", http.StatusUnauthorized)
		return
	}

	var notification struct {
		Body string `json:"body"`
	}

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement notification sending logic
	w.WriteHeader(http.StatusOK)
}
