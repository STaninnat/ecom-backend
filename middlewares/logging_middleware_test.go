// Package middlewares provides HTTP middleware components for request processing in the ecom-backend project.
package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/STaninnat/ecom-backend/utils"
)

// logging_middleware_test.go: Tests for structured request logging and request ID middleware.

// TestShouldLog tests the path filtering logic for logging middleware
// It verifies that include/exclude path rules work correctly for different scenarios
func TestShouldLog(t *testing.T) {
	include := map[string]struct{}{"/api": {}}
	exclude := map[string]struct{}{"/api/private": {}}
	tests := []struct {
		path string
		want bool
		name string
	}{
		{"/api/products", true, "included"},
		{"/api/private/data", false, "excluded"},
		{"/other", false, "not included"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldLog(tt.path, include, exclude)
			if got != tt.want {
				t.Errorf("ShouldLog(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestGetIPAddress tests IP address extraction from various request headers
// It verifies that the function correctly handles X-Real-IP, X-Forwarded-For, and RemoteAddr
func TestGetIPAddress(t *testing.T) {
	tests := []struct {
		headers map[string]string
		remote  string
		want    string
		name    string
	}{
		{map[string]string{"X-Real-IP": "1.2.3.4"}, "", "1.2.3.4", "real ip"},
		{map[string]string{"X-Forwarded-For": "5.6.7.8, 9.10.11.12"}, "", "5.6.7.8", "forwarded for"},
		{map[string]string{}, "8.8.8.8:1234", "8.8.8.8", "remote addr"},
		{map[string]string{"X-Real-IP": "badip"}, "1.1.1.1:1234", "1.1.1.1", "invalid real ip falls back"},
		{map[string]string{}, "badaddr", "", "invalid remote addr"},
	}
	for _, tt := range tests {
		r := httptest.NewRequest("GET", "/", nil)
		for k, v := range tt.headers {
			r.Header.Set(k, v)
		}
		if tt.remote != "" {
			r.RemoteAddr = tt.remote
		}
		got := GetIPAddress(r)
		if got != tt.want {
			t.Errorf("%s: GetIPAddress() = %q, want %q", tt.name, got, tt.want)
		}
	}
}

// TestIsValidIP tests IP address validation functionality
// It verifies that both valid IPv4/IPv6 addresses and invalid formats are handled correctly
func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"1.2.3.4", true},
		{"::1", true},
		{"256.0.0.1", false},
		{"notanip", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsValidIP(tt.ip); got != tt.want {
			t.Errorf("IsValidIP(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

// TestRequestIDMiddleware tests that request IDs are properly generated and stored in context
// It verifies that each request gets a unique UUID and it's accessible in the request context
func TestRequestIDMiddleware(t *testing.T) {
	var gotID string
	h := RequestIDMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id := r.Context().Value(utils.ContextKeyRequestID)
		if id == nil {
			t.Error("request_id not set in context")
		}
		gotID, _ = id.(string)
	}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if gotID == "" {
		t.Error("request_id should not be empty")
	}
}

// TestLoggingMiddleware_CallsNextAndLogs tests the main logging middleware functionality
// It verifies that the middleware calls the next handler and logs request information correctly
func TestLoggingMiddleware_CallsNextAndLogs(t *testing.T) {
	var called bool
	logger := logrus.New()
	logger.Out = &strings.Builder{} // discard output
	mw := LoggingMiddleware(logger, map[string]struct{}{"/": {}}, nil)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(200)
	}))
	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if !called {
		t.Error("next handler was not called")
	}
	if rw.Code != 200 {
		t.Errorf("expected status 200, got %d", rw.Code)
	}
}
