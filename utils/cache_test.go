package utils

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisCmdable is a testify mock for redis.Cmdable, used to simulate Redis operations in CacheService tests.
type MockRedisCmdable struct {
	mock.Mock
	redis.Cmdable
}

// Get mocks the Redis GET command for testing.
func (m *MockRedisCmdable) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.StringCmd)
}

// Set mocks the Redis SET command for testing.
func (m *MockRedisCmdable) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	return args.Get(0).(*redis.StatusCmd)
}

// Del mocks the Redis DEL command for testing.
func (m *MockRedisCmdable) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

// Keys mocks the Redis KEYS command for testing.
func (m *MockRedisCmdable) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	args := m.Called(ctx, pattern)
	return args.Get(0).(*redis.StringSliceCmd)
}

// Exists mocks the Redis EXISTS command for testing.
func (m *MockRedisCmdable) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	return args.Get(0).(*redis.IntCmd)
}

// TTL mocks the Redis TTL command for testing.
func (m *MockRedisCmdable) TTL(ctx context.Context, key string) *redis.DurationCmd {
	args := m.Called(ctx, key)
	return args.Get(0).(*redis.DurationCmd)
}

// FlushAll mocks the Redis FLUSHALL command for testing.
func (m *MockRedisCmdable) FlushAll(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	return args.Get(0).(*redis.StatusCmd)
}

// TestCacheService_Get tests the Get method of CacheService for:
// - Key does not exist
// - Redis error
// - Unmarshal error
// - Success
func TestCacheService_Get(t *testing.T) {
	t.Run("key does not exist", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "missing-key"
		// Simulate Redis returning redis.Nil
		cmd := redis.NewStringResult("", redis.Nil)
		mockRedis.On("Get", ctx, key).Return(cmd)

		var dest string
		found, err := cache.Get(ctx, key, &dest)
		assert.NoError(t, err)
		assert.False(t, found)
		mockRedis.AssertExpectations(t)
	})
	t.Run("redis error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "error-key"
		errRedis := redis.NewStringResult("", assert.AnError)
		mockRedis.On("Get", ctx, key).Return(errRedis)

		var dest string
		found, err := cache.Get(ctx, key, &dest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache get error")
		assert.False(t, found)
		mockRedis.AssertExpectations(t)
	})
	t.Run("unmarshal error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "bad-json"
		// Simulate Redis returning a value that is not valid JSON for the dest type
		cmd := redis.NewStringResult("not-json", nil)
		mockRedis.On("Get", ctx, key).Return(cmd)

		var dest int // int can't unmarshal from "not-json"
		found, err := cache.Get(ctx, key, &dest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache unmarshal error")
		assert.False(t, found)
		mockRedis.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "good-key"
		value := "hello"
		jsonVal := `"hello"`
		cmd := redis.NewStringResult(jsonVal, nil)
		mockRedis.On("Get", ctx, key).Return(cmd)

		var dest string
		found, err := cache.Get(ctx, key, &dest)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, value, dest)
		mockRedis.AssertExpectations(t)
	})
}

// TestCacheService_Set tests the Set method of CacheService for:
// - Marshal error
// - Redis error
// - Success
func TestCacheService_Set(t *testing.T) {
	t.Run("marshal error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "bad-value"
		ch := make(chan int) // channels can't be marshaled to JSON
		err := cache.Set(ctx, key, ch, time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache marshal error")
	})

	t.Run("redis error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "err-key"
		val := "data"
		jsonVal, _ := json.Marshal(val)
		cmd := redis.NewStatusResult("", assert.AnError)
		mockRedis.On("Set", ctx, key, jsonVal, time.Minute).Return(cmd)

		err := cache.Set(ctx, key, val, time.Minute)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache set error")
		mockRedis.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "ok-key"
		val := "data"
		jsonVal, _ := json.Marshal(val)
		cmd := redis.NewStatusResult("OK", nil)
		mockRedis.On("Set", ctx, key, jsonVal, time.Minute).Return(cmd)

		err := cache.Set(ctx, key, val, time.Minute)
		assert.NoError(t, err)
		mockRedis.AssertExpectations(t)
	})
}

// TestCacheService_Delete tests the Delete method of CacheService for:
// - Redis error
// - Success
func TestCacheService_Delete(t *testing.T) {
	t.Run("redis error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "err-key"
		cmd := redis.NewIntResult(0, assert.AnError)
		mockRedis.On("Del", ctx, []string{key}).Return(cmd)

		err := cache.Delete(ctx, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache delete error")
		mockRedis.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "ok-key"
		cmd := redis.NewIntResult(1, nil)
		mockRedis.On("Del", ctx, []string{key}).Return(cmd)

		err := cache.Delete(ctx, key)
		assert.NoError(t, err)
		mockRedis.AssertExpectations(t)
	})
}

// TestCacheService_DeletePattern tests the DeletePattern method of CacheService for:
// - Keys error
// - Del error
// - No keys match
// - Success
func TestCacheService_DeletePattern(t *testing.T) {
	t.Run("keys error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		pattern := "bad*"
		cmd := redis.NewStringSliceResult(nil, assert.AnError)
		mockRedis.On("Keys", ctx, pattern).Return(cmd)

		err := cache.DeletePattern(ctx, pattern)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache keys pattern error")
		mockRedis.AssertExpectations(t)
	})

	t.Run("del error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		pattern := "err*"
		keys := []string{"k1", "k2"}
		cmdKeys := redis.NewStringSliceResult(keys, nil)
		cmdDel := redis.NewIntResult(0, assert.AnError)
		mockRedis.On("Keys", ctx, pattern).Return(cmdKeys)
		mockRedis.On("Del", ctx, keys).Return(cmdDel)

		err := cache.DeletePattern(ctx, pattern)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache delete pattern error")
		mockRedis.AssertExpectations(t)
	})

	t.Run("no keys match", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		pattern := "none*"
		cmd := redis.NewStringSliceResult([]string{}, nil)
		mockRedis.On("Keys", ctx, pattern).Return(cmd)

		err := cache.DeletePattern(ctx, pattern)
		assert.NoError(t, err)
		mockRedis.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		pattern := "ok*"
		keys := []string{"k1", "k2"}
		cmdKeys := redis.NewStringSliceResult(keys, nil)
		cmdDel := redis.NewIntResult(2, nil)
		mockRedis.On("Keys", ctx, pattern).Return(cmdKeys)
		mockRedis.On("Del", ctx, keys).Return(cmdDel)

		err := cache.DeletePattern(ctx, pattern)
		assert.NoError(t, err)
		mockRedis.AssertExpectations(t)
	})
}

// TestCacheService_Exists tests the Exists method of CacheService for:
// - Redis error
// - Exists
// - Does not exist
func TestCacheService_Exists(t *testing.T) {
	t.Run("redis error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "err-key"
		cmd := redis.NewIntResult(0, assert.AnError)
		mockRedis.On("Exists", ctx, []string{key}).Return(cmd)

		found, err := cache.Exists(ctx, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache exists error")
		assert.False(t, found)
		mockRedis.AssertExpectations(t)
	})

	t.Run("exists", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "ok-key"
		cmd := redis.NewIntResult(1, nil)
		mockRedis.On("Exists", ctx, []string{key}).Return(cmd)

		found, err := cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.True(t, found)
		mockRedis.AssertExpectations(t)
	})

	t.Run("does not exist", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "none-key"
		cmd := redis.NewIntResult(0, nil)
		mockRedis.On("Exists", ctx, []string{key}).Return(cmd)

		found, err := cache.Exists(ctx, key)
		assert.NoError(t, err)
		assert.False(t, found)
		mockRedis.AssertExpectations(t)
	})
}

// TestCacheService_TTL tests the TTL method of CacheService for:
// - Redis error
// - Success
func TestCacheService_TTL(t *testing.T) {
	t.Run("redis error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "err-key"
		cmd := redis.NewDurationResult(0, assert.AnError)
		mockRedis.On("TTL", ctx, key).Return(cmd)

		_, err := cache.TTL(ctx, key)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache TTL error")
		mockRedis.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		key := "ok-key"
		exp := 42 * time.Second
		cmd := redis.NewDurationResult(exp, nil)
		mockRedis.On("TTL", ctx, key).Return(cmd)

		ttl, err := cache.TTL(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, exp, ttl)
		mockRedis.AssertExpectations(t)
	})
}

// TestCacheService_FlushAll tests the FlushAll method of CacheService for:
// - Redis error
// - Success
func TestCacheService_FlushAll(t *testing.T) {
	t.Run("redis error", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		cmd := redis.NewStatusResult("", assert.AnError)
		mockRedis.On("FlushAll", ctx).Return(cmd)

		err := cache.FlushAll(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cache flush all error")
		mockRedis.AssertExpectations(t)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		mockRedis := new(MockRedisCmdable)
		cache := NewCacheService(mockRedis)

		cmd := redis.NewStatusResult("OK", nil)
		mockRedis.On("FlushAll", ctx).Return(cmd)

		err := cache.FlushAll(ctx)
		assert.NoError(t, err)
		mockRedis.AssertExpectations(t)
	})
}
