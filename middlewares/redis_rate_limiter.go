// Package middlewares provides HTTP middleware components for request processing in the ecom-backend project.
package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// redis_rate_limiter.go: Distributed rate limiting middleware using Redis for request throttling.

// RedisRateLimiter creates a distributed rate limiter middleware using Redis.
// Tracks requests per client IP, sets rate limit headers, and returns HTTP 429 if the limit is exceeded.
// Uses Redis pipeline for atomic operations and supports custom limits and windows.
func RedisRateLimiter(redisClient redis.Cmdable, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			key := "rate_limit:" + getClientIP(r)

			// Get context for the request
			ctx := r.Context()

			// Use Redis pipeline for atomic operations
			pipe := redisClient.TxPipeline()

			// Increment the counter
			incr := pipe.Incr(ctx, key)

			// Set expiration if key doesn't exist
			pipe.Expire(ctx, key, window)

			// Execute pipeline
			_, err := pipe.Exec(ctx)
			if err != nil {
				http.Error(w, `{"error":"Internal server error","code":"INTERNAL_ERROR","message":"An unexpected error occurred. Please try again later."}`, http.StatusInternalServerError)
				return
			}

			// Get current count
			currentCount := incr.Val()

			// Get TTL for reset time
			ttl, err := redisClient.TTL(ctx, key).Result()
			if err != nil {
				ttl = window
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(int64(limit), 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(int64(limit)-currentCount, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(ttl).Unix(), 10))

			// Check if limit exceeded
			if currentCount > int64(limit) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the real client IP from request headers
// It checks X-Forwarded-For and X-Real-IP headers first, then falls back to RemoteAddr
// This ensures proper IP detection when behind proxies or load balancers
func getClientIP(r *http.Request) string {
	// Check for forwarded headers first
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fallback to remote address
	return r.RemoteAddr
}
