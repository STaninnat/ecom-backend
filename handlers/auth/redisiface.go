package authhandlers

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type MinimalRedis interface {
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}
