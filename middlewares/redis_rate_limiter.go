package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRateLimiter creates a distributed rate limiter using Redis
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
				http.Error(w, "Rate limit error", http.StatusInternalServerError)
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
