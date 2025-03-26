package config

import (
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthConfig struct {
	Google *oauth2.Config
}

func NewOAuthConfig() (*OAuthConfig, error) {
	credentialsPath := os.Getenv("GOOGLE_CREDENTIALS_PATH")
	if credentialsPath == "" {
		log.Fatal("Warning: Google credentials path environment variable is not set")
	}

	data, err := os.ReadFile(credentialsPath)
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
