package middlewares_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

type bufWriter struct{ bytes.Buffer }

func (b *bufWriter) Write(p []byte) (int, error) { return b.Buffer.Write(p) }

func TestShouldLog(t *testing.T) {
	testcases := []struct {
		name    string
		path    string
		include map[string]struct{}
		exclude map[string]struct{}
		expect  bool
	}{
		{
			name:    "include_match_should_log",
			path:    "/v1/user",
			include: map[string]struct{}{"/v1": {}},
			exclude: map[string]struct{}{"/v1/healthz": {}, "/v1/error": {}},
			expect:  true,
		},
		{
			name:    "exclude_healthz_should_not_log",
			path:    "/v1/healthz",
			include: map[string]struct{}{"/v1": {}},
			exclude: map[string]struct{}{"/v1/healthz": {}, "/v1/error": {}},
			expect:  false,
		},
		{
			name:    "exclude_error_should_not_log",
			path:    "/v1/error",
			include: map[string]struct{}{"/v1": {}},
			exclude: map[string]struct{}{"/v1/healthz": {}, "/v1/error": {}},
			expect:  false,
		},
		{
			name:    "not_included_path_should_not_log",
			path:    "/v2/user",
			include: map[string]struct{}{"/v1": {}},
			exclude: map[string]struct{}{"/v1/healthz": {}, "/v1/error": {}},
			expect:  false,
		},
		{
			name:    "no_rules_should_log",
			path:    "/any",
			include: nil,
			exclude: nil,
			expect:  false,
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			got := middlewares.ShouldLog(c.path, c.include, c.exclude)
			require.Equal(t, c.expect, got)
		})
	}
}

func TestGetIPAddress(t *testing.T) {
	testcases := []struct {
		name   string
		hdr    map[string]string
		remote string
		expect string
	}{
		{
			name:   "real_ip",
			hdr:    map[string]string{"X-Real-IP": "1.1.1.1"},
			remote: "",
			expect: "1.1.1.1",
		},
		{
			name:   "forwarded",
			hdr:    map[string]string{"X-Forwarded-For": "2.2.2.2, 10.0.0.1"},
			remote: "",
			expect: "2.2.2.2",
		},
		{
			name:   "remote_addr",
			hdr:    nil,
			remote: "3.3.3.3:1234",
			expect: "3.3.3.3",
		},
		{
			name:   "invalid",
			hdr:    map[string]string{"X-Real-IP": "bad"},
			remote: "bad:123",
			expect: "",
		},
	}

	for _, c := range testcases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if c.remote != "" {
				r.RemoteAddr = c.remote
			}
			for k, v := range c.hdr {
				r.Header.Set(k, v)
			}
			got := middlewares.GetIPAddress(r)
			require.Equal(t, c.expect, got)
		})
	}
}

func TestIsValidIP(t *testing.T) {
	tests := []struct {
		ip      string
		isValid bool
	}{
		{"192.168.1.1", true},
		{"256.256.256.256", false},
		{"not-a-ip", false},
		{"::1", true},
		{"::", true},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			result := middlewares.IsValidIP(tt.ip)
			if result != tt.isValid {
				t.Errorf("Expected %v for IP %v, got %v", tt.isValid, tt.ip, result)
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// Prepare logger -> Write to buffer
	buf := &bufWriter{}
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(buf)

	// Create middleware: include /v1*, exclude /v1/skip*
	mw := middlewares.LoggingMiddleware(logger,
		map[string]struct{}{"/v1": {}},
		map[string]struct{}{"/v1/healthz": {}})

	// helper creates a handler with the desired status
	makeHandler := func(status int) http.Handler {
		return mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))
	}

	testCases := []struct {
		name         string
		path         string
		status       int
		shouldLog    bool
		expectStatus string
		expectCode   int
	}{
		{
			name:         "log_201",
			path:         "/v1/resource",
			status:       http.StatusCreated,
			shouldLog:    true,
			expectStatus: "success",
			expectCode:   http.StatusCreated,
		},
		{
			name:         "log_500",
			path:         "/v1/resource",
			status:       http.StatusInternalServerError,
			shouldLog:    true,
			expectStatus: "error",
			expectCode:   http.StatusInternalServerError,
		},
		{
			name:         "skip_path",
			path:         "/v1/healthz",
			status:       http.StatusOK,
			shouldLog:    false,
			expectStatus: "",
			expectCode:   http.StatusOK,
		},
		{
			name:         "no_include",
			path:         "/random/path",
			status:       http.StatusOK,
			shouldLog:    false,
			expectStatus: "",
			expectCode:   http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf.Reset()

			req := httptest.NewRequest("GET", tc.path, nil)
			rec := httptest.NewRecorder()
			makeHandler(tc.status).ServeHTTP(rec, req)

			require.Equal(t, tc.status, rec.Code)

			if !tc.shouldLog {
				require.Zero(t, buf.Len(), "expected no log for skipped path")
				return
			}

			// log → parse JSON
			var entry map[string]any
			require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))

			require.Equal(t, float64(tc.status), entry["code"])
			require.Equal(t, tc.expectStatus, entry["status"])
			require.Equal(t, "GET", entry["method"])
			require.Equal(t, tc.path, entry["path"])

			// latency_ms
			lat, ok := entry["latency_ms"].(float64)
			require.True(t, ok)
			require.GreaterOrEqual(t, lat, float64(0))
		})
	}
}
