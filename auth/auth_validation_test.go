// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/internal/config"
	redismock "github.com/go-redis/redismock/v9"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// auth_validation_test.go: Tests for authentication-related utilities, including token validation, format checks, and Redis-backed refresh token handling.

const (
	testJWTSecret = "supersecretkeysupersecretkey123456"
)

// TestIsValidUserNameFormat tests the validation logic for allowed username formats.
func TestIsValidUserNameFormat(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"user1", true},
		{"us", false},
		{"user..name", false},
		{"user_name", true},
		{"user!name", false},
	}
	for _, c := range cases {
		if got := IsValidUserNameFormat(c.in); got != c.want {
			t.Errorf("IsValidUserNameFormat(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestIsValidEmailFormat tests the validation logic for email format correctness.
func TestIsValidEmailFormat(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"test@example.com", true},
		{"test..email@example.com", false},
		{"test@.com", false},
		{"test@domain", false},
		{"test@domain.com", true},
	}
	for _, c := range cases {
		if got := IsValidEmailFormat(c.in); got != c.want {
			t.Errorf("IsValidEmailFormat(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// makeJWT generates a signed JWT token with given claims and timing constraints.
func makeJWT(secret, issuer, audience string, notBefore, expires time.Time) (string, error) {
	claims := Claims{
		UserID: "user1",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Audience:  []string{audience},
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			NotBefore: jwt.NewNumericDate(notBefore),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// TestValidateAccessToken tests validation of access tokens under various conditions.
func TestValidateAccessToken(t *testing.T) {
	cfg := &Config{APIConfig: &config.APIConfig{Issuer: "issuer", Audience: "aud"}}
	secret := testJWTSecret
	now := time.Now().UTC()
	validToken, _ := makeJWT(secret, "issuer", "aud", now.Add(-time.Minute), now.Add(time.Hour))
	expiredToken, _ := makeJWT(secret, "issuer", "aud", now.Add(-2*time.Hour), now.Add(-time.Hour))
	notYetToken, _ := makeJWT(secret, "issuer", "aud", now.Add(time.Hour), now.Add(2*time.Hour))
	wrongIssuer, _ := makeJWT(secret, "wrong", "aud", now.Add(-time.Minute), now.Add(time.Hour))
	wrongAudience, _ := makeJWT(secret, "issuer", "wrong", now.Add(-time.Minute), now.Add(time.Hour))

	t.Run("valid", func(t *testing.T) {
		_, err := cfg.ValidateAccessToken(validToken, secret)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("expired", func(t *testing.T) {
		_, err := cfg.ValidateAccessToken(expiredToken, secret)
		if err == nil || !strings.Contains(err.Error(), "token is expired") {
			t.Errorf("expected token is expired error, got %v", err)
		}
	})
	t.Run("not yet valid", func(t *testing.T) {
		_, err := cfg.ValidateAccessToken(notYetToken, secret)
		if err == nil || !strings.Contains(err.Error(), "not valid yet") {
			t.Errorf("expected not valid yet error, got %v", err)
		}
	})
	t.Run("wrong issuer", func(t *testing.T) {
		_, err := cfg.ValidateAccessToken(wrongIssuer, secret)
		if err == nil || !strings.Contains(err.Error(), "invalid issuer") {
			t.Errorf("expected invalid issuer error, got %v", err)
		}
	})
	t.Run("wrong audience", func(t *testing.T) {
		_, err := cfg.ValidateAccessToken(wrongAudience, secret)
		if err == nil || !strings.Contains(err.Error(), "invalid audience") {
			t.Errorf("expected invalid audience error, got %v", err)
		}
	})
}

// TestValidateRefreshToken tests refresh token validation including format and signature.
func TestValidateRefreshToken(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cfg := &Config{APIConfig: &config.APIConfig{RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", RedisClient: db}}
	userID := uuid.New().String()
	refreshToken, _ := cfg.GenerateRefreshToken(userID)

	t.Run("valid", func(t *testing.T) {
		_, err := cfg.ValidateRefreshToken(refreshToken)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	t.Run("invalid format", func(t *testing.T) {
		mock.ExpectKeys("refresh_token:*").SetVal([]string{})
		_, err := cfg.ValidateRefreshToken("notavalidtoken")
		if err == nil {
			t.Error("expected error for invalid format")
		}
	})
	t.Run("signature mismatch", func(t *testing.T) {
		parts := strings.Split(refreshToken, ":")
		badToken := parts[0] + ":" + parts[1] + ":badsignature"
		_, err := cfg.ValidateRefreshToken(badToken)
		if err == nil || !strings.Contains(err.Error(), "invalid refresh token signature") {
			t.Errorf("expected signature error, got %v", err)
		}
	})
}

// TestValidateCookieRefreshTokenData is a placeholder for cookie+Redis-based refresh token validation tests.
func TestValidateCookieRefreshTokenData(_ *testing.T) {
	// Placeholder for future Redis-mocking tests using redismock
}

// TestValidateAccessToken_Errors tests edge cases for token parsing and invalid secrets.
func TestValidateAccessToken_Errors(t *testing.T) {
	cfg := &Config{APIConfig: &config.APIConfig{Issuer: "issuer", Audience: "aud"}}
	secret := testJWTSecret
	// Invalid JWT format
	_, err := cfg.ValidateAccessToken("not.a.jwt", secret)
	if err == nil {
		t.Error("expected error for invalid JWT format")
	}
	// Invalid token (token.Valid == false): use a valid JWT but with wrong secret
	validToken, _ := makeJWT(secret, "issuer", "aud", time.Now().Add(-time.Minute), time.Now().Add(time.Hour))
	_, err = cfg.ValidateAccessToken(validToken, "wrongsecretwrongsecretwrongsecret12")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

// TestValidateRefreshToken_InvalidUserID ensures invalid userID formats are rejected.
func TestValidateRefreshToken_InvalidUserID(t *testing.T) {
	cfg := &Config{APIConfig: &config.APIConfig{RefreshSecret: "refreshsecretkeyrefreshsecretkey1234"}}
	badToken := "invalid:test:token" // nolint:gosec // This is a test token, not real credentials
	_, err := cfg.ValidateRefreshToken(badToken)
	if err == nil || !strings.Contains(err.Error(), "invalid userID in refresh token") {
		t.Error("expected error for invalid userID in refresh token")
	}
}

// TestValidateCookieRefreshTokenData_ErrorsAndHappyPath tests cookie-based refresh token flow with Redis.
func TestValidateCookieRefreshTokenData_ErrorsAndHappyPath(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cfg := &Config{APIConfig: &config.APIConfig{RefreshSecret: "refreshsecretkeyrefreshsecretkey1234", RedisClient: db}}
	w := &dummyResponseWriter{}

	// Missing cookie
	r, _ := http.NewRequest("GET", "/", nil)
	_, _, err := cfg.ValidateCookieRefreshTokenData(w, r)
	if err == nil {
		t.Error("expected error for missing cookie")
	}

	// Error from ValidateRefreshToken
	r2, _ := http.NewRequest("GET", "/", nil)
	r2.AddCookie(&http.Cookie{Name: "refresh_token", Value: "badtoken"})
	mock.ExpectKeys("refresh_token:*").SetVal([]string{})
	_, _, err = cfg.ValidateCookieRefreshTokenData(w, r2)
	if err == nil {
		t.Error("expected error from ValidateRefreshToken")
	}

	// Redis Get error
	userID := uuid.New().String()
	refreshToken, _ := cfg.GenerateRefreshToken(userID)
	r3, _ := http.NewRequest("GET", "/", nil)
	r3.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	mock.ExpectGet("refresh_token:" + userID).SetErr(fmt.Errorf("redis get error"))
	_, _, err = cfg.ValidateCookieRefreshTokenData(w, r3)
	if err == nil || !strings.Contains(err.Error(), "redis get error") {
		t.Error("expected redis get error")
	}

	// ParseRefreshTokenData error
	r4, _ := http.NewRequest("GET", "/", nil)
	r4.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	mock.ExpectGet("refresh_token:" + userID).SetVal("notjson")
	_, _, err = cfg.ValidateCookieRefreshTokenData(w, r4)
	if err == nil {
		t.Error("expected parse error")
	}

	// Invalid session (storedData.Token != refreshToken)
	r5, _ := http.NewRequest("GET", "/", nil)
	r5.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	stored, _ := json.Marshal(RefreshTokenData{Token: "othertoken", Provider: "local"})
	mock.ExpectGet("refresh_token:" + userID).SetVal(string(stored))
	_, _, err = cfg.ValidateCookieRefreshTokenData(w, r5)
	if err == nil || !strings.Contains(err.Error(), "invalid session") {
		t.Error("expected invalid session error")
	}

	// Happy path
	r6, _ := http.NewRequest("GET", "/", nil)
	r6.AddCookie(&http.Cookie{Name: "refresh_token", Value: refreshToken})
	stored, _ = json.Marshal(RefreshTokenData{Token: refreshToken, Provider: "local"})
	mock.ExpectGet("refresh_token:" + userID).SetVal(string(stored))
	uid, data, err := cfg.ValidateCookieRefreshTokenData(w, r6)
	if err != nil || uid.String() != userID || data.Token != refreshToken {
		t.Errorf("expected happy path, got uid=%v, data=%v, err=%v", uid, data, err)
	}
}

// dummyResponseWriter is a stub for http.ResponseWriter used in tests.
type dummyResponseWriter struct{}

func (d *dummyResponseWriter) Header() http.Header       { return http.Header{} }
func (d *dummyResponseWriter) Write([]byte) (int, error) { return 0, nil }
func (d *dummyResponseWriter) WriteHeader(_ int)         {}

// TestGetUserIDFromRefreshToken_ErrorsAndHappyPath tests mapping refresh tokens back to user IDs via Redis.
func TestGetUserIDFromRefreshToken_ErrorsAndHappyPath(t *testing.T) {
	db, mock := redismock.NewClientMock()
	cfg := &Config{APIConfig: &config.APIConfig{RedisClient: db}}

	// Redis Keys error
	mock.ExpectKeys("refresh_token:*").SetErr(fmt.Errorf("keys error"))
	_, err := cfg.GetUserIDFromRefreshToken("token")
	if err == nil || !strings.Contains(err.Error(), "keys error") {
		t.Error("expected keys error")
	}

	// Redis Get error
	mock.ExpectKeys("refresh_token:*").SetVal([]string{"refresh_token:uid1"})
	mock.ExpectGet("refresh_token:uid1").SetErr(fmt.Errorf("get error"))
	_, err = cfg.GetUserIDFromRefreshToken("token")
	if err == nil || !strings.Contains(err.Error(), "refresh token not found") {
		t.Error("expected not found error due to get error")
	}

	// ParseRefreshTokenData error
	mock.ExpectKeys("refresh_token:*").SetVal([]string{"refresh_token:uid1"})
	mock.ExpectGet("refresh_token:uid1").SetVal("notjson")
	_, err = cfg.GetUserIDFromRefreshToken("token")
	if err == nil || !strings.Contains(err.Error(), "refresh token not found") {
		t.Error("expected not found error due to parse error")
	}

	// Invalid user ID format in Redis key
	stored, _ := json.Marshal(RefreshTokenData{Token: "token", Provider: "google"})
	mock.ExpectKeys("refresh_token:*").SetVal([]string{"refresh_token:baduid"})
	mock.ExpectGet("refresh_token:baduid").SetVal(string(stored))
	_, err = cfg.GetUserIDFromRefreshToken("token")
	if err == nil || !strings.Contains(err.Error(), "invalid user ID format") {
		t.Error("expected invalid user ID format error")
	}

	// Not found (no match)
	mock.ExpectKeys("refresh_token:*").SetVal([]string{"refresh_token:uid1"})
	mock.ExpectGet("refresh_token:uid1").SetVal(string(stored))
	_, err = cfg.GetUserIDFromRefreshToken("notfound")
	if err == nil || !strings.Contains(err.Error(), "refresh token not found") {
		t.Error("expected not found error")
	}

	// Happy path (provider == "google")
	uid := uuid.New().String()
	stored, _ = json.Marshal(RefreshTokenData{Token: "token", Provider: "google"})
	mock.ExpectKeys("refresh_token:*").SetVal([]string{"refresh_token:" + uid})
	mock.ExpectGet("refresh_token:" + uid).SetVal(string(stored))
	got, err := cfg.GetUserIDFromRefreshToken("token")
	if err != nil || got.String() != uid {
		t.Errorf("expected happy path, got %v, err %v", got, err)
	}
}
