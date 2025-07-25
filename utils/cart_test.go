// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// cart_test.go: Tests for cart session helpers and session ID retrieval.

// TestGetSessionIDFromRequest tests the GetSessionIDFromRequest function for:
// - When the session cookie is present
// - When the session cookie is absent
func TestGetSessionIDFromRequest(t *testing.T) {
	t.Run("cookie present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		cookieValue := "test-session-id"
		req.AddCookie(&http.Cookie{Name: GuestCartSessionCookie, Value: cookieValue})
		got := GetSessionIDFromRequest(req)
		if got != cookieValue {
			t.Errorf("expected %q, got %q", cookieValue, got)
		}
	})

	t.Run("cookie absent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		got := GetSessionIDFromRequest(req)
		if got != "" {
			t.Errorf("expected empty string, got %q", got)
		}
	})
}
