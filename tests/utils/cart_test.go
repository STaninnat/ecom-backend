package utils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/STaninnat/ecom-backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetSessionIDFromRequest(t *testing.T) {
	t.Run("should return session ID when cookie exists", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{
			Name:  utils.GuestCartSessionCookie,
			Value: "abc123",
		})

		sessionID := utils.GetSessionIDFromRequest(req)
		assert.Equal(t, "abc123", sessionID)
	})

	t.Run("should return empty string when cookie not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		sessionID := utils.GetSessionIDFromRequest(req)
		assert.Equal(t, "", sessionID)
	})
}
