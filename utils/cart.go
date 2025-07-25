// Package utils provides utility functions and helpers used throughout the ecom-backend project.
package utils

import "net/http"

// cart.go: Provides helpers for managing guest cart sessions and retrieving session IDs from HTTP requests.

// GuestCartSessionCookie is the name of the cookie used to store the guest session ID.
const GuestCartSessionCookie = "guest_session_id"

// GetSessionIDFromRequest retrieves the guest session ID from the request cookie.
// Returns an empty string if the cookie is not found.
func GetSessionIDFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(GuestCartSessionCookie)
	if err != nil {
		// Uncomment the next line to enable debug logging for missing cookie
		// log.Printf("Guest session cookie not found: %v", err)
		return ""
	}
	return cookie.Value
}
