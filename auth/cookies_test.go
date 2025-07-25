// Package auth provides authentication, token management, validation, and session utilities for the ecom-backend project.
package auth

import (
	"net/http/httptest"
	"testing"
	"time"
)

// cookies_test.go: Tests for setting access and refresh tokens as secure HTTP cookies.

// TestSetTokensAsCookies verifies that SetTokensAsCookies sets both tokens with correct values and attributes.
func TestSetTokensAsCookies(t *testing.T) {
	w := httptest.NewRecorder()
	accessToken := "access123"
	refreshToken := "refresh456"
	accessExp := time.Now().Add(time.Hour)
	refreshExp := time.Now().Add(2 * time.Hour)

	SetTokensAsCookies(w, accessToken, refreshToken, accessExp, refreshExp)

	// Check cookies
	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}

	var foundAccess, foundRefresh bool
	for _, c := range cookies {
		switch c.Name {
		case "access_token":
			foundAccess = true
			if c.Value != accessToken {
				t.Errorf("access_token value = %q, want %q", c.Value, accessToken)
			}
			if !c.HttpOnly || !c.Secure || c.Path != "/" {
				t.Error("access_token cookie attributes incorrect")
			}
			if !c.Expires.IsZero() {
				delta := c.Expires.Sub(accessExp)
				if delta < -2*time.Second || delta > 2*time.Second {
					t.Errorf("access_token expires = %v, want %v (delta %v)", c.Expires, accessExp, delta)
				}
			}
		case "refresh_token":
			foundRefresh = true
			if c.Value != refreshToken {
				t.Errorf("refresh_token value = %q, want %q", c.Value, refreshToken)
			}
			if !c.HttpOnly || !c.Secure || c.Path != "/" {
				t.Error("refresh_token cookie attributes incorrect")
			}
			if !c.Expires.IsZero() {
				delta := c.Expires.Sub(refreshExp)
				if delta < -2*time.Second || delta > 2*time.Second {
					t.Errorf("refresh_token expires = %v, want %v (delta %v)", c.Expires, refreshExp, delta)
				}
			}
		}
	}
	if !foundAccess || !foundRefresh {
		t.Error("missing access_token or refresh_token cookie")
	}
}

// TestSetTokensAsCookies_EmptyAndZero verifies behavior when tokens and expirations are empty or zero.
func TestSetTokensAsCookies_EmptyAndZero(t *testing.T) {
	w := httptest.NewRecorder()
	SetTokensAsCookies(w, "", "", time.Time{}, time.Time{})
	cookies := w.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("expected 2 cookies, got %d", len(cookies))
	}
	for _, c := range cookies {
		if c.Value != "" {
			t.Errorf("expected empty value, got %q", c.Value)
		}
		if !c.Expires.IsZero() {
			t.Errorf("expected zero expiration, got %v", c.Expires)
		}
	}
}
