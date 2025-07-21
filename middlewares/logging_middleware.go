// Package middlewares provides HTTP middleware components for request processing in the ecom-backend project.
package middlewares

import (
	"context"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/STaninnat/ecom-backend/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// logging_middleware.go: Middleware for structured request logging and request ID tracing.

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the status code for logging purposes
func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware creates a middleware that logs HTTP requests with detailed information.
// It measures request duration, captures status codes, and logs structured information.
// The middleware supports path filtering through include/exclude path maps.
// It categorizes responses as success (2xx), fail (4xx), or error (5xx) for monitoring.
// Request metadata includes method, path, status, duration, IP, user agent, and request ID.
func LoggingMiddleware(logger *logrus.Logger, includePaths, excludePaths map[string]struct{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			if !ShouldLog(path, includePaths, excludePaths) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			sw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sw, r)
			duration := time.Since(start)

			statusText := "success"
			if sw.status >= 400 && sw.status < 500 {
				statusText = "fail"
			} else if sw.status >= 500 {
				statusText = "error"
			}

			requestID := r.Context().Value(utils.ContextKeyRequestID)

			logger.WithFields(logrus.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"status":     statusText,
				"code":       sw.status,
				"duration":   duration.String(),
				"latency_ms": duration.Milliseconds(),
				"ip":         GetIPAddress(r),
				"user_agent": r.UserAgent(),
				"referrer":   r.Referer(),
				"request_id": requestID,
			}).Info("HTTP request")
		})
	}
}

// RequestIDMiddleware adds a unique request ID to each request context for tracing and correlation.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		ctx := context.WithValue(r.Context(), utils.ContextKeyRequestID, requestID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// ShouldLog determines whether a request path should be logged based on include/exclude rules.
func ShouldLog(path string, includePaths, excludePaths map[string]struct{}) bool {
	for ex := range excludePaths {
		if strings.HasPrefix(path, ex) {
			return false
		}
	}

	for in := range includePaths {
		if strings.HasPrefix(path, in) {
			return true
		}
	}

	return false
}

// GetIPAddress extracts the real client IP address from various request headers or RemoteAddr.
func GetIPAddress(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip != "" && IsValidIP(ip) {
		return ip
	}

	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		for _, candidate := range ips {
			ip = strings.TrimSpace(candidate)
			if IsValidIP(ip) {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && IsValidIP(host) {
		return host
	}

	return ""
}

// IsValidIP validates whether a string represents a valid IP address (IPv4 or IPv6).
func IsValidIP(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil
}
