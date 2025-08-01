// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	redismock "github.com/go-redis/redismock/v9"

	"github.com/STaninnat/ecom-backend/internal/config"
	"github.com/STaninnat/ecom-backend/utils"
)

// token_manager_test.go: Tests for token generation, storage in Redis, and config validation.

const (
	shortSecret = "short"
)

// Use the mockRedisClient from auth_validation_test.go for Redis mocking in tests.

// TestGenerateAccessToken verifies access token generation with valid, expired, short secret, and nil config cases.
func TestGenerateAccessToken(t *testing.T) {
	cfg := &Config{APIConfig: &config.APIConfig{JWTSecret: "supersecretkeysupersecretkey123456", Issuer: "issuer", Audience: "aud"}}
	expires := time.Now().Add(time.Hour)
	t.Run("valid", func(t *testing.T) {
		tok, err := cfg.GenerateAccessToken("user1", expires)
		if err != nil || tok == "" {
			t.Errorf("expected token, got err: %v", err)
		}
	})
	t.Run("expired", func(t *testing.T) {
		_, err := cfg.GenerateAccessToken("user1", time.Now().Add(-time.Hour))
		if err == nil {
			t.Error("expected error for expired token")
		}
	})
	cfg.JWTSecret = shortSecret
	_, err := cfg.GenerateAccessToken("user1", expires)
	if err == nil {
		t.Error("expected error for short secret")
	}
	// New: nil cfg
	t.Run("nil cfg", func(t *testing.T) {
		_, err := (*Config)(nil).GenerateAccessToken("user1", expires)
		if err == nil || err.Error() != "cfg is nil" {
			t.Error("expected cfg is nil error")
		}
	})
}

// TestGenerateRefreshToken verifies refresh token generation with valid, short secret, and nil config cases.
func TestGenerateRefreshToken(t *testing.T) {
	cfg := &Config{APIConfig: &config.APIConfig{RefreshSecret: "refreshsecretkeyrefreshsecretkey1234"}}
	t.Run("valid", func(t *testing.T) {
		tok, err := cfg.GenerateRefreshToken(utils.NewUUIDString())
		if err != nil || tok == "" {
			t.Errorf("expected token, got err: %v", err)
		}
	})
	cfg.RefreshSecret = shortSecret
	_, err := cfg.GenerateRefreshToken(utils.NewUUIDString())
	if err == nil {
		t.Error("expected error for short secret")
	}
	// New: nil cfg
	t.Run("nil cfg", func(t *testing.T) {
		_, err := (*Config)(nil).GenerateRefreshToken("user1")
		if err == nil || err.Error() != "cfg is nil" {
			t.Error("expected cfg is nil error")
		}
	})
}

// TestGenerateTokens verifies generation of both access and refresh tokens and error cases with short secrets and nil config.
func TestGenerateTokens(t *testing.T) {
	cfg := &Config{APIConfig: &config.APIConfig{JWTSecret: "supersecretkeysupersecretkey123456", RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", Issuer: "issuer", Audience: "aud"}}
	access, refresh, err := cfg.GenerateTokens("user1", time.Now().Add(time.Hour))
	if err != nil || access == "" || refresh == "" {
		t.Errorf("expected tokens, got err: %v", err)
	}
	// Error from GenerateAccessToken (short secret)
	cfg.JWTSecret = shortSecret
	_, _, err = cfg.GenerateTokens("user1", time.Now().Add(time.Hour))
	if err == nil {
		t.Error("expected error from GenerateAccessToken")
	}
	// Error from GenerateRefreshToken (short secret)
	cfg.JWTSecret = "supersecretkeysupersecretkey123456"
	cfg.RefreshSecret = shortSecret
	_, _, err = cfg.GenerateTokens("user1", time.Now().Add(time.Hour))
	if err == nil {
		t.Error("expected error from GenerateRefreshToken")
	}
	// Nil config
	_, _, err = (*Config)(nil).GenerateTokens("user1", time.Now().Add(time.Hour))
	if err == nil {
		t.Error("expected error from nil config")
	}
}

// TestStoreRefreshTokenInRedis tests storing refresh tokens in Redis with various scenarios including errors.
func TestStoreRefreshTokenInRedis(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cfg := &Config{APIConfig: &config.APIConfig{RedisClient: db}}
	r, _ := http.NewRequest("GET", "/", nil)

	// Prepare expected JSON as []byte
	tokenData := RefreshTokenData{Token: "token", Provider: "local"}
	jsonData, _ := json.Marshal(tokenData)

	// Set up expected Redis calls for the valid case
	mock.ExpectSet("refresh_token:user1", jsonData, time.Minute).SetVal("OK")
	mock.ExpectSet("refresh_token_lookup:token", "user1", time.Minute).SetVal("OK")

	err := cfg.StoreRefreshTokenInRedis(r, "user1", "token", "local", time.Minute)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = cfg.StoreRefreshTokenInRedis(r, "user1", "", "local", time.Minute)
	if err == nil {
		t.Error("expected error for empty token")
	}
	err = cfg.StoreRefreshTokenInRedis(r, "user1", "token", "unsupported", time.Minute)
	if err == nil {
		t.Error("expected error for unsupported provider")
	}
	cfg.RedisClient = nil
	err = cfg.StoreRefreshTokenInRedis(r, "user1", "token", "local", time.Minute)
	if err == nil {
		t.Error("expected error for nil RedisClient")
	}
	// New: negative TTL
	db, _ = redismock.NewClientMock()
	cfg = &Config{APIConfig: &config.APIConfig{RedisClient: db}}
	err = cfg.StoreRefreshTokenInRedis(r, "user1", "token", "local", -1)
	if err == nil || err.Error() != "invalid TTL" {
		t.Error("expected invalid TTL error")
	}
	// New: Redis Set error
	db, mock = redismock.NewClientMock()
	cfg = &Config{APIConfig: &config.APIConfig{RedisClient: db}}
	tokenData = RefreshTokenData{Token: "token", Provider: "local"}
	jsonData, _ = json.Marshal(tokenData)
	mock.ExpectSet("refresh_token:user1", jsonData, time.Minute).SetErr(fmt.Errorf("redis set error"))
	err = cfg.StoreRefreshTokenInRedis(r, "user1", "token", "local", time.Minute)
	if err == nil || err.Error() != "redis set error" {
		t.Error("expected redis set error")
	}
}

// TestStoreRefreshTokenInRedis_GoogleProvider tests storing refresh tokens with Google provider in Redis.
func TestStoreRefreshTokenInRedis_GoogleProvider(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cfg := &Config{APIConfig: &config.APIConfig{RedisClient: db}}
	r, _ := http.NewRequest("GET", "/", nil)

	tokenData := RefreshTokenData{Token: "token", Provider: "google"}
	jsonData, _ := json.Marshal(tokenData)
	mock.ExpectSet("refresh_token:user1", jsonData, time.Minute).SetVal("OK")
	mock.ExpectSet("refresh_token_lookup:token", "user1", time.Minute).SetVal("OK")

	err := cfg.StoreRefreshTokenInRedis(r, "user1", "token", "google", time.Minute)
	if err != nil {
		t.Errorf("expected no error for google provider, got %v", err)
	}
}

// TestParseRefreshTokenData tests parsing of refresh token data JSON with valid and error cases.
func TestParseRefreshTokenData(t *testing.T) {
	data := RefreshTokenData{Token: "tok", Provider: "local"}
	b, _ := json.Marshal(data)
	parsed, err := ParseRefreshTokenData(string(b))
	if err != nil || parsed.Token != "tok" {
		t.Errorf("expected parsed data, got err: %v", err)
	}
	_, err = ParseRefreshTokenData("notjson")
	if err == nil {
		t.Error("expected error for invalid json")
	}
	b, _ = json.Marshal(struct{ Token string }{"tok"})
	_, err = ParseRefreshTokenData(string(b))
	if err == nil {
		t.Error("expected error for missing provider")
	}
	// New: empty token
	b, _ = json.Marshal(struct{ Provider string }{"local"})
	_, err = ParseRefreshTokenData(string(b))
	if err == nil {
		t.Error("expected error for missing token")
	}
	// New: empty provider
	b, _ = json.Marshal(struct{ Token string }{"tok"})
	_, err = ParseRefreshTokenData(string(b))
	if err == nil {
		t.Error("expected error for missing provider")
	}
}

// TestValidateConfig tests secret validation logic with empty, short, and valid secrets.
func TestValidateConfig(t *testing.T) {
	err := ValidateConfig("", "test")
	if err == nil {
		t.Error("expected error for empty secret")
	}
	err = ValidateConfig(shortSecret, "test")
	if err == nil {
		t.Error("expected error for short secret")
	}
	err = ValidateConfig("longenoughsecretlongenoughsecret", "test")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestGetUserIDByRefreshToken tests retrieval of userID by refresh token from Redis with success and error scenarios.
func TestGetUserIDByRefreshToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cfg := &Config{APIConfig: &config.APIConfig{RedisClient: db}}

	// Set up expected Redis call for the valid case
	mock.ExpectGet("refresh_token_lookup:validtoken").SetVal("user1")
	mock.ExpectGet("refresh_token_lookup:invalidtoken").RedisNil()

	userID, err := cfg.GetUserIDByRefreshToken(context.Background(), "validtoken")
	if err != nil || userID != "user1" {
		t.Errorf("expected user1, got %v, err %v", userID, err)
	}
	_, err = cfg.GetUserIDByRefreshToken(context.Background(), "invalidtoken")
	if err == nil {
		t.Error("expected error for invalid token")
	}
	// New: Redis Get error
	mock.ExpectGet("refresh_token_lookup:error").SetErr(fmt.Errorf("redis get error"))
	_, err = cfg.GetUserIDByRefreshToken(context.Background(), "error")
	if err == nil || err.Error() != "redis get error" {
		t.Error("expected redis get error")
	}
}
