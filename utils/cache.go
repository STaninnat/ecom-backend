// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// cache.go: Implements Redis-based caching service with get, set, delete, and pattern operations.

// CacheService provides Redis-based caching functionality.
type CacheService struct {
	client redis.Cmdable
}

// NewCacheService creates a new CacheService instance using the provided Redis client.
func NewCacheService(client redis.Cmdable) *CacheService {
	return &CacheService{
		client: client,
	}
}

// Get retrieves a value from cache by key and unmarshals it into the provided destination.
// Returns true if the key exists, false otherwise.
func (c *CacheService) Get(ctx context.Context, key string, dest any) (bool, error) {
	val, err := c.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil // Key doesn't exist
	}
	if err != nil {
		return false, fmt.Errorf("cache get error: %w", err)
	}

	// Unmarshal the JSON value into the destination
	err = json.Unmarshal([]byte(val), dest)
	if err != nil {
		return false, fmt.Errorf("cache unmarshal error: %w", err)
	}

	return true, nil
}

// Set stores a value in cache under the given key, marshaling it as JSON, with the specified TTL.
func (c *CacheService) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	// Marshal the value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal error: %w", err)
	}

	// Store in Redis with TTL
	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}

	return nil
}

// Delete removes a key from cache.
func (c *CacheService) Delete(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("cache delete error: %w", err)
	}
	return nil
}

// DeletePattern removes all keys matching a pattern (e.g., "products:*").
func (c *CacheService) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := c.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("cache keys pattern error: %w", err)
	}

	if len(keys) > 0 {
		err = c.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("cache delete pattern error: %w", err)
		}
	}

	return nil
}

// Exists checks if a key exists in cache.
func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("cache exists error: %w", err)
	}
	return result > 0, nil
}

// TTL gets the remaining time-to-live for a key.
func (c *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("cache TTL error: %w", err)
	}
	return ttl, nil
}

// FlushAll clears all cache entries from Redis (use with caution!).
func (c *CacheService) FlushAll(ctx context.Context) error {
	err := c.client.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("cache flush all error: %w", err)
	}
	return nil
}
