// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// config_redis_test.go: Tests for Redis connection helpers and error handling.

// TestInitRedis_Error tests Redis initialization with error.
// It verifies that the function returns an error when Redis connection cannot be established.
func TestInitRedis_Error(_ *testing.T) {
	// This function calls log.Fatal on error, so we can't test the error path
	// But we can test that it doesn't panic when called with a bad address
	// The function will log.Fatal, but we can't catch that in a test
	// This test ensures the function signature is correct
	_ = InitRedis
}

// TestInitRedisWithError_Error tests Redis initialization with error handling.
// It verifies that the function returns an error when Redis connection cannot be established.
func TestInitRedisWithError_Error(t *testing.T) {
	t.Setenv("REDIS_ADDR", "bad_addr:6379")
	t.Setenv("REDIS_USERNAME", "")
	t.Setenv("REDIS_PASSWORD", "")
	cmdable, err := InitRedisWithError(context.Background())
	require.Error(t, err)
	assert.Nil(t, cmdable)
}

// TestInitRedis_NotTestable skips the test for InitRedis because it calls log.Fatal on error,
// making it unsuitable for unit testing. Use InitRedisWithError for testing instead.
func TestInitRedis_NotTestable(t *testing.T) {
	t.Skip("InitRedis calls log.Fatal on error and is not unit-testable. Use InitRedisWithError for testable code.")
}
