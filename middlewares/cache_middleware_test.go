package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockCacheService struct {
	mock.Mock
}

// Get mocks the cache Get method for testing purposes
func (m *MockCacheService) Get(ctx context.Context, key string, dest any) (bool, error) {
	args := m.Called(ctx, key, dest)
	if args.Get(2) != nil {
		reflect.ValueOf(dest).Elem().Set(reflect.ValueOf(args.Get(2)))
	}
	return args.Bool(0), args.Error(1)
}
func (m *MockCacheService) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}
func (m *MockCacheService) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

// TestCacheMiddleware_CacheHit tests the cache middleware when a cached response is found
// It verifies that cached responses are returned immediately without calling the handler
func TestCacheMiddleware_CacheHit(t *testing.T) {
	cached := CachedResponse{
		StatusCode: 200,
		Headers:    map[string][]string{"X-Test": {"1"}},
		Body:       []byte("cached body"),
	}
	mockCache := new(MockCacheService)
	mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(true, nil, cached)

	config := CacheConfig{TTL: time.Minute, KeyPrefix: "test", CacheService: mockCache}
	mw := CacheMiddleware(config)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called on cache hit")
	}))
	r := httptest.NewRequest("GET", "/foo", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Body.String() != "cached body" {
		t.Errorf("expected cached body, got %q", rw.Body.String())
	}
	if rw.Header().Get("X-Test") != "1" {
		t.Errorf("expected header X-Test=1, got %q", rw.Header().Get("X-Test"))
	}
	mockCache.AssertExpectations(t)
}

// TestCacheMiddleware_CacheMiss tests the cache middleware when no cached response exists
// It verifies that the handler is called and the response is cached for future requests
func TestCacheMiddleware_CacheMiss(t *testing.T) {
	mockCache := new(MockCacheService)
	mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(false, nil, nil)
	mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	config := CacheConfig{TTL: time.Minute, KeyPrefix: "test", CacheService: mockCache}
	mw := CacheMiddleware(config)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Fresh", "yes")
		w.Write([]byte("fresh body"))
	}))
	r := httptest.NewRequest("GET", "/bar", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Body.String() != "fresh body" {
		t.Errorf("expected fresh body, got %q", rw.Body.String())
	}
	if rw.Header().Get("X-Fresh") != "yes" {
		t.Errorf("expected header X-Fresh=yes, got %q", rw.Header().Get("X-Fresh"))
	}
	mockCache.AssertCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockCache.AssertExpectations(t)
}

// TestCacheMiddleware_NonGET tests the cache middleware with non-GET requests
// It verifies that caching is bypassed for non-GET requests and the handler is called normally
func TestCacheMiddleware_NonGET(t *testing.T) {
	mockCache := new(MockCacheService)
	config := CacheConfig{TTL: time.Minute, KeyPrefix: "test", CacheService: mockCache}
	mw := CacheMiddleware(config)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not cached"))
	}))
	r := httptest.NewRequest("POST", "/baz", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Body.String() != "not cached" {
		t.Errorf("expected handler body, got %q", rw.Body.String())
	}
	mockCache.AssertNotCalled(t, "Get", mock.Anything, mock.Anything, mock.Anything)
	mockCache.AssertNotCalled(t, "Set", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

// TestCacheMiddleware_GetError tests the cache middleware when cache retrieval fails
// It verifies that the middleware gracefully handles cache errors and falls back to the handler
func TestCacheMiddleware_GetError(t *testing.T) {
	mockCache := new(MockCacheService)
	mockCache.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(false, errors.New("fail"), nil)

	config := CacheConfig{TTL: time.Minute, KeyPrefix: "test", CacheService: mockCache}
	mw := CacheMiddleware(config)
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fallback"))
	}))
	r := httptest.NewRequest("GET", "/err", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Body.String() != "fallback" {
		t.Errorf("expected fallback body, got %q", rw.Body.String())
	}
	mockCache.AssertExpectations(t)
}

// TestInvalidateCache tests the cache invalidation middleware functionality
// It verifies that cache entries are deleted after the handler executes successfully
func TestInvalidateCache(t *testing.T) {
	mockCache := new(MockCacheService)
	mockCache.On("DeletePattern", mock.Anything, "test*").Return(nil)

	mw := InvalidateCache(mockCache, "test*")
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("done"))
	}))
	r := httptest.NewRequest("DELETE", "/invalidate", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, r)
	if rw.Body.String() != "done" {
		t.Errorf("expected handler body, got %q", rw.Body.String())
	}
	mockCache.AssertExpectations(t)
}
