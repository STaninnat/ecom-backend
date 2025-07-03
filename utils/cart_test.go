package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
