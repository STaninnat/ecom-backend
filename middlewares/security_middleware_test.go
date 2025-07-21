// Package middlewares provides HTTP middleware components for request processing in the ecom-backend project.
package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// security_middleware_test.go: Tests for security headers and request validation middleware.

// TestSecurityHeaders tests that security headers are properly set on responses
// It verifies that all required security headers are present with correct values
func TestSecurityHeaders(t *testing.T) {
	h := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	headers := map[string]string{
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY",
		"X-XSS-Protection":          "1; mode=block",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Content-Security-Policy":   "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'",
		"Permissions-Policy":        "geolocation=(), microphone=(), camera=()",
	}
	for k, v := range headers {
		if got := rw.Header().Get(k); got != v {
			t.Errorf("header %q: expected %q, got %q", k, v, got)
		}
	}
}

// TestNoCacheHeaders tests that no-cache headers are properly set on responses
// It verifies that Cache-Control, Pragma, and Expires headers prevent caching
func TestNoCacheHeaders(t *testing.T) {
	h := NoCacheHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
	}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)

	headers := map[string]string{
		"Cache-Control": "no-cache, no-store, must-revalidate",
		"Pragma":        "no-cache",
		"Expires":       "0",
	}
	for k, v := range headers {
		if got := rw.Header().Get(k); got != v {
			t.Errorf("header %q: expected %q, got %q", k, v, got)
		}
	}
}
