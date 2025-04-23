package middlewares

import (
	"net"
	"net/http"
	"strings"
	"time"

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

func LoggingMiddleware(logger *logrus.Logger, includePaths []string, excludePaths []string) func(http.Handler) http.Handler {
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
			}).Info("HTTP request")
		})
	}
}
func ShouldLog(path string, includePaths, excludePaths []string) bool {
	for _, ex := range excludePaths {
		if strings.HasPrefix(path, ex) {
			return false
		}
	}

	for _, in := range includePaths {
		if strings.HasPrefix(path, in) {
			return true
		}
	}

	return true
}

func GetIPAddress(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip != "" && isValidIP(ip) {
		return ip
	}

	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		for _, candidate := range ips {
			ip = strings.TrimSpace(candidate)
			if isValidIP(ip) {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && isValidIP(host) {
		return host
	}

	return ""
}

func isValidIP(ip string) bool {
	parsed := net.ParseIP(ip)
	return parsed != nil
}
