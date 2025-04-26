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

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

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

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewString()
		ctx := context.WithValue(r.Context(), utils.ContextKeyRequestID, requestID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

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

func IsValidIP(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil
}
