package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthConfig struct {
	Google *oauth2.Config
}

func NewOAuthConfig(credsPath string) (*OAuthConfig, error) {
	safePath := filepath.Clean(credsPath)

	if !isSafePath(safePath) {
		return nil, fmt.Errorf("unsafe file path")
	}

	data, err := os.ReadFile(safePath)
	if err != nil {
		return nil, err
	}

	googleConfig, err := google.ConfigFromJSON(data,
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
	)
	if err != nil {
		return nil, err
	}

	return &OAuthConfig{
		Google: googleConfig,
	}, nil
}

func isSafePath(path string) bool {
	return !strings.Contains(path, "..")
}
