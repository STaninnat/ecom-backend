package utils

import "net/http"

const GuestCartSessionCookie = "guest_session_id"

// GetSessionIDFromRequest retrieves the guest session ID from the request cookie.
func GetSessionIDFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(GuestCartSessionCookie)
	if err != nil {
		// Uncomment the next line to enable debug logging for missing cookie
		// log.Printf("Guest session cookie not found: %v", err)
		return ""
	}
	return cookie.Value
}
