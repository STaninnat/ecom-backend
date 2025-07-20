package auth

import (
	"net/http"
	"time"
)

// NOTE: This file only sets cookies. No performance improvements needed.

// SetTokensAsCookies sets the access and refresh tokens as secure, HTTP-only cookies in the response writer.
func SetTokensAsCookies(w http.ResponseWriter, accessToken, refreshToken string, accessTokenExpiresAt, refreshTokenExpiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Expires:  accessTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  refreshTokenExpiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})
}
