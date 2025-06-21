package utils

import "net/http"

const GuestCartSessionCookie = "guest_session_id"

func GetSessionIDFromRequest(r *http.Request) string {
	cookie, err := r.Cookie(GuestCartSessionCookie)
	if err != nil {
		return ""
	}

	return cookie.Value
}
