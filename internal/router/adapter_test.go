// Package router defines HTTP routing, adapters, and related logic for the ecom-backend project.
package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// adapter_test.go: Tests for handler adapter utilities and context handling.

// TestAdapt verifies that the Adapt function correctly wraps an http.HandlerFunc
// and calls the underlying handler when ServeHTTP is invoked.
func TestAdapt(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	}
	hf := Adapt(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	hf.ServeHTTP(rw, req)
	assert.True(t, called)
}

// TestWithUser_NoUser tests the WithUser adapter when no user is present in context.
// It should return 401 Unauthorized and not call the handler function.
func TestWithUser_NoUser(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
		called = true
	}
	hf := WithUser(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	hf.ServeHTTP(rw, req)
	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rw.Code)
}

// TestWithUser_WithUser tests the WithUser adapter when a user is present in context.
// It should call the handler function with the user from context.
func TestWithUser_WithUser(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, user database.User) {
		called = true
		assert.Equal(t, "user", user.Role)
	}
	hf := WithUser(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	user := database.User{Role: "user"}
	ctx := context.WithValue(req.Context(), contextKeyUser, user)
	hf.ServeHTTP(rw, req.WithContext(ctx))
	assert.True(t, called)
}

// TestWithOptionalUser_NilUser tests the WithOptionalUser adapter when no user is present.
// It should call the handler function with a nil user pointer.
func TestWithOptionalUser_NilUser(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, user *database.User) {
		called = true
		assert.Nil(t, user)
	}
	hf := WithOptionalUser(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	hf.ServeHTTP(rw, req)
	assert.True(t, called)
}

// TestWithOptionalUser_WithUser tests the WithOptionalUser adapter when a user is present.
// It should call the handler function with a pointer to the user from context.
func TestWithOptionalUser_WithUser(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, user *database.User) {
		called = true
		assert.NotNil(t, user)
		assert.Equal(t, "user", user.Role)
	}
	hf := WithOptionalUser(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	user := database.User{Role: "user"}
	ctx := context.WithValue(req.Context(), contextKeyUser, user)
	hf.ServeHTTP(rw, req.WithContext(ctx))
	assert.True(t, called)
}

// TestWithAdmin_NotAdmin tests the WithAdmin adapter when user is not an admin.
// It should return 403 Forbidden and not call the handler function.
func TestWithAdmin_NotAdmin(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
		called = true
	}
	hf := WithAdmin(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	user := database.User{Role: "user"}
	ctx := context.WithValue(req.Context(), contextKeyUser, user)
	hf.ServeHTTP(rw, req.WithContext(ctx))
	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rw.Code)
}

// TestWithAdmin_NoUserInContext tests the WithAdmin adapter when no user is in context.
func TestWithAdmin_NoUserInContext(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
		called = true
	}
	hf := WithAdmin(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	hf.ServeHTTP(rw, req)
	assert.False(t, called)
	assert.Equal(t, http.StatusForbidden, rw.Code)
}

// TestWithOptionalUser_InvalidUserData tests WithOptionalUser with invalid user data in context.
func TestWithOptionalUser_InvalidUserData(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, user *database.User) {
		called = true
		assert.Nil(t, user)
	}
	hf := WithOptionalUser(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	// Put invalid user data in context
	ctx := context.WithValue(req.Context(), contextKeyUser, "invalid user data")
	hf.ServeHTTP(rw, req.WithContext(ctx))
	assert.True(t, called)
}

// TestWithUser_InvalidUserData tests WithUser with invalid user data in context.
func TestWithUser_InvalidUserData(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, _ database.User) {
		called = true
	}
	hf := WithUser(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	// Put invalid user data in context
	ctx := context.WithValue(req.Context(), contextKeyUser, "invalid user data")
	hf.ServeHTTP(rw, req.WithContext(ctx))
	assert.False(t, called)
	assert.Equal(t, http.StatusUnauthorized, rw.Code)
}

// runConcurrentRequests is a helper function to test concurrent requests for both WithUser and WithAdmin adapters
func runConcurrentRequests(t *testing.T, adapter func(func(http.ResponseWriter, *http.Request, database.User)) http.HandlerFunc, expectedRole string) {
	h := func(_ http.ResponseWriter, _ *http.Request, user database.User) {
		assert.Equal(t, expectedRole, user.Role)
	}
	hf := adapter(h)

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Run multiple requests concurrently
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rw := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			user := database.User{Role: expectedRole}
			ctx := context.WithValue(req.Context(), contextKeyUser, user)
			hf.ServeHTTP(rw, req.WithContext(ctx))
			assert.Equal(t, http.StatusOK, rw.Code)
		}()
	}

	// Wait with timeout
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Test completed successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out")
	}
}

// TestWithAdmin_ConcurrentRequests tests WithAdmin adapter with concurrent requests.
func TestWithAdmin_ConcurrentRequests(t *testing.T) {
	runConcurrentRequests(t, WithAdmin, "admin")
}

// TestWithUser_ConcurrentRequests tests WithUser adapter with concurrent requests.
func TestWithUser_ConcurrentRequests(t *testing.T) {
	runConcurrentRequests(t, WithUser, "user")
}

func TestWithUserAndAdminAdapters(t *testing.T) {
	tests := []struct {
		name    string
		adapter func(func(http.ResponseWriter, *http.Request, database.User)) http.HandlerFunc
		role    string
	}{
		{
			name:    "WithUser adapter with user role",
			adapter: WithUser,
			role:    "user",
		},
		{
			name:    "WithAdmin adapter with admin role",
			adapter: WithAdmin,
			role:    "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			h := func(_ http.ResponseWriter, _ *http.Request, user database.User) {
				called = true
				assert.Equal(t, tt.role, user.Role)
			}
			hf := tt.adapter(h)
			rw := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			user := database.User{Role: tt.role}
			ctx := context.WithValue(req.Context(), contextKeyUser, user)
			hf.ServeHTTP(rw, req.WithContext(ctx))
			assert.True(t, called)
		})
	}
}

func TestAdapters_TableDriven(t *testing.T) {
	t.Run("Adapt calls handler", func(t *testing.T) {
		called := false
		h := Adapt(func(w http.ResponseWriter, r *http.Request) {
			called = true
		})
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h(w, req)
		assert.True(t, called)
	})

	t.Run("WithUser: user present", func(t *testing.T) {
		called := false
		h := WithUser(func(w http.ResponseWriter, r *http.Request, user database.User) {
			called = true
			assert.NotNil(t, user)
		})
		req := httptest.NewRequest("GET", "/", nil)
		ctx := context.WithValue(req.Context(), contextKeyUser, database.User{ID: "1"})
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		h(w, req)
		assert.True(t, called)
	})

	t.Run("WithUser: user missing", func(t *testing.T) {
		called := false
		h := WithUser(func(w http.ResponseWriter, r *http.Request, user database.User) {
			called = true
		})
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h(w, req)
		assert.False(t, called)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("WithAdmin: admin present", func(t *testing.T) {
		called := false
		admin := database.User{ID: "1", Role: "admin"}
		h := WithAdmin(func(w http.ResponseWriter, r *http.Request, user database.User) {
			called = true
			assert.Equal(t, "admin", user.Role)
		})
		req := httptest.NewRequest("GET", "/", nil)
		ctx := context.WithValue(req.Context(), contextKeyUser, admin)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		h(w, req)
		assert.True(t, called)
	})

	t.Run("WithAdmin: not admin", func(t *testing.T) {
		called := false
		user := database.User{ID: "1", Role: "user"}
		h := WithAdmin(func(w http.ResponseWriter, r *http.Request, user database.User) {
			called = true
		})
		req := httptest.NewRequest("GET", "/", nil)
		ctx := context.WithValue(req.Context(), contextKeyUser, user)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		h(w, req)
		assert.False(t, called)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("WithOptionalUser: user present", func(t *testing.T) {
		called := false
		user := database.User{ID: "1"}
		h := WithOptionalUser(func(w http.ResponseWriter, r *http.Request, user *database.User) {
			called = true
			assert.NotNil(t, user)
		})
		req := httptest.NewRequest("GET", "/", nil)
		ctx := context.WithValue(req.Context(), contextKeyUser, user)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		h(w, req)
		assert.True(t, called)
	})

	t.Run("WithOptionalUser: user missing", func(t *testing.T) {
		called := false
		h := WithOptionalUser(func(w http.ResponseWriter, r *http.Request, user *database.User) {
			called = true
			assert.Nil(t, user)
		})
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		h(w, req)
		assert.True(t, called)
	})
}
