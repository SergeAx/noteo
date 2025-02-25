package db

import (
	"log/slog"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("noteo.sqlite"), &gorm.Config{})
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
