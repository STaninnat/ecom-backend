package config

import (
	"strings"

	"golang.org/x/oauth2"
)

// OAuthConfig holds OAuth2 configuration for Google authentication.
type OAuthConfig struct {
	Google *oauth2.Config
}

// NewOAuthConfig creates OAuth configuration using the existing pattern for backward compatibility.
// Loads Google OAuth configuration from a credentials file using the OAuthProvider interface.
func NewOAuthConfig(credsPath string) (*OAuthConfig, error) {
	provider := NewOAuthProvider()
	return provider.LoadGoogleConfig(credsPath)
}

// isSafePath validates that a file path is safe for file system operations.
// This function checks for path traversal attempts and ensures the path doesn't contain
// potentially dangerous patterns that could lead to security vulnerabilities.
func isSafePath(path string) bool {
	return !strings.Contains(path, "..")
}
