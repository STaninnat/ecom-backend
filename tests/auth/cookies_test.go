package auth_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/STaninnat/ecom-backend/auth"
)

func TestSetTokensAsCookies(t *testing.T) {
	// Create a new HTTP response recorder to capture the response
	w := httptest.NewRecorder()

	// Sample tokens and their expiry times
	accessToken := "access_token_value"
	refreshToken := "refresh_token_value"
	accessTokenExpiresAt := time.Now().UTC().Add(time.Hour)
	refreshTokenExpiresAt := time.Now().UTC().Add(24 * time.Hour)

	auth.SetTokensAsCookies(w, accessToken, refreshToken, accessTokenExpiresAt, refreshTokenExpiresAt)

	resp := w.Result()
	cookies := resp.Cookies()

	if len(cookies) != 2 {
		t.Errorf("Expected 2 cookies, got %d", len(cookies))
	}

	// Loop through the cookies to check their values and properties
	for _, cookie := range cookies {
		// Check the access_token cookie
		switch cookie.Name {
		case "access_token":
			// Validate the access_token value
			if cookie.Value != accessToken {
				t.Errorf("Expected access_token value %v, got %v", accessToken, cookie.Value)
			}
			// Validate the expiration time for access_token
			if !cookie.Expires.Truncate(time.Second).Equal(accessTokenExpiresAt.Truncate(time.Second)) {
				t.Errorf("Expected access_token expiry %v, got %v", accessTokenExpiresAt, cookie.Expires)
			}
		// Check the refresh_token cookie
		case "refresh_token":
			// Validate the refresh_token value
			if cookie.Value != refreshToken {
				t.Errorf("Expected refresh_token value %v, got %v", refreshToken, cookie.Value)
			}
			// Validate the expiration time for refresh_token
			if !cookie.Expires.Truncate(time.Second).Equal(refreshTokenExpiresAt.Truncate(time.Second)) {
				t.Errorf("Expected refresh_token expiry %v, got %v", refreshTokenExpiresAt, cookie.Expires)
			}
		default:
			// If an unexpected cookie is found, report an error
			t.Errorf("Unexpected cookie name: %v", cookie.Name)
		}

		// Test Secure flag: the cookie should be set as Secure
		if !cookie.Secure {
			t.Errorf("Expected cookie %v to be Secure", cookie.Name)
		}

		// Test HttpOnly flag: the cookie should be set as HttpOnly
		if !cookie.HttpOnly {
			t.Errorf("Expected cookie %v to be HttpOnly", cookie.Name)
		}

		// Test Path attribute: the cookie path should be '/'
		if cookie.Path != "/" {
			t.Errorf("Expected cookie %v path to be '/', got %v", cookie.Name, cookie.Path)
		}
	}
}
