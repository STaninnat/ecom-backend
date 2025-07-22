// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Import for PostgreSQL driver registration
)

// config_db.go: PostgreSQL database connection helpers and legacy patterns.

// ConnectDB establishes a connection to the PostgreSQL database and initializes the queries object.
// Uses the legacy pattern and calls log.Fatal on errors.
// Prefer ConnectDBWithError for better error handling.
func (cfg *APIConfig) ConnectDB() {
	if cfg.DBConn != nil {
		log.Println("Database already connected")
		return
	}

	dbURL := getEnvOrDefault("DATABASE_URL", "")
	if dbURL == "" {
		log.Println("Warning: database URL is not set")
		return
	}

	provider := NewPostgresProvider(dbURL)
	db, dbQueries, err := provider.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	cfg.DB = dbQueries
	cfg.DBConn = db
	log.Println("Connected to database successfully...")
}

// ConnectDBWithError establishes a connection to the PostgreSQL database and initializes the queries object.
// Returns errors instead of calling log.Fatal, making it suitable for testing and graceful error handling.
// Verifies connectivity with a ping operation and sets up the database queries for use.
func (cfg *APIConfig) ConnectDBWithError(ctx context.Context) error {
	if cfg.DBConn != nil {
		return nil
	}

	dbURL := getEnvOrDefault("DATABASE_URL", "")
	if dbURL == "" {
		return fmt.Errorf("database URL is not set")
	}

	provider := NewPostgresProvider(dbURL)
	db, dbQueries, err := provider.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	cfg.DB = dbQueries
	cfg.DBConn = db
	return nil
}

// getEnvOrDefault retrieves an environment variable value or returns a default if not set.
// This helper function provides a convenient way to access environment variables
// with fallback default values for optional configuration settings.
func getEnvOrDefault(key, defaultValue string) string {
	provider := NewEnvironmentProvider()
	return provider.GetStringOrDefault(key, defaultValue)
}
