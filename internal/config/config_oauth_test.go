package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
