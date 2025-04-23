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
		include []string
		exclude []string
		expect  bool
	}{
		{"include_match", "/api/v1", []string{"/api"}, nil, true},
		{"exclude_match", "/health", []string{"/"}, []string{"/health"}, false},
		{"no_rule", "/random", nil, nil, true},
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
		{"real_ip", map[string]string{"X-Real-IP": "1.1.1.1"}, "", "1.1.1.1"},
		{"forwarded", map[string]string{"X-Forwarded-For": "2.2.2.2, 10.0.0.1"}, "", "2.2.2.2"},
		{"remote_addr", nil, "3.3.3.3:1234", "3.3.3.3"},
		{"invalid", map[string]string{"X-Real-IP": "bad"}, "bad:123", ""},
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

func TestLoggingMiddleware(t *testing.T) {
	// Prepare logger -> Write to buffer
	buf := &bufWriter{}
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(buf)

	// Create middleware: include /v1*, exclude /v1/skip*
	mw := middlewares.LoggingMiddleware(logger,
		[]string{"/v1"},
		[]string{"/healthz"},
	)

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
	}{
		{"log_201", "/v1/resource", http.StatusCreated, true, "success"},
		{"log_500", "/v1/resource", http.StatusInternalServerError, true, "error"},
		{"skip_path", "/healthz", http.StatusOK, false, ""},
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
