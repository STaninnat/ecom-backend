// Package router defines HTTP routing, adapters, and related logic for the ecom-backend project.
package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/internal/database"
	"github.com/stretchr/testify/assert"
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

// TestWithAdmin_Admin tests the WithAdmin adapter when user is an admin.
// It should call the handler function with the admin user from context.
func TestWithAdmin_Admin(t *testing.T) {
	called := false
	h := func(_ http.ResponseWriter, _ *http.Request, user database.User) {
		called = true
		assert.Equal(t, "admin", user.Role)
	}
	hf := WithAdmin(h)
	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	user := database.User{Role: "admin"}
	ctx := context.WithValue(req.Context(), contextKeyUser, user)
	hf.ServeHTTP(rw, req.WithContext(ctx))
	assert.True(t, called)
}
