package handlers

import (
	"net/http"

	"github.com/STaninnat/ecom-backend/internal/database"
)

type optionalHandler func(http.ResponseWriter, *http.Request, *database.User)

func (apicfg *HandlersConfig) HandlerOptionalMiddleware(handler optionalHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user *database.User
		ip, userAgent := GetRequestMetadata(r)

		cookie, err := r.Cookie("access_token")
		if err == nil {
			token := cookie.Value

			claims, err := apicfg.Auth.ValidateAccessToken(token, apicfg.JWTSecret)
			if err != nil {
				apicfg.LogHandlerError(r.Context(), "optional auth", "invalid token", "token validation failed", ip, userAgent, err)
			} else {
				u, err := apicfg.DB.GetUserByID(r.Context(), claims.UserID)
				if err != nil {
					apicfg.LogHandlerError(r.Context(), "optional auth", "user not found", "user lookup failed", ip, userAgent, err)
				} else {
					user = &u
				}
			}
		}

		handler(w, r, user)
	}
}
