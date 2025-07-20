package config

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// InitRedis initializes a Redis connection using the legacy pattern.
// Uses log.Fatal on errors and is provided for backward compatibility.
// Prefer InitRedisWithError for better error handling.
func InitRedis() *redis.Client {
	ctx := context.Background()

	redisAddr := getEnvOrDefault("REDIS_ADDR", "")
	redisUsername := getEnvOrDefault("REDIS_USERNAME", "")
	redisPassword := getEnvOrDefault("REDIS_PASSWORD", "")

	provider := NewRedisProvider(redisAddr, redisUsername, redisPassword)
	redisClient, err := provider.Connect(ctx)
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	log.Println("Connected to Redis successfully...")
	return redisClient.(*redis.Client)
}

// InitRedisWithError initializes a Redis connection and returns the client and error.
// Returns errors instead of calling log.Fatal, making it suitable for testing and graceful error handling.
func InitRedisWithError(ctx context.Context) (redis.Cmdable, error) {
	redisAddr := getEnvOrDefault("REDIS_ADDR", "")
	redisUsername := getEnvOrDefault("REDIS_USERNAME", "")
	redisPassword := getEnvOrDefault("REDIS_PASSWORD", "")

	provider := NewRedisProvider(redisAddr, redisUsername, redisPassword)
	return provider.Connect(ctx)
}
