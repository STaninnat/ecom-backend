// Package config provides configuration management, validation, and provider logic for the ecom-backend project.
package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// config_oauth_test.go: Tests for Google OAuth2 configuration helpers and path validation.

// TestIsSafePath tests the IsSafePath function.
// It verifies that the function correctly identifies safe and unsafe file paths.
func TestIsSafePath(t *testing.T) {
	assert.True(t, isSafePath("/safe/path/creds.json"))
	assert.False(t, isSafePath("../unsafe/creds.json"))
	assert.False(t, isSafePath("creds/../../secrets.json"))
}

// TestNewOAuthConfig_UnsafePath tests OAuth config creation with unsafe path.
// It verifies that the function returns an error when an unsafe credentials path is provided.
func TestNewOAuthConfig_UnsafePath(t *testing.T) {
	cfg, err := NewOAuthConfig("../unsafe/creds.json")
	require.Error(t, err)
	assert.Nil(t, cfg)
}
