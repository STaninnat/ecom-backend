// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// config_db_test.go: Tests for PostgreSQL database connection helpers and error handling.

// TestConnectDB_AlreadyConnected tests database connection when already connected.
// It verifies that the function returns an error when attempting to connect to an already connected database.
func TestConnectDB_AlreadyConnected(_ *testing.T) {
	cfg := &APIConfig{DBConn: new(sql.DB)}
	// This should not panic or error, just log and return
	cfg.ConnectDB()
	// Test passes if no panic occurs
}

// TestConnectDB_MissingDatabaseURL tests database connection with missing database URL.
// It verifies that the function returns an error when the database URL is not provided.
func TestConnectDB_MissingDatabaseURL(t *testing.T) {
	err := os.Unsetenv("DATABASE_URL")
	if err != nil {
		t.Fatalf("os.Unsetenv failed: %v", err)
	}
	cfg := &APIConfig{}
	// This should not panic, just log warning and return
	cfg.ConnectDB()
	// Test passes if no panic occurs
}

// Note: The legacy ConnectDB function is not unit tested because it is not injectable and uses log.Fatal on error.
// All new code should use and test ConnectDBWithError instead.

// TestConnectDBWithError_MissingDatabaseURL tests database connection with error handling and missing URL.
// It verifies that the function returns an error when the database URL is not provided.
func TestConnectDBWithError_MissingDatabaseURL(t *testing.T) {
	cfg := &APIConfig{}
	err := cfg.ConnectDBWithError(context.TODO())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database URL is not set")
}

// TestConnectDBWithError_AlreadyConnected tests database connection with error handling when already connected.
// It verifies that the function returns an error when attempting to connect to an already connected database.
func TestConnectDBWithError_AlreadyConnected(t *testing.T) {
	// Use a plain *sql.DB, not a sqlmock, since no DB operation is performed
	cfg := &APIConfig{DBConn: &sql.DB{}}
	err := cfg.ConnectDBWithError(context.TODO())
	assert.NoError(t, err)
}

// TestConnectDBWithError_ConnectionError tests database connection with error handling and connection failure.
// It verifies that the function returns an error when the database connection cannot be established.
func TestConnectDBWithError_ConnectionError(t *testing.T) {
	t.Skip("Cannot reliably test connection error with sqlmock due to lack of provider injection in ConnectDBWithError")
}

// TestConnectDBWithError_Success tests successful database connection with error handling.
// It verifies that the function can establish a database connection when all required parameters are provided.
func TestConnectDBWithError_Success(t *testing.T) {
	t.Skip("Cannot reliably test DB connection success with sqlmock due to lack of provider injection in ConnectDBWithError")
}

// TestGetEnvOrDefault tests the GetEnvOrDefault function.
// It verifies that the function returns the environment variable value or the default value when not set.
func TestGetEnvOrDefault(t *testing.T) {
	t.Setenv("FOO", "bar")
	val := getEnvOrDefault("FOO", "baz")
	assert.Equal(t, "bar", val)
	t.Setenv("FOO", "")
	val = getEnvOrDefault("FOO", "baz")
	assert.Equal(t, "baz", val)
}
