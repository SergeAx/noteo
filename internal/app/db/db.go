package db

import (
	"log/slog"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Config holds database configuration
type Config struct {
	DSN string
}

// NewDB creates a new database connection using the provided configuration
func NewDB(cfg *Config) (*gorm.DB, error) {
	slog.Info("Using database", "dsn", cfg.DSN)
	if cfg.DSN == ":memory:" {
		slog.Warn("In-memory database: all data will be lost when the application stops or restarts")
	}

	db, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schemas using db package models
	if err := db.AutoMigrate(
		&project{},
		&subscription{},
	); err != nil {
		slog.Error("Failed to auto-migrate database", "error", err)
		os.Exit(1)
	}

	return db, nil
}
