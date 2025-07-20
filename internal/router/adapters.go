package router

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

// Adapter helpers for router handler registration
//
// Use these to convert custom handler signatures to http.HandlerFunc for chi.
// - Adapt: for standard handlers (w, r)
// - WithUser: for handlers needing a user (w, r, user)
// - WithOptionalUser: for handlers with optional user (w, r, *user)
// - WithAdmin: for admin-only handlers (w, r, user)
//
// This ensures all routes are registered as http.HandlerFunc and middleware is applied consistently.

// Adapt adapts a standard handler (w, r) to http.HandlerFunc for chi router compatibility.
func Adapt(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return http.HandlerFunc(h)
}

// WithUser adapts a handler (w, r, user) to http.HandlerFunc, expects user in context.
func WithUser(h func(http.ResponseWriter, *http.Request, database.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(contextKeyUser).(database.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		h(w, r, user)
	}
}

// WithOptionalUser adapts a handler (w, r, *user) to http.HandlerFunc, user may be nil.
func WithOptionalUser(h func(http.ResponseWriter, *http.Request, *database.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user *database.User
		if u, ok := r.Context().Value(contextKeyUser).(database.User); ok {
			user = &u
		}
		h(w, r, user)
	}
}

// WithAdmin adapts a handler (w, r, user) to http.HandlerFunc, expects user in context and checks admin role.
func WithAdmin(h func(http.ResponseWriter, *http.Request, database.User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(contextKeyUser).(database.User)
		if !ok || user.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h(w, r, user)
	}
}

// contextKeyUser is the context key for user, should match your handlers/user package
type contextKey string

const contextKeyUser contextKey = "user"
