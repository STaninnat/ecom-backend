// Package middlewares provides HTTP middleware components for request processing in the ecom-backend project.
package middlewares

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"context"
)

// cache_middleware.go: Provides middleware for HTTP response caching and cache invalidation.

// CacheServiceIface defines the interface for cache operations used by middleware.
// This allows for easier testing and mocking.
// (You can use testify/mock or mockery to generate mocks)
//
//go:generate mockery --name=CacheServiceIface --output=../utils/mocks --outpkg=mocks
type CacheServiceIface interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	DeletePattern(ctx context.Context, pattern string) error
}

// CacheConfig holds configuration for caching, including TTL, key prefix, and the cache service implementation.
type CacheConfig struct {
	TTL          time.Duration
	KeyPrefix    string
	CacheService CacheServiceIface
}

// CacheMiddleware creates a middleware that caches HTTP responses for GET requests.
// It checks for cached responses before processing and caches successful responses for future requests.
func CacheMiddleware(config CacheConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Generate cache key
			cacheKey := generateCacheKey(config.KeyPrefix, r)

			// Try to get from cache
			var cachedResponse CachedResponse
			found, err := config.CacheService.Get(r.Context(), cacheKey, &cachedResponse)
			if err != nil {
				// Log error but continue without cache
				next.ServeHTTP(w, r)
				return
			}

			if found {
				// Return cached response
				for key, values := range cachedResponse.Headers {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				w.WriteHeader(cachedResponse.StatusCode)
				_, _ = w.Write(cachedResponse.Body)
				return
			}

			// Cache miss - capture the response
			responseWriter := &responseCapture{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				headers:        make(http.Header),
				body:           &bytes.Buffer{},
			}

			next.ServeHTTP(responseWriter, r)

			// Cache the response if it was successful
			if responseWriter.statusCode == http.StatusOK {
				cachedResponse := CachedResponse{
					StatusCode: responseWriter.statusCode,
					Headers:    responseWriter.headers,
					Body:       responseWriter.body.Bytes(),
				}

				err := config.CacheService.Set(r.Context(), cacheKey, cachedResponse, config.TTL)
				_ = err
			}
		})
	}
}

// CachedResponse represents a cached HTTP response, including status code, headers, and body.
type CachedResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
}

// responseCapture captures the response for caching
type responseCapture struct {
	http.ResponseWriter
	statusCode int
	headers    http.Header
	body       *bytes.Buffer
}

// WriteHeader captures the status code and propagates headers to the real ResponseWriter
// It ensures all headers are properly set before writing the status code
func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	// Propagate all headers set on rc.headers to the real ResponseWriter
	for k, vv := range rc.headers {
		for _, v := range vv {
			rc.ResponseWriter.Header().Add(k, v)
		}
	}
	rc.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body and writes it to both the buffer and real ResponseWriter
// It also ensures headers are propagated if WriteHeader hasn't been called yet
func (rc *responseCapture) Write(data []byte) (int, error) {
	// Copy headers from rc.headers to the real ResponseWriter if not already done
	for k, vv := range rc.headers {
		for _, v := range vv {
			rc.ResponseWriter.Header().Add(k, v)
		}
	}
	rc.body.Write(data)
	return rc.ResponseWriter.Write(data)
}

// Header returns the captured headers for the response
func (rc *responseCapture) Header() http.Header {
	return rc.headers
}

// generateCacheKey creates a unique cache key based on the request
// It combines method, path, query parameters, and user ID (if available) into a hash
// The resulting key is prefixed with the configured key prefix for organization
// SHA-256 hashing ensures consistent key length regardless of input size
func generateCacheKey(prefix string, r *http.Request) string {
	// Create a unique identifier for this request
	keyData := fmt.Sprintf("%s:%s:%s", r.Method, r.URL.Path, r.URL.RawQuery)

	// Add user-specific data if available (e.g., user ID from context)
	if userID := r.Context().Value("userID"); userID != nil {
		keyData += fmt.Sprintf(":user:%v", userID)
	}

	// Create SHA-256 hash for consistent key length
	hash := sha256.Sum256([]byte(keyData))
	return fmt.Sprintf("%s:%x", prefix, hash)
}

// InvalidateCache removes cached entries matching a pattern after the handler executes.
// Useful for cache invalidation after data modifications.
func InvalidateCache(cacheService CacheServiceIface, pattern string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Execute the handler first
			next.ServeHTTP(w, r)

			// Then invalidate cache
			err := cacheService.DeletePattern(r.Context(), pattern)
			_ = err
		})
	}
}
