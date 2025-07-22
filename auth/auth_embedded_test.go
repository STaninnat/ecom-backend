// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"encoding/json"
	"testing"
)

// auth_embedded_test.go: Tests for JSON marshaling of Claims and RefreshTokenData structs.

// TestClaimsJSONTags checks that Claims struct can be marshaled into valid JSON.
func TestClaimsJSONTags(t *testing.T) {
	claims := Claims{UserID: "user1"}
	b, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(b) == "" || !json.Valid(b) {
		t.Error("invalid JSON for Claims")
	}
}

// TestRefreshTokenDataJSONTags checks that RefreshTokenData struct can be marshaled into valid JSON.
func TestRefreshTokenDataJSONTags(t *testing.T) {
	data := RefreshTokenData{Token: "tok", Provider: "local"}
	b, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(b) == "" || !json.Valid(b) {
		t.Error("invalid JSON for RefreshTokenData")
	}
}
